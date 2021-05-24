package restore

import (
	"bufio"
	"compress/gzip"
	"errors"
	"github.com/clevyr/kubedb/internal/kubernetes"
	"github.com/clevyr/kubedb/internal/postgres"
	"github.com/spf13/cobra"
	"io"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"log"
	"os"
	"strings"
)

const (
	GzipContentType = iota
	CustomContentType
	PlainContentType
)

var Command = &cobra.Command{
	Use:     "restore",
	Aliases: []string{"r"},
	Short:   "Restore a database",
	Args:    cobra.ExactArgs(1),
	PreRunE: preRun,
	RunE:    run,
}

var (
	dbname            string
	username          string
	password          string
	inputFormat       uint8
	singleTransaction bool
	clean             bool
	noOwner           bool
)

func init() {
	Command.Flags().StringVarP(&dbname, "dbname", "d", "db", "database name to connect to")
	Command.Flags().StringVarP(&username, "username", "U", "postgres", "database username")
	Command.Flags().StringVarP(&password, "password", "p", "", "database password")

	Command.Flags().StringP("format", "F", "", "input format. inferred by default ([g]zip, [c]ustom, [p]lain text)")

	Command.Flags().BoolVarP(&singleTransaction, "single-transaction", "1", true, "restore as a single transaction")
	Command.Flags().BoolVarP(&clean, "clean", "c", true, "clean (drop) database objects before recreating")
	Command.Flags().BoolVarP(&noOwner, "no-owner", "O", true, "skip restoration of object ownership in plain-text format")
}

func preRun(cmd *cobra.Command, args []string) error {
	format, _ := cmd.Flags().GetString("format")
	switch format {
	case "gzip", "gz", "g":
		inputFormat = GzipContentType
	case "plain", "sql", "p":
		inputFormat = PlainContentType
	case "custom", "c":
		inputFormat = CustomContentType
	default:
		lower := strings.ToLower(args[0])
		switch {
		case strings.HasSuffix(lower, ".sql.gz"):
			inputFormat = GzipContentType
		case strings.HasSuffix(lower, ".dmp"):
			inputFormat = CustomContentType
		case strings.HasSuffix(lower, ".sql"):
			inputFormat = PlainContentType
		default:
			return errors.New("invalid input file type")
		}
	}
	return nil
}

func run(cmd *cobra.Command, args []string) (err error) {
	cmd.SilenceUsage = true

	client, err := kubernetes.CreateClientForCmd(cmd)
	if err != nil {
		return err
	}

	postgresPod, err := postgres.GetPod(client)
	if err != nil {
		return err
	}

	if password == "" {
		password, err = postgres.GetSecret(client)
		if err != nil {
			return err
		}
	}

	infile, err := os.Open(args[0])
	if err != nil {
		return err
	}
	defer infile.Close()
	fileReader := bufio.NewReader(infile)

	pr, pw := io.Pipe()

	log.Println("Restoring \"" + args[0] + "\" to \"" + postgresPod.Name + "\"")

	ch := make(chan error)
	go func() {
		if clean {
			resetReader := strings.NewReader("drop schema public cascade; create schema public;")
			err := kubernetes.Exec(client, postgresPod, buildCommand(PlainContentType, false), resetReader, os.Stdout, false)
			if err != nil {
				pw.Close()
				ch <- err
				return
			}
		}

		err := kubernetes.Exec(client, postgresPod, buildCommand(inputFormat, true), pr, os.Stdout, false)
		pw.Close()
		ch <- err
	}()

	switch inputFormat {
	case GzipContentType, CustomContentType:
		_, err = io.Copy(pw, fileReader)
	case PlainContentType:
		gzw := gzip.NewWriter(pw)
		_, err = io.Copy(gzw, fileReader)
		gzw.Close()
	}
	if err != nil {
		return err
	}
	pw.Close()

	err = <-ch
	if err == nil {
		log.Println("Finished")
	}
	return err
}

func buildCommand(inputFormat uint8, gunzip bool) []string {
	cmd := []string{"PGPASSWORD=" + password}
	switch inputFormat {
	case GzipContentType, PlainContentType:
		if gunzip {
			cmd = append([]string{"gunzip", "--force", "|"}, cmd...)
		}
		cmd = append(cmd, "psql")
	case CustomContentType:
		cmd = append(cmd, "pg_restore", "--format=custom", "--verbose")
		if noOwner {
			cmd = append(cmd, "--no-owner")
		}
	}
	cmd = append(cmd, "--username="+username, "--dbname="+dbname)
	if singleTransaction {
		cmd = append(cmd, "--single-transaction")
	}
	return []string{"sh", "-c", strings.Join(cmd, " ")}
}

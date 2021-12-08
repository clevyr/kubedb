package restore

import (
	"bufio"
	"compress/gzip"
	"github.com/clevyr/kubedb/internal/cli"
	"github.com/clevyr/kubedb/internal/database/sqlformat"
	"github.com/clevyr/kubedb/internal/kubernetes"
	"github.com/clevyr/kubedb/internal/postgres"
	"github.com/spf13/cobra"
	"io"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"log"
	"os"
	"strings"
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
	inputFormat       sqlformat.Format
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
	formatStr, err := cmd.Flags().GetString("format")
	if err != nil {
		return err
	}
	inputFormat, err = sqlformat.ParseFormat(formatStr)
	if err != nil {
		inputFormat, err = sqlformat.ParseFilename(args[0])
		if err != nil {
			return err
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

	if err = cli.Confirm(os.Stdin, false); err != nil {
		return err
	}

	ch := make(chan error)
	go func() {
		if clean {
			resetReader := strings.NewReader("drop schema public cascade; create schema public;")
			err := kubernetes.Exec(client, postgresPod, buildCommand(sqlformat.Plain, false), resetReader, os.Stdout, false)
			if err != nil {
				pw.Close()
				ch <- err
				return
			}
		}

		err := kubernetes.Exec(client, postgresPod, buildCommand(inputFormat, true), pr, os.Stdout, false)
		pw.Close()
		if err != nil {
			ch <- err
			return
		}

		err = kubernetes.Exec(client, postgresPod, buildCommand(sqlformat.Plain, false), strings.NewReader("analyze"), os.Stdout, false)
		if err != nil {
			ch <- err
			return
		}

		ch <- nil
	}()

	switch inputFormat {
	case sqlformat.Gzip, sqlformat.Custom:
		_, err = io.Copy(pw, fileReader)
	case sqlformat.Plain:
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

func buildCommand(inputFormat sqlformat.Format, gunzip bool) []string {
	cmd := []string{"PGPASSWORD=" + password}
	switch inputFormat {
	case sqlformat.Gzip, sqlformat.Plain:
		if gunzip {
			cmd = append([]string{"gunzip", "--force", "|"}, cmd...)
		}
		cmd = append(cmd, "psql")
	case sqlformat.Custom:
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

package restore

import (
	"bufio"
	"compress/gzip"
	"github.com/clevyr/kubedb/internal/kubernetes"
	"github.com/clevyr/kubedb/internal/postgres"
	"github.com/spf13/cobra"
	"io"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"log"
	"net/http"
	"os"
	"strings"
)

var Command = &cobra.Command{
	Use:     "restore",
	Aliases: []string{"r"},
	Short:   "Restore a database",
	RunE:    run,
}

var (
	dbname            string
	username          string
	password          string
	singleTransaction bool
)

func init() {
	Command.Flags().StringVarP(&dbname, "dbname", "d", "db", "database name to connect to")
	Command.Flags().StringVarP(&username, "username", "U", "postgres", "database username")
	Command.Flags().StringVarP(&password, "password", "p", "", "database password")

	Command.Flags().BoolVar(&singleTransaction, "single-transaction", true, "execute as a single transaction")
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
		resetReader := strings.NewReader("drop schema public cascade; create schema public;")
		err := kubernetes.Exec(client, postgresPod, buildCommand(false), resetReader, os.Stdout, false)
		if err != nil {
			ch <- err
			return
		}

		err = kubernetes.Exec(client, postgresPod, buildCommand(true), pr, os.Stdout, false)
		ch <- err
	}()

	contentType, err := getFileContentType(infile)
	if err != nil {
		return err
	}

	switch contentType {
	case "application/x-gzip":
		_, err = io.Copy(pw, fileReader)
	default:
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

func buildCommand(gunzip bool) []string {
	var cmd []string
	if gunzip {
		cmd = []string{"gunzip", "|"}
	}
	cmd = append(cmd, "PGPASSWORD="+password, "psql", "--username="+username, "--dbname="+dbname)
	if singleTransaction {
		cmd = append(cmd, "--single-transaction")
	}
	return []string{"sh", "-c", strings.Join(cmd, " ")}
}

func getFileContentType(infile *os.File) (string, error) {
	// Only the first 512 bytes are used to sniff the content type.
	buffer := make([]byte, 512)

	_, err := infile.Read(buffer)
	if err != nil {
		return "", err
	}

	// Use the net/http package's handy DectectContentType function. Always returns a valid
	// content-type by returning "application/octet-stream" if no others seemed to match.
	contentType := http.DetectContentType(buffer)

	_, err = infile.Seek(0, 0)
	return contentType, err
}

package dump

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"github.com/clevyr/kubedb/internal/database/sqlformat"
	"github.com/clevyr/kubedb/internal/kubernetes"
	"github.com/clevyr/kubedb/internal/postgres"
	"github.com/spf13/cobra"
	"io"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"log"
	"os"
	"path"
	"strings"
	"text/template"
	"time"
)

var Command = &cobra.Command{
	Use:     "dump",
	Aliases: []string{"d"},
	Short:   "Dump a database",
	PreRunE: preRun,
	RunE:    run,
}

var (
	dbname           string
	username         string
	password         string
	directory        string
	outputFormat     uint8
	ifExists         bool
	clean            bool
	noOwner          bool
	excludeTable     []string
	excludeTableData []string
)

func init() {
	Command.Flags().StringVarP(&dbname, "dbname", "d", "db", "database name to connect to")
	Command.Flags().StringVarP(&username, "username", "U", "postgres", "database username")
	Command.Flags().StringVarP(&password, "password", "p", "", "database password")
	Command.Flags().StringVarP(&directory, "directory", "C", ".", "directory to dump to")

	Command.Flags().StringP("format", "F", "g", "output file format ([g]zip, [c]ustom, [p]lain text)")

	Command.Flags().BoolVar(&ifExists, "if-exists", true, "use IF EXISTS when dropping objects")
	Command.Flags().BoolVarP(&clean, "clean", "c", true, "clean (drop) database objects before recreating")
	Command.Flags().BoolVarP(&noOwner, "no-owner", "O", true, "skip restoration of object ownership in plain-text format")
	Command.Flags().StringSliceVarP(&excludeTable, "exclude-table", "T", []string{}, "do NOT dump the specified table(s)")
	Command.Flags().StringSliceVar(&excludeTableData, "exclude-table-data", []string{}, "do NOT dump data for the specified table(s)")
}

func preRun(cmd *cobra.Command, args []string) (err error) {
	formatStr, err := cmd.Flags().GetString("format")
	if err != nil {
		return err
	}
	outputFormat, err = sqlformat.ParseFormat(formatStr)
	if err != nil {
		return err
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

	filename, err := generateFilename(directory, client.Namespace)
	if err != nil {
		return err
	}

	if _, err := os.Stat(path.Dir(filename)); os.IsNotExist(err) {
		err = os.Mkdir(path.Dir(filename), os.ModePerm)
		if err != nil {
			return err
		}
	}

	outfile, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer outfile.Close()
	fileWriter := bufio.NewWriter(outfile)
	defer fileWriter.Flush()

	pr, pw := io.Pipe()

	log.Println("Dumping \"" + postgresPod.Name + "\" to \"" + filename + "\"")

	ch := make(chan error)
	go func() {
		err := kubernetes.Exec(client, postgresPod, buildCommand(), os.Stdin, pw, false)
		pw.Close()
		ch <- err
	}()

	switch outputFormat {
	case sqlformat.Gzip, sqlformat.Custom:
		_, err = io.Copy(fileWriter, pr)
	case sqlformat.Plain:
		var gzr *gzip.Reader
		gzr, err = gzip.NewReader(pr)
		if err != nil {
			return err
		}
		defer gzr.Close()
		_, err = io.Copy(fileWriter, gzr)
	}
	if err != nil {
		return err
	}

	err = <-ch
	if err == nil {
		log.Println("Finished")
	}
	return err
}

func generateFilename(directory, namespace string) (string, error) {
	directory = path.Clean(directory)
	t, err := template.New("filename").Parse("{{.directory}}/{{.namespace}}-{{.now.Format \"2006-01-02-150405\"}}")
	if err != nil {
		return "", err
	}

	var tpl bytes.Buffer
	data := map[string]interface{}{
		"directory": directory,
		"namespace": namespace,
		"now":       time.Now(),
	}
	err = t.Execute(&tpl, data)

	switch outputFormat {
	case sqlformat.Gzip:
		tpl.WriteString(".sql.gz")
	case sqlformat.Plain:
		tpl.WriteString(".sql")
	case sqlformat.Custom:
		tpl.WriteString(".dmp")
	}

	return tpl.String(), err
}

func buildCommand() []string {
	cmd := []string{"PGPASSWORD=" + password, "pg_dump", "--username=" + username, "--dbname=" + dbname}
	if clean {
		cmd = append(cmd, "--clean")
	}
	if noOwner {
		cmd = append(cmd, "--no-owner")
	}
	if ifExists {
		cmd = append(cmd, "--if-exists")
	}
	for _, table := range excludeTable {
		cmd = append(cmd, "--exclude-table=" + table)
	}
	for _, table := range excludeTableData {
		cmd = append(cmd, "--exclude-table-data=" + table)
	}
	if outputFormat == sqlformat.Custom {
		cmd = append(cmd, "--format=c")
	} else {
		cmd = append(cmd, "|", "gzip", "--force")
	}
	return []string{"sh", "-c", strings.Join(cmd, " ")}
}

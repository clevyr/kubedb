package dump

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/clevyr/kubedb/internal/config"
	"github.com/clevyr/kubedb/internal/database"
	"github.com/clevyr/kubedb/internal/database/sqlformat"
	"github.com/clevyr/kubedb/internal/gzips"
	"github.com/clevyr/kubedb/internal/kubernetes"
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

var conf config.Dump

func init() {
	Command.Flags().StringVarP(&conf.Database, "dbname", "d", "", "database name to connect to")
	Command.Flags().StringVarP(&conf.Username, "username", "U", "", "database username")
	Command.Flags().StringVarP(&conf.Password, "password", "p", "", "database password")
	Command.Flags().StringVarP(&conf.Directory, "directory", "C", ".", "directory to dump to")

	Command.Flags().StringP("format", "F", "g", "output file format ([g]zip, [c]ustom, [p]lain text)")

	Command.Flags().BoolVar(&conf.IfExists, "if-exists", true, "use IF EXISTS when dropping objects")
	Command.Flags().BoolVarP(&conf.Clean, "clean", "c", true, "clean (drop) database objects before recreating")
	Command.Flags().BoolVarP(&conf.NoOwner, "no-owner", "O", true, "skip restoration of object ownership in plain-text format")
	Command.Flags().StringSliceVarP(&conf.Tables, "table", "t", []string{}, "dump the specified table(s) only")
	Command.Flags().StringSliceVarP(&conf.ExcludeTable, "exclude-table", "T", []string{}, "do NOT dump the specified table(s)")
	Command.Flags().StringSliceVar(&conf.ExcludeTableData, "exclude-table-data", []string{}, "do NOT dump data for the specified table(s)")
}

func preRun(cmd *cobra.Command, args []string) (err error) {
	formatStr, err := cmd.Flags().GetString("format")
	if err != nil {
		return err
	}
	conf.OutputFormat, err = sqlformat.ParseFormat(formatStr)
	if err != nil {
		return err
	}
	return nil
}

func run(cmd *cobra.Command, args []string) (err error) {
	cmd.SilenceUsage = true

	dbName, err := cmd.Flags().GetString("type")
	if err != nil {
		return err
	}
	db, err := database.New(dbName)

	if conf.Database == "" {
		conf.Database = db.DefaultDatabase()
	}
	if conf.Username == "" {
		conf.Username = db.DefaultUser()
	}

	client, err := kubernetes.CreateClientForCmd(cmd)
	if err != nil {
		return err
	}

	pod, err := db.GetPod(client)
	if err != nil {
		return err
	}

	if conf.Password == "" {
		conf.Password, err = db.GetSecret(client)
		if err != nil {
			return err
		}
	}

	filename, err := generateFilename(conf.Directory, client.Namespace, conf.OutputFormat)
	if err != nil {
		return err
	}

	if _, err := os.Stat(path.Dir(filename)); os.IsNotExist(err) {
		err = os.Mkdir(path.Dir(filename), os.ModePerm)
		if err != nil {
			return err
		}
	}

	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	log.Println("Dumping \"" + pod.Name + "\" to \"" + filename + "\"")
	if githubActions, _ := cmd.Flags().GetBool("github-actions"); githubActions {
		fmt.Println("::set-output name=filename::" + filename)
	}

	ch := make(chan error, 1)
	var w io.WriteCloser
	switch conf.OutputFormat {
	case sqlformat.Gzip, sqlformat.Custom:
		w = f
		close(ch)
	case sqlformat.Plain:
		w = gzips.NewDecompressWriter(f, ch)
		if err != nil {
			return err
		}
	}
	defer func(w io.WriteCloser) {
		_ = w.Close()
	}(w)

	err = kubernetes.Exec(client, pod, buildCommand(db, conf), os.Stdin, w, false)
	if err != nil {
		return err
	}

	// Close writer
	err = w.Close()
	if err != nil {
		return err
	}

	// Check errors in channel
	err = <-ch
	if err != nil {
		return err
	}

	// Close file
	err = f.Close()
	if err != nil {
		// Ignore file already closed errors since w can be the same as f
		if !errors.Is(err, os.ErrClosed) {
			return err
		}
	}

	log.Println("Finished")
	return nil
}

func generateFilename(directory, namespace string, outputFormat sqlformat.Format) (string, error) {
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
	if err != nil {
		return "", err
	}

	ext, err := sqlformat.WriteExtension(outputFormat)
	if err != nil {
		return "", err
	}
	tpl.WriteString(ext)

	return tpl.String(), err
}

func buildCommand(db database.Databaser, conf config.Dump) []string {
	cmd := db.DumpCommand(conf)
	if conf.OutputFormat != sqlformat.Custom {
		cmd = append(cmd, "|", "gzip", "--force")
	}
	return []string{"sh", "-c", strings.Join(cmd, " ")}
}

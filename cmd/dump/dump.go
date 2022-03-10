package dump

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/clevyr/kubedb/internal/config"
	"github.com/clevyr/kubedb/internal/database/sqlformat"
	"github.com/clevyr/kubedb/internal/gzips"
	"github.com/clevyr/kubedb/internal/progressbar"
	"github.com/clevyr/kubedb/internal/util"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"io"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
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
	util.DefaultFlags(Command, &conf.Global)

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
		panic(err)
	}

	conf.OutputFormat, err = sqlformat.ParseFormat(formatStr)
	if err != nil {
		return err
	}

	return util.DefaultSetup(cmd, &conf.Global)
}

func run(cmd *cobra.Command, args []string) (err error) {
	filename, err := generateFilename(conf.Directory, conf.Client.Namespace, conf.OutputFormat)
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

	log.WithFields(log.Fields{
		"pod":  conf.Pod.Name,
		"file": filename,
	}).Info("exporting database")

	if githubActions, err := cmd.Flags().GetBool("github-actions"); err != nil {
		panic(err)
	} else if githubActions {
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

	bar := progressbar.New()

	err = conf.Client.Exec(conf.Pod, buildCommand(conf.Grammar, conf), os.Stdin, io.MultiWriter(w, bar), false)
	if err != nil {
		return err
	}

	_ = bar.Finish()

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

	log.Info("finished")
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

func buildCommand(db config.Databaser, conf config.Dump) []string {
	cmd := db.DumpCommand(conf)
	if conf.OutputFormat != sqlformat.Custom {
		cmd = append(cmd, "|", "gzip", "--force")
	}
	return []string{"sh", "-c", strings.Join(cmd, " ")}
}

package dump

import (
	"errors"
	"fmt"
	"github.com/clevyr/kubedb/internal/config"
	"github.com/clevyr/kubedb/internal/config/flags"
	"github.com/clevyr/kubedb/internal/database/sqlformat"
	"github.com/clevyr/kubedb/internal/gzips"
	"github.com/clevyr/kubedb/internal/progressbar"
	"github.com/clevyr/kubedb/internal/util"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"io"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"os"
	"path/filepath"
	"strings"
)

var Command = &cobra.Command{
	Use:     "dump [filename]",
	Aliases: []string{"d", "export"},
	Short:   "dump a database to a sql file",
	Long: `The "dump" command dumps a running database to a sql file.

If no filename is provided, the filename will be generated.
For example, if a dump is performed in the namespace "clevyr" with no extra flags,
the generated filename might look like "` + HelpFilename() + `"`,

	Args:              cobra.MaximumNArgs(1),
	ValidArgsFunction: validArgs,

	PreRunE: preRun,
	RunE:    run,
}

var conf config.Dump

func init() {
	flags.Directory(Command, &conf.Directory)
	flags.Format(Command)
	flags.IfExists(Command, &conf.IfExists)
	flags.Clean(Command, &conf.Clean)
	flags.NoOwner(Command, &conf.NoOwner)
	flags.Tables(Command, &conf.Tables)
	flags.ExcludeTable(Command, &conf.ExcludeTable)
	flags.ExcludeTableData(Command, &conf.ExcludeTableData)
}

func validArgs(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) != 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	return []string{"sql", "sql.gz", "dmp"}, cobra.ShellCompDirectiveFilterFileExt
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
	var filename string
	if len(args) > 0 {
		filename = args[0]
	} else {
		filename, err = Filename{
			Dir:       conf.Directory,
			Namespace: conf.Client.Namespace,
			Format:    conf.OutputFormat,
		}.Generate()
		if err != nil {
			return err
		}
	}

	if _, err := os.Stat(filepath.Dir(filename)); os.IsNotExist(err) {
		err = os.Mkdir(filepath.Dir(filename), os.ModePerm)
		if err != nil {
			return err
		}
	}

	var f io.WriteCloser
	if filename == "-" {
		f = os.Stdout
	} else {
		f, err = os.Create(filename)
		if err != nil {
			return err
		}
		defer func(f io.WriteCloser) {
			_ = f.Close()
		}(f)
	}

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

	bar := progressbar.New(-1)
	log.SetOutput(progressbar.NewBarSafeLogger(os.Stderr))

	err = conf.Client.Exec(conf.Pod, buildCommand(conf.Grammar, conf), os.Stdin, io.MultiWriter(w, bar), false)
	if err != nil {
		return err
	}

	_ = bar.Finish()
	log.SetOutput(os.Stderr)

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

	log.WithField("file", filename).Info("dump complete")
	return nil
}

func buildCommand(db config.Databaser, conf config.Dump) []string {
	cmd := db.DumpCommand(conf)
	if conf.OutputFormat != sqlformat.Custom {
		cmd = append(cmd, "|", "gzip", "--force")
	}
	return []string{"sh", "-c", strings.Join(cmd, " ")}
}

package restore

import (
	"compress/gzip"
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"github.com/clevyr/kubedb/internal/command"
	"github.com/clevyr/kubedb/internal/config"
	"github.com/clevyr/kubedb/internal/config/flags"
	"github.com/clevyr/kubedb/internal/database/sqlformat"
	"github.com/clevyr/kubedb/internal/progressbar"
	"github.com/clevyr/kubedb/internal/util"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"io"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"os"
	"strings"
)

var Command = &cobra.Command{
	Use:     "restore filename",
	Aliases: []string{"r", "import"},
	Short:   "restore a database from a sql file",
	Long: `The "restore" command restores a given sql file to a running database pod.

Supported Input Filetypes:
  - Raw sql file. Typically with the ".sql" extension
  - Gzipped sql file. Typically with the ".sql.gz" extension
  - Postgres custom dump file. Typically with the ".dmp" extension (Only if the target database is Postgres)`,

	Args:              cobra.ExactArgs(1),
	ValidArgsFunction: validArgs,

	PreRunE: preRun,
	RunE:    run,
}

var (
	conf config.Restore
)

func init() {
	flags.Format(Command, &conf.Format)
	flags.SingleTransaction(Command, &conf.SingleTransaction)
	flags.Clean(Command, &conf.Clean)
	flags.NoOwner(Command, &conf.NoOwner)
	flags.Force(Command, &conf.Force)
}

func validArgs(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) != 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	return []string{"sql", "sql.gz", "dmp", "archive", "archive.gz"}, cobra.ShellCompDirectiveFilterFileExt
}

func preRun(cmd *cobra.Command, args []string) (err error) {
	conf.Filename = args[0]

	if err := util.DefaultSetup(cmd, &conf.Global); err != nil {
		return err
	}

	if !cmd.Flags().Lookup("format").Changed {
		conf.Format = conf.Dialect.FormatFromFilename(conf.Filename)
	}

	return nil
}

func run(cmd *cobra.Command, args []string) (err error) {
	f, err := os.Open(conf.Filename)
	if err != nil {
		return err
	}
	defer f.Close()

	pr, pw := io.Pipe()

	log.WithFields(log.Fields{
		"file": conf.Filename,
		"pod":  conf.Pod.Name,
	}).Info("ready to restore database")

	if !conf.Force {
		var response bool
		err = survey.AskOne(&survey.Confirm{
			Message: "Restore to " + conf.Pod.Name + " in " + conf.Client.Namespace + "?",
		}, &response)
		if err != nil {
			return err
		}
		if !response {
			fmt.Println("restore canceled")
			return
		}
	}

	ch := make(chan error, 1)
	go runInDatabasePod(pr, ch, conf.Format)

	bar := progressbar.New(-1)
	log.SetOutput(progressbar.NewBarSafeLogger(os.Stderr))

	w := io.MultiWriter(pw, bar)

	if conf.Clean && conf.Format != sqlformat.Custom {
		dropQuery := conf.Dialect.DropDatabaseQuery(conf.Database)
		if dropQuery != "" {
			log.Info("cleaning existing data")
			err = gzipCopy(w, strings.NewReader(dropQuery))
			if err != nil {
				return err
			}
		}
	}

	log.Info("restoring database")
	switch conf.Format {
	case sqlformat.Gzip, sqlformat.Custom:
		_, err = io.Copy(w, f)
		if err != nil {
			return err
		}
	case sqlformat.Plain:
		err = gzipCopy(w, f)
		if err != nil {
			return err
		}
	}

	analyzeQuery := conf.Dialect.AnalyzeQuery()
	if analyzeQuery != "" {
		if conf.Format == sqlformat.Custom {
			_ = pw.Close()

			err = <-ch
			if err != nil {
				return err
			}

			pr, pw = io.Pipe()
			w = io.MultiWriter(pw, bar)
			go runInDatabasePod(pr, ch, sqlformat.Gzip)
		}

		log.Info("analyzing data")
		err = gzipCopy(w, strings.NewReader(analyzeQuery))
		if err != nil {
			return err
		}
	}

	_ = bar.Finish()
	log.SetOutput(os.Stderr)

	_ = pw.Close()

	err = <-ch
	if err != nil {
		return err
	}

	log.WithField("file", conf.Filename).Info("restore complete")
	return nil
}

func buildCommand(conf config.Restore, inputFormat sqlformat.Format) *command.Builder {
	return conf.Dialect.RestoreCommand(conf, inputFormat).
		Unshift("gunzip", "--force", command.Pipe)
}

func gzipCopy(w io.Writer, r io.Reader) (err error) {
	gzw := gzip.NewWriter(w)

	_, err = io.Copy(gzw, r)
	if err != nil {
		return err
	}

	err = gzw.Close()
	if err != nil {
		return err
	}

	return nil
}

func runInDatabasePod(r *io.PipeReader, ch chan error, inputFormat sqlformat.Format) {
	var err error
	defer func(pr *io.PipeReader) {
		_ = pr.Close()
	}(r)

	err = conf.Client.Exec(conf.Pod, buildCommand(conf, inputFormat).String(), r, os.Stdout, os.Stderr, false, nil)
	if err != nil {
		_ = r.CloseWithError(err)
		ch <- err
		return
	}

	ch <- nil
}

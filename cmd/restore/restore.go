package restore

import (
	"compress/gzip"
	"fmt"
	"github.com/AlecAivazis/survey/v2"
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
	conf        config.Restore
	inputFormat sqlformat.Format
)

func init() {
	flags.Format(Command)
	flags.SingleTransaction(Command, &conf.SingleTransaction)
	flags.Clean(Command, &conf.Clean)
	flags.NoOwner(Command, &conf.NoOwner)
	flags.Force(Command, &conf.Force)
}

func validArgs(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) != 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	return []string{"sql", "sql.gz", "dmp"}, cobra.ShellCompDirectiveFilterFileExt
}

func preRun(cmd *cobra.Command, args []string) error {
	formatStr, err := cmd.Flags().GetString("format")
	if err != nil {
		panic(err)
	}

	inputFormat, err = sqlformat.ParseFormat(formatStr)
	if err != nil {
		inputFormat, err = sqlformat.ParseFilename(args[0])
		if err != nil {
			return err
		}
	}

	return util.DefaultSetup(cmd, &conf.Global)
}

func run(cmd *cobra.Command, args []string) (err error) {
	f, err := os.Open(args[0])
	if err != nil {
		return err
	}
	defer f.Close()

	pr, pw := io.Pipe()

	log.WithFields(log.Fields{
		"file": args[0],
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
	go func() {
		var err error
		defer func(pw *io.PipeWriter) {
			_ = pw.Close()
		}(pw)

		if conf.Clean && conf.Grammar.DropDatabaseQuery(conf.Database) != "" {
			log.Info("cleaning existing data")
			resetReader := strings.NewReader(conf.Grammar.DropDatabaseQuery(conf.Database))
			err = conf.Client.Exec(conf.Pod, buildCommand(conf.Grammar, conf, sqlformat.Plain, false), resetReader, os.Stdout, false)
			if err != nil {
				ch <- err
				return
			}
		}

		log.Info("restoring database")
		err = conf.Client.Exec(conf.Pod, buildCommand(conf.Grammar, conf, inputFormat, true), pr, os.Stdout, false)
		if err != nil {
			ch <- err
			return
		}

		if conf.Grammar.AnalyzeQuery() != "" {
			log.Info("analyzing data")
			analyzeReader := strings.NewReader(conf.Grammar.AnalyzeQuery())
			err = conf.Client.Exec(conf.Pod, buildCommand(conf.Grammar, conf, sqlformat.Plain, false), analyzeReader, os.Stdout, false)
			if err != nil {
				ch <- err
				return
			}
		}

		ch <- nil
	}()

	stat, _ := f.Stat()
	bar := progressbar.New(stat.Size())
	log.SetOutput(progressbar.NewBarSafeLogger(os.Stderr))

	switch inputFormat {
	case sqlformat.Gzip, sqlformat.Custom:
		_, err = io.Copy(io.MultiWriter(pw, bar), f)
		if err != nil {
			return err
		}
	case sqlformat.Plain:
		gzw := gzip.NewWriter(pw)

		_, err = io.Copy(io.MultiWriter(pw, bar), f)
		if err != nil {
			return err
		}

		err = gzw.Close()
		if err != nil {
			return err
		}
	}
	_ = pw.Close()

	_ = bar.Finish()
	log.SetOutput(os.Stderr)

	err = <-ch
	if err != nil {
		return err
	}

	log.WithField("file", args[0]).Info("restore complete")
	return nil
}

func buildCommand(db config.Databaser, conf config.Restore, inputFormat sqlformat.Format, gunzip bool) []string {
	var cmd []string
	switch inputFormat {
	case sqlformat.Gzip, sqlformat.Plain:
		if gunzip {
			cmd = append([]string{"gunzip", "--force", "|"}, cmd...)
		}
	}
	cmd = append(cmd, db.RestoreCommand(conf, inputFormat)...)
	return []string{"sh", "-c", strings.Join(cmd, " ")}
}

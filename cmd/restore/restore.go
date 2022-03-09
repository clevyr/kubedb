package restore

import (
	"compress/gzip"
	"github.com/clevyr/kubedb/internal/cli"
	"github.com/clevyr/kubedb/internal/config"
	"github.com/clevyr/kubedb/internal/database/sqlformat"
	"github.com/clevyr/kubedb/internal/util"
	"github.com/schollz/progressbar/v3"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"io"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
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
	conf        config.Restore
	inputFormat sqlformat.Format
)

func init() {
	util.DefaultFlags(Command, conf.Global)

	Command.Flags().StringP("format", "F", "", "input format. inferred by default ([g]zip, [c]ustom, [p]lain text)")

	Command.Flags().BoolVarP(&conf.SingleTransaction, "single-transaction", "1", true, "restore as a single transaction")
	Command.Flags().BoolVarP(&conf.Clean, "clean", "c", true, "clean (drop) database objects before recreating")
	Command.Flags().BoolVarP(&conf.NoOwner, "no-owner", "O", true, "skip restoration of object ownership in plain-text format")

	Command.Flags().BoolVarP(&conf.Force, "force", "f", false, "do not prompt before restore")
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
		if err = cli.Confirm(os.Stdin, false); err != nil {
			return err
		}
	}

	ch := make(chan error, 1)
	go func() {
		var err error
		defer func(pw *io.PipeWriter) {
			_ = pw.Close()
		}(pw)

		if conf.Clean {
			log.Info("cleaning existing data")
			resetReader := strings.NewReader(conf.Databaser.DropDatabaseQuery(conf.Database))
			err = conf.Client.Exec(conf.Pod, buildCommand(conf.Databaser, conf, sqlformat.Plain, false), resetReader, os.Stdout, false)
			if err != nil {
				ch <- err
				return
			}
		}

		log.Info("restoring database")
		err = conf.Client.Exec(conf.Pod, buildCommand(conf.Databaser, conf, inputFormat, true), pr, os.Stdout, false)
		if err != nil {
			ch <- err
			return
		}

		analyzeReader := strings.NewReader(conf.Databaser.AnalyzeQuery())
		err = conf.Client.Exec(conf.Pod, buildCommand(conf.Databaser, conf, sqlformat.Plain, false), analyzeReader, os.Stdout, false)
		if err != nil {
			ch <- err
			return
		}

		ch <- nil
	}()

	bar := progressbar.DefaultBytes(-1)

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

	_ = bar.Finish()

	log.Info("finished")
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

package dump

import (
	"compress/gzip"
	"encoding/base64"
	"errors"
	"fmt"
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
	"k8s.io/client-go/tools/remotecommand"
	"k8s.io/kubectl/pkg/util/term"
	"os"
	"path/filepath"
	"strconv"
	"time"
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

	Annotations: map[string]string{
		"access": strconv.Itoa(config.ReadOnly),
	},
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
			Date:      time.Now(),
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

	bar := progressbar.New(-1)
	log.SetOutput(progressbar.NewBarSafeLogger(os.Stderr))

	pr, pw := io.Pipe()
	ch := make(chan error, 1)
	go func() {
		var err error
		defer func() {
			ch <- err
		}()

		// base64 is required since TTYs use CRLF
		pr := base64.NewDecoder(base64.StdEncoding, pr)

		switch conf.OutputFormat {
		case sqlformat.Gzip, sqlformat.Custom:
			_, err = io.Copy(io.MultiWriter(f, bar), pr)
		case sqlformat.Plain:
			var gzr *gzip.Reader
			gzr, err = gzip.NewReader(pr)
			if err != nil {
				return
			}

			_, err = io.Copy(io.MultiWriter(f, bar), gzr)
			if err != nil {
				return
			}
		}
	}()

	t := term.TTY{
		In:  os.Stdin,
		Out: pw,
	}
	t.Raw = t.IsTerminalIn()
	var sizeQueue remotecommand.TerminalSizeQueue
	if t.Raw {
		sizeQueue = t.MonitorSize(t.GetSize())
	}

	err = t.Safe(func() error {
		return conf.Client.Exec(
			conf.Pod,
			buildCommand(conf.Grammar, conf).String(),
			t.In,
			t.Out,
			os.Stderr,
			t.IsTerminalIn(),
			sizeQueue,
		)
	})
	if err != nil {
		_ = pw.CloseWithError(err)
		return err
	}

	err = pw.Close()
	if err != nil {
		return err
	}

	_ = bar.Finish()
	log.SetOutput(os.Stderr)

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

func buildCommand(db config.Databaser, conf config.Dump) *command.Builder {
	cmd := db.DumpCommand(conf)
	if conf.OutputFormat != sqlformat.Custom {
		cmd.Push(command.Raw("|"), "gzip", "--force")
	}
	// base64 is required since TTYs use CRLF
	cmd.Push(command.Raw("|"), "base64", "-w0")
	return cmd
}

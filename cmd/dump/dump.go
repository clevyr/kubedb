package dump

import (
	"compress/gzip"
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
	flags.Format(Command, &conf.Format)
	flags.IfExists(Command, &conf.IfExists)
	flags.Clean(Command, &conf.Clean)
	flags.NoOwner(Command, &conf.NoOwner)
	flags.Tables(Command, &conf.Tables)
	flags.ExcludeTable(Command, &conf.ExcludeTable)
	flags.ExcludeTableData(Command, &conf.ExcludeTableData)
	flags.Quiet(Command, &conf.Quiet)
}

func validArgs(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) != 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	err := preRun(cmd, args)
	if err != nil {
		return []string{"sql", "sql.gz", "dmp", "archive", "archive.gz"}, cobra.ShellCompDirectiveFilterFileExt
	}

	var exts []string
	for _, ext := range conf.Dialect.Formats() {
		exts = append(exts, ext[1:])
	}
	return exts, cobra.ShellCompDirectiveFilterFileExt
}

func preRun(cmd *cobra.Command, args []string) (err error) {
	if len(args) > 0 {
		conf.Filename = args[0]
	}

	if conf.Directory != "" {
		cmd.SilenceUsage = true
		if err = os.Chdir(conf.Directory); err != nil {
			return err
		}
		cmd.SilenceUsage = false
	}

	if err := util.DefaultSetup(cmd, &conf.Global); err != nil {
		return err
	}

	if conf.Filename != "" && !cmd.Flags().Lookup("format").Changed {
		conf.Format = conf.Dialect.FormatFromFilename(conf.Filename)
	}

	return nil
}

func run(cmd *cobra.Command, args []string) (err error) {
	if conf.Filename == "" {
		conf.Filename, err = Filename{
			Namespace: conf.Client.Namespace,
			Ext:       conf.Dialect.DumpExtension(conf.Format),
			Date:      time.Now(),
		}.Generate()
		if err != nil {
			return err
		}
	}

	var f io.WriteCloser
	switch conf.Filename {
	case "-":
		f = os.Stdout
	default:
		if _, err := os.Stat(filepath.Dir(conf.Filename)); os.IsNotExist(err) {
			err = os.MkdirAll(filepath.Dir(conf.Filename), os.ModePerm)
			if err != nil {
				return err
			}
		}

		f, err = os.Create(conf.Filename)
		if err != nil {
			return err
		}
		defer func(f io.WriteCloser) {
			_ = f.Close()
		}(f)
	}

	log.WithFields(log.Fields{
		"pod":       conf.Pod.Name,
		"namespace": conf.Client.Namespace,
		"file":      conf.Filename,
	}).Info("exporting database")

	if githubActions, err := cmd.Flags().GetBool("github-actions"); err != nil {
		panic(err)
	} else if githubActions {
		fmt.Println("::set-output name=filename::" + conf.Filename)
	}

	var startTime = time.Now()

	bar := progressbar.New(-1, "downloading")
	plogger := progressbar.NewBarSafeLogger(os.Stderr, bar)
	log.SetOutput(plogger)

	pr, pw := io.Pipe()
	ch := make(chan error, 1)
	go func() {
		var err error
		defer func() {
			ch <- err
		}()

		pr := io.Reader(pr)

		if conf.Format == sqlformat.Plain {
			pr, err = gzip.NewReader(pr)
			if err != nil {
				return
			}
		}

		_, err = io.Copy(io.MultiWriter(f, bar), pr)
		if err != nil {
			return
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

	err = conf.Client.Exec(
		conf.Pod,
		buildCommand(conf.Dialect, conf).String(),
		t.In,
		t.Out,
		plogger,
		false,
		sizeQueue,
	)
	if err != nil {
		_ = pw.CloseWithError(err)
		return err
	}

	err = pw.Close()
	if err != nil {
		return err
	}

	err = <-ch
	if err != nil {
		return err
	}

	_ = bar.Finish()
	log.SetOutput(os.Stderr)

	// Close file
	err = f.Close()
	if err != nil {
		// Ignore file already closed errors since w can be the same as f
		if !errors.Is(err, os.ErrClosed) {
			return err
		}
	}

	log.WithFields(log.Fields{
		"file": conf.Filename,
		"in":   time.Since(startTime).Truncate(10 * time.Millisecond),
	}).Info("dump complete")
	return nil
}

func buildCommand(db config.Databaser, conf config.Dump) *command.Builder {
	cmd := db.DumpCommand(conf)
	if conf.Format != sqlformat.Custom {
		cmd.Push(command.Pipe, "gzip", "--force")
	}
	log.WithField("cmd", cmd).Trace("finished building command")
	return cmd
}

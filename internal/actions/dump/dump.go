package dump

import (
	"errors"
	"fmt"
	"github.com/clevyr/kubedb/internal/command"
	"github.com/clevyr/kubedb/internal/config"
	"github.com/clevyr/kubedb/internal/database/sqlformat"
	"github.com/clevyr/kubedb/internal/progressbar"
	gzip "github.com/klauspost/pgzip"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"io"
	"k8s.io/client-go/tools/remotecommand"
	"k8s.io/kubectl/pkg/util/term"
	"os"
	"path/filepath"
	"time"
)

type Dump struct {
	config.Dump `mapstructure:",squash"`
}

func (action Dump) Run() (err error) {
	if action.Filename == "" {
		action.Filename, err = Filename{
			Namespace: action.Client.Namespace,
			Ext:       action.Dialect.DumpExtension(action.Format),
			Date:      time.Now(),
		}.Generate()
		if err != nil {
			return err
		}
	}

	var f io.WriteCloser
	switch action.Filename {
	case "-":
		f = os.Stdout
	default:
		if _, err := os.Stat(filepath.Dir(action.Filename)); os.IsNotExist(err) {
			err = os.MkdirAll(filepath.Dir(action.Filename), os.ModePerm)
			if err != nil {
				return err
			}
		}

		f, err = os.Create(action.Filename)
		if err != nil {
			return err
		}
		defer func(f io.WriteCloser) {
			_ = f.Close()
		}(f)
	}

	log.WithFields(log.Fields{
		"pod":       action.Pod.Name,
		"namespace": action.Client.Namespace,
		"file":      action.Filename,
	}).Info("exporting database")

	if viper.GetBool("github-actions") {
		fmt.Println("::set-output name=filename::" + action.Filename)
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

		if action.Format == sqlformat.Plain {
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

	err = action.Client.Exec(
		action.Pod,
		action.buildCommand().String(),
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
		"file": action.Filename,
		"in":   time.Since(startTime).Truncate(10 * time.Millisecond),
	}).Info("dump complete")

	return nil
}

func (action Dump) buildCommand() *command.Builder {
	cmd := action.Dialect.DumpCommand(action.Dump)
	if action.Format != sqlformat.Custom {
		cmd.Push(command.Pipe, "gzip", "--force")
	}
	log.WithField("cmd", cmd).Trace("finished building command")
	return cmd
}

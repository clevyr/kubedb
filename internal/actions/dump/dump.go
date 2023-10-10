package dump

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/clevyr/kubedb/internal/command"
	"github.com/clevyr/kubedb/internal/config"
	"github.com/clevyr/kubedb/internal/database/sqlformat"
	"github.com/clevyr/kubedb/internal/github"
	"github.com/clevyr/kubedb/internal/progressbar"
	gzip "github.com/klauspost/pgzip"
	log "github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

type Dump struct {
	config.Dump `mapstructure:",squash"`
}

func (action Dump) Run(ctx context.Context) (err error) {
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
		"namespace": action.Client.Namespace,
		"name":      "pod/" + action.DbPod.Name,
		"file":      action.Filename,
	}).Info("exporting database")

	if err := github.SetOutput("filename", action.Filename); err != nil {
		return err
	}

	startTime := time.Now()

	bar := progressbar.New(-1, "downloading", action.Spinner)
	plogger := progressbar.NewBarSafeLogger(os.Stderr, bar)
	log.SetOutput(plogger)

	errGroup, ctx := errgroup.WithContext(ctx)

	pr, pw := io.Pipe()
	// Begin database export
	errGroup.Go(func() error {
		defer func(pw io.WriteCloser) {
			_ = pw.Close()
		}(pw)

		if err := action.Client.Exec(
			ctx,
			action.JobPod,
			action.buildCommand().String(),
			os.Stdin,
			pw,
			plogger,
			false,
			nil,
			0,
		); err != nil {
			return err
		}

		return pw.Close()
	})

	if !action.RemoteGzip {
		gzPipeReader, gzPipeWriter := io.Pipe()
		plainReader := pr
		errGroup.Go(func() error {
			defer func() {
				_ = gzPipeWriter.Close()
			}()

			gzw := gzip.NewWriter(gzPipeWriter)
			if _, err := io.Copy(gzw, plainReader); err != nil {
				return err
			}

			return gzw.Close()
		})
		pr = gzPipeReader
	}

	// Begin copying export to local file
	errGroup.Go(func() error {
		defer func(pr io.ReadCloser) {
			_ = pr.Close()
		}(pr)

		r := io.Reader(pr)

		if action.RemoteGzip {
			if action.Format == sqlformat.Plain {
				r, err = gzip.NewReader(r)
				if err != nil {
					return err
				}
			}
		}

		_, err := io.Copy(io.MultiWriter(f, bar), r)
		return err
	})

	if err := errGroup.Wait(); err != nil {
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
	if action.RemoteGzip && action.Format != sqlformat.Custom {
		cmd.Push(command.Pipe, "gzip", "--force")
	}
	log.WithField("cmd", cmd).Trace("finished building command")
	return cmd
}

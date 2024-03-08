package dump

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/clevyr/kubedb/internal/command"
	"github.com/clevyr/kubedb/internal/config"
	"github.com/clevyr/kubedb/internal/consts"
	"github.com/clevyr/kubedb/internal/database/sqlformat"
	"github.com/clevyr/kubedb/internal/github"
	"github.com/clevyr/kubedb/internal/kubernetes"
	"github.com/clevyr/kubedb/internal/progressbar"
	"github.com/clevyr/kubedb/internal/storage"
	"github.com/clevyr/kubedb/internal/util"
	gzip "github.com/klauspost/pgzip"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"golang.org/x/sync/errgroup"
)

type Dump struct {
	config.Dump `mapstructure:",squash"`
}

func (action Dump) Run(ctx context.Context) (err error) {
	errGroup, ctx := errgroup.WithContext(ctx)

	var f io.WriteCloser
	switch {
	case action.Filename == "-":
		f = os.Stdout
	case storage.IsS3(action.Filename):
		pr, pw := io.Pipe()
		f = pw
		defer func(pw *io.PipeWriter) {
			_ = pw.Close()
		}(pw)

		errGroup.Go(func() error {
			return storage.CreateS3Upload(ctx, pr, action.Filename)
		})
	case storage.IsGCS(action.Filename):
		f, err = storage.CreateGCSUpload(ctx, action.Filename)
		if err != nil {
			return err
		}
		defer func(f io.WriteCloser) {
			_ = f.Close()
		}(f)
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

	bar, plogger := progressbar.New(os.Stderr, -1, "downloading", action.Spinner)
	defer bar.Close()

	pr, pw := io.Pipe()
	// Begin database export
	errGroup.Go(func() error {
		defer func(pw io.WriteCloser) {
			_ = pw.Close()
		}(pw)

		cmd, err := action.buildCommand()
		if err != nil {
			return err
		}

		if err := action.Client.Exec(ctx, kubernetes.ExecOptions{
			Pod:         action.JobPod,
			Cmd:         cmd.String(),
			Stdin:       os.Stdin,
			Stdout:      pw,
			Stderr:      plogger,
			DisablePing: true,
		}); err != nil {
			return err
		}

		return pw.Close()
	})

	if !action.RemoteGzip && action.Format == sqlformat.Gzip {
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

		if _, err := io.Copy(io.MultiWriter(f, bar), r); err != nil {
			return err
		}

		if err := f.Close(); err != nil && !errors.Is(err, os.ErrClosed) {
			return err
		}

		return nil
	})

	if err := errGroup.Wait(); err != nil {
		return err
	}

	_ = bar.Finish()

	log.WithFields(log.Fields{
		"file": action.Filename,
		"in":   time.Since(startTime).Truncate(10 * time.Millisecond),
	}).Info("dump complete")

	return nil
}

func (action Dump) buildCommand() (*command.Builder, error) {
	db, ok := action.Dialect.(config.DatabaseDump)
	if !ok {
		return nil, fmt.Errorf("%w: %s", util.ErrNoDump, action.Dialect.Name())
	}

	cmd := db.DumpCommand(action.Dump)
	if opts := viper.GetString(consts.OptsKey); opts != "" {
		cmd.Push(command.Split(opts))
	}

	if action.RemoteGzip && action.Format != sqlformat.Custom {
		cmd.Push(command.Pipe, "gzip", "--force")
	}
	log.WithField("cmd", cmd).Trace("finished building command")
	return cmd, nil
}

package restore

import (
	"context"
	"io"
	"os"
	"strings"
	"time"

	"github.com/clevyr/kubedb/internal/command"
	"github.com/clevyr/kubedb/internal/config"
	"github.com/clevyr/kubedb/internal/database/sqlformat"
	"github.com/clevyr/kubedb/internal/kubernetes"
	"github.com/clevyr/kubedb/internal/progressbar"
	gzip "github.com/klauspost/pgzip"
	log "github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

type Restore struct {
	config.Restore `mapstructure:",squash"`

	Analyze bool
}

func (action Restore) Run(ctx context.Context) (err error) {
	var f io.ReadCloser
	switch action.Filename {
	case "-":
		f = os.Stdin
	default:
		f, err = os.Open(action.Filename)
		if err != nil {
			return err
		}
		defer func(f io.ReadCloser) {
			_ = f.Close()
		}(f)
	}

	log.WithFields(log.Fields{
		"file":      action.Filename,
		"namespace": action.Client.Namespace,
		"name":      "pod/" + action.DbPod.Name,
	}).Info("ready to restore database")

	startTime := time.Now()

	bar, errLog := progressbar.New(os.Stderr, -1, "uploading", action.Spinner)
	defer bar.Close()

	errGroup, ctx := errgroup.WithContext(ctx)

	pr, pw := io.Pipe()
	errGroup.Go(func() error {
		// Connect to pod and begin piping from io.PipeReader
		defer func(pr io.ReadCloser) {
			_ = pr.Close()
		}(pr)
		return action.runInDatabasePod(ctx, pr, errLog, errLog, action.Format)
	})

	errGroup.Go(func() error {
		defer func(pw io.WriteCloser) {
			_ = pw.Close()
		}(pw)

		w := io.MultiWriter(pw, bar)

		// Clean database
		if action.Clean && action.Format != sqlformat.Custom {
			dropQuery := action.Dialect.DropDatabaseQuery(action.Database)
			if dropQuery != "" {
				log.Info("cleaning existing data")
				if action.RemoteGzip {
					err = gzipCopy(w, strings.NewReader(dropQuery))
				} else {
					_, err = io.Copy(w, strings.NewReader(dropQuery))
				}
				if err != nil {
					return err
				}
			}
		}

		// Main restore
		log.Info("restoring database")
		switch action.Format {
		case sqlformat.Gzip, sqlformat.Unknown:
			if !action.RemoteGzip {
				f, err = gzip.NewReader(f)
				if err != nil {
					return err
				}
				defer func(f io.ReadCloser) {
					_ = f.Close()
				}(f)
			}

			_, err = io.Copy(w, f)
			if err != nil {
				return err
			}
		case sqlformat.Plain, sqlformat.Custom:
			if action.RemoteGzip {
				err = gzipCopy(w, f)
			} else {
				_, err = io.Copy(w, f)
			}
			if err != nil {
				return err
			}
		}

		// Analyze query
		analyzeQuery := action.Dialect.AnalyzeQuery()
		if analyzeQuery != "" && action.Analyze {
			if action.Format == sqlformat.Custom {
				if err := pw.Close(); err != nil {
					return err
				}

				pr, pw = io.Pipe()
				defer func(pw io.WriteCloser) {
					_ = pw.Close()
				}(pw)
				w = io.MultiWriter(pw, bar)
				errGroup.Go(func() error {
					defer func(pr io.ReadCloser) {
						_ = pr.Close()
					}(pr)
					return action.runInDatabasePod(ctx, pr, errLog, errLog, sqlformat.Gzip)
				})
			}

			if action.RemoteGzip {
				err = gzipCopy(w, strings.NewReader(analyzeQuery))
			} else {
				_, err = io.Copy(w, strings.NewReader(analyzeQuery))
			}
			if err != nil {
				return err
			}
		}

		if err := pw.Close(); err != nil {
			return err
		}

		bar.Describe("finishing")
		return nil
	})

	if err := errGroup.Wait(); err != nil {
		return err
	}

	_ = bar.Finish()

	log.WithFields(log.Fields{
		"file": action.Filename,
		"in":   time.Since(startTime).Truncate(10 * time.Millisecond),
	}).Info("restore complete")
	return nil
}

func (action Restore) buildCommand(inputFormat sqlformat.Format) *command.Builder {
	cmd := action.Dialect.RestoreCommand(action.Restore, inputFormat)
	if action.RemoteGzip && action.Format != sqlformat.Custom {
		cmd.Unshift("gunzip", "--force", command.Pipe)
	}
	log.WithField("cmd", cmd).Trace("finished building command")
	return cmd
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

func (action Restore) runInDatabasePod(ctx context.Context, r *io.PipeReader, stdout, stderr io.Writer, inputFormat sqlformat.Format) error {
	defer func(r *io.PipeReader) {
		_ = r.Close()
	}(r)

	if err := action.Client.Exec(ctx, kubernetes.ExecOptions{
		Pod:         action.JobPod,
		Cmd:         action.buildCommand(inputFormat).String(),
		Stdin:       r,
		Stdout:      stdout,
		Stderr:      stderr,
		DisablePing: true,
	}); err != nil {
		return err
	}

	return nil
}

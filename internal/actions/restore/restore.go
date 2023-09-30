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
		"pod":       action.Pod.Name,
	}).Info("ready to restore database")

	startTime := time.Now()

	bar := progressbar.New(-1, "uploading")
	errLog := progressbar.NewBarSafeLogger(os.Stderr, bar)
	outLog := progressbar.NewBarSafeLogger(os.Stdout, bar)
	log.SetOutput(errLog)

	errGroup, ctx := errgroup.WithContext(ctx)

	pr, pw := io.Pipe()
	errGroup.Go(func() error {
		// Connect to pod and begin piping from io.PipeReader
		defer func(pr io.ReadCloser) {
			_ = pr.Close()
		}(pr)
		return action.runInDatabasePod(ctx, pr, outLog, errLog, action.Format)
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
		case sqlformat.Gzip, sqlformat.Custom, sqlformat.Unknown:
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
		case sqlformat.Plain:
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
					return action.runInDatabasePod(ctx, pr, outLog, errLog, sqlformat.Gzip)
				})
			}

			log.Info("running analyze query")
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

		return nil
	})

	if err := errGroup.Wait(); err != nil {
		return err
	}

	_ = bar.Finish()
	log.SetOutput(os.Stderr)

	log.WithFields(log.Fields{
		"file": action.Filename,
		"in":   time.Since(startTime).Truncate(10 * time.Millisecond),
	}).Info("restore complete")
	return nil
}

func buildCommand(conf config.Restore, inputFormat sqlformat.Format) *command.Builder {
	cmd := conf.Dialect.RestoreCommand(conf, inputFormat).
		Unshift("gunzip", "--force", command.Pipe)
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

	if err := action.Client.Exec(
		ctx,
		action.Pod,
		buildCommand(action.Restore, inputFormat).String(),
		r,
		stdout,
		stderr,
		false,
		nil,
	); err != nil {
		return err
	}

	return nil
}

package restore

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/clevyr/kubedb/internal/command"
	"github.com/clevyr/kubedb/internal/config"
	"github.com/clevyr/kubedb/internal/consts"
	"github.com/clevyr/kubedb/internal/database/sqlformat"
	"github.com/clevyr/kubedb/internal/kubernetes"
	"github.com/clevyr/kubedb/internal/progressbar"
	"github.com/clevyr/kubedb/internal/util"
	gzip "github.com/klauspost/pgzip"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"golang.org/x/sync/errgroup"
)

type Restore struct {
	config.Restore `mapstructure:",squash"`

	Analyze bool
}

func (action Restore) Run(ctx context.Context) error {
	var f io.ReadCloser
	switch action.Filename {
	case "-":
		f = os.Stdin
	default:
		var err error
		if f, err = os.Open(action.Filename); err != nil {
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
			if db, ok := action.Dialect.(config.DatabaseDbDrop); ok {
				dropQuery := db.DropDatabaseQuery(action.Database)
				log.Info("cleaning existing data")
				if action.RemoteGzip {
					if err := gzipCopy(w, strings.NewReader(dropQuery)); err != nil {
						return err
					}
				} else {
					if _, err := io.Copy(w, strings.NewReader(dropQuery)); err != nil {
						return err
					}
				}
			}
		}

		// Main restore
		log.Info("restoring database")
		switch action.Format {
		case sqlformat.Gzip, sqlformat.Unknown:
			if !action.RemoteGzip {
				var err error
				if f, err = gzip.NewReader(f); err != nil {
					return err
				}
				defer func(f io.ReadCloser) {
					_ = f.Close()
				}(f)
			}

			if _, err := io.Copy(w, f); err != nil {
				return err
			}
		case sqlformat.Plain, sqlformat.Custom:
			if action.RemoteGzip {
				if err := gzipCopy(w, f); err != nil {
					return err
				}
			} else {
				if _, err := io.Copy(w, f); err != nil {
					return err
				}
			}
		}

		// Analyze query
		if action.Analyze {
			if db, ok := action.Dialect.(config.DatabaseAnalyze); ok {
				analyzeQuery := db.AnalyzeQuery()
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
					if err := gzipCopy(w, strings.NewReader(analyzeQuery)); err != nil {
						return err
					}
				} else {
					if _, err := io.Copy(w, strings.NewReader(analyzeQuery)); err != nil {
						return err
					}
				}
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

func (action Restore) buildCommand(inputFormat sqlformat.Format) (*command.Builder, error) {
	db, ok := action.Dialect.(config.DatabaseRestore)
	if !ok {
		return nil, fmt.Errorf("%w: %s", util.ErrNoRestore, action.Dialect.Name())
	}

	cmd := db.RestoreCommand(action.Restore, inputFormat)
	if opts := viper.GetString(consts.OptsKey); opts != "" {
		cmd.Push(command.Split(opts))
	}

	if action.RemoteGzip && action.Format != sqlformat.Custom {
		cmd.Unshift("gunzip", "--force", command.Pipe)
	}
	log.WithField("cmd", cmd).Trace("finished building command")
	return cmd, nil
}

func gzipCopy(w io.Writer, r io.Reader) error {
	gzw := gzip.NewWriter(w)
	if _, err := io.Copy(gzw, r); err != nil {
		return err
	}
	return gzw.Close()
}

func (action Restore) runInDatabasePod(ctx context.Context, r *io.PipeReader, stdout, stderr io.Writer, inputFormat sqlformat.Format) error {
	defer func(r *io.PipeReader) {
		_ = r.Close()
	}(r)

	cmd, err := action.buildCommand(inputFormat)
	if err != nil {
		return err
	}

	if err := action.Client.Exec(ctx, kubernetes.ExecOptions{
		Pod:         action.JobPod,
		Cmd:         cmd.String(),
		Stdin:       r,
		Stdout:      stdout,
		Stderr:      stderr,
		DisablePing: true,
	}); err != nil {
		return err
	}

	return nil
}

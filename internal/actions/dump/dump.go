package dump

import (
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"sync/atomic"
	"time"

	"gabe565.com/utils/bytefmt"
	"gabe565.com/utils/slogx"
	"github.com/charmbracelet/lipgloss"
	"github.com/clevyr/kubedb/internal/command"
	"github.com/clevyr/kubedb/internal/config/conftypes"
	"github.com/clevyr/kubedb/internal/database/sqlformat"
	"github.com/clevyr/kubedb/internal/finalizer"
	"github.com/clevyr/kubedb/internal/github"
	"github.com/clevyr/kubedb/internal/kubernetes"
	"github.com/clevyr/kubedb/internal/notifier"
	"github.com/clevyr/kubedb/internal/progressbar"
	"github.com/clevyr/kubedb/internal/storage"
	"github.com/clevyr/kubedb/internal/tui"
	"github.com/clevyr/kubedb/internal/util"
	"github.com/muesli/termenv"
	"golang.org/x/sync/errgroup"
)

type Dump struct {
	conftypes.Dump `koanf:",squash"`
}

//nolint:funlen
func (action Dump) Run(ctx context.Context) error {
	errGroup, ctx := errgroup.WithContext(ctx)

	var f io.WriteCloser
	var rename bool
	switch {
	case action.Output == "-":
		f = os.Stdout
	case storage.IsCloud(action.Output):
		client, err := storage.NewClient(ctx, action.Output)
		if err != nil {
			return err
		}

		pr, pw := io.Pipe()
		f = pw
		defer func(pw *io.PipeWriter) {
			_ = pw.Close()
		}(pw)

		errGroup.Go(func() error {
			defer func() {
				_ = pr.Close()
			}()
			return client.PutObject(ctx, pr, action.Output)
		})
	default:
		dir := filepath.Dir(action.Output)
		if err := os.MkdirAll(dir, 0o755); err != nil && !os.IsExist(err) {
			return err
		}

		tmp, err := os.CreateTemp(dir, filepath.Base(action.Output)+"-*")
		if err != nil {
			return err
		}
		defer func() {
			_ = tmp.Close()
			_ = os.Remove(tmp.Name())
		}()

		f = tmp
		rename = true
	}

	actionLog := slog.With(
		"namespace", action.Client.Namespace,
		"pod", action.DBPod.Name,
		"file", action.Output,
	)

	actionLog.Info("Exporting database")

	if err := github.SetOutput("filename", action.Output); err != nil {
		return err
	}

	startTime := time.Now()
	bar := progressbar.New(os.Stderr, -1, "downloading", action.Progress, action.Spinner)
	defer bar.Close()

	pr, pw := io.Pipe()
	errGroup.Go(func() error {
		// Begin database export
		defer func(pw io.WriteCloser) {
			_ = pw.Close()
		}(pw)

		cmd, err := action.buildCommand()
		if err != nil {
			return err
		}

		return action.Client.Exec(ctx, kubernetes.ExecOptions{
			Pod:         action.JobPod,
			Cmd:         cmd.String(),
			Stdin:       os.Stdin,
			Stdout:      pw,
			Stderr:      bar.Logger(),
			DisablePing: true,
		})
	})

	if !action.RemoteGzip && action.Format == sqlformat.Gzip {
		// Gzip locally
		gzPipeReader, gzPipeWriter := io.Pipe()
		plainReader := pr
		errGroup.Go(func() error {
			defer func() {
				_ = gzPipeWriter.Close()
				_ = plainReader.Close()
			}()

			gzw := gzip.NewWriter(gzPipeWriter)
			if _, err := io.Copy(gzw, plainReader); err != nil {
				return err
			}
			return gzw.Close()
		})
		pr = gzPipeReader
	}

	var written atomic.Int64
	errGroup.Go(func() error {
		// Begin copying export to local file
		defer func(pr io.ReadCloser) {
			_ = pr.Close()
		}(pr)

		r := io.Reader(pr)
		if action.RemoteGzip {
			if action.Format == sqlformat.Plain {
				var err error
				if r, err = gzip.NewReader(r); err != nil {
					return err
				}
			}
		}

		n, err := io.Copy(io.MultiWriter(f, bar), r) //nolint:gosec
		written.Add(n)
		if err != nil {
			return err
		}
		return f.Close()
	})

	finalizer.Add(func(err error) {
		action.printSummary(err, time.Since(startTime).Truncate(10*time.Millisecond), written.Load())
	})

	if err := errGroup.Wait(); err != nil {
		return err
	}

	_ = bar.Finish()

	if rename {
		if f, ok := f.(*os.File); ok {
			if err := os.Rename(f.Name(), action.Output); err != nil {
				return err
			}
		}
	}

	actionLog.Info("Dump complete",
		"took", time.Since(startTime).Truncate(10*time.Millisecond),
		"size", bytefmt.Encode(written.Load()),
	)

	if handler, ok := notifier.FromContext(ctx); ok {
		if logger, ok := handler.(notifier.Logs); ok {
			logger.SetLog(
				action.summary(nil, time.Since(startTime).Truncate(10*time.Millisecond), written.Load(), true),
			)
		}
	}
	return nil
}

func (action Dump) buildCommand() (*command.Builder, error) {
	db, ok := action.Dialect.(conftypes.DBDumper)
	if !ok {
		return nil, fmt.Errorf("%w: %s", util.ErrNoDump, action.Dialect.Name())
	}

	cmd := db.DumpCommand(&action.Dump)
	if action.Opts != "" {
		cmd.Push(command.Split(action.Opts))
	}
	cmd.Unshift(command.Raw("{"))
	cmd.Push(command.Raw("|| kill $$; }"))

	if action.RemoteGzip && action.Format != sqlformat.Custom {
		cmd.Push(command.Pipe, "gzip", "--force")
	}
	slogx.Trace("Finished building command", "cmd", cmd)
	return cmd, nil
}

func (action Dump) summary(err error, took time.Duration, written int64, plain bool) string {
	var r *lipgloss.Renderer
	if plain {
		r = lipgloss.NewRenderer(os.Stdout, termenv.WithTTY(false))
		r.SetColorProfile(termenv.Ascii)
		r.SetHasDarkBackground(lipgloss.HasDarkBackground())
	}

	t := tui.MinimalTable(r).
		RowIfNotEmpty("Context", action.Context).
		Row("Namespace", tui.NamespaceStyle(r, action.Global.NamespaceColors, action.Namespace).Render()).
		Row("Pod", action.DBPod.Name).
		RowIfNotEmpty("Username", action.Username).
		RowIfNotEmpty("Database", action.Database).
		Row("File", tui.OutPath(action.Output, r)).
		Row("Took", took.String())
	if err != nil {
		t.Row("Error", tui.ErrStyle(r).Render(err.Error()))
	} else {
		t.Row("Size", bytefmt.Encode(written))
	}

	if plain {
		t.Border(lipgloss.NormalBorder())
	}

	return lipgloss.JoinVertical(lipgloss.Center,
		tui.HeaderStyle(r).Render("Dump Summary"),
		t.Render(),
	)
}

func (action Dump) printSummary(err error, took time.Duration, written int64) {
	out := os.Stdout
	if action.Output == "-" {
		out = os.Stderr
	}
	_, _ = io.WriteString(out, "\n"+action.summary(err, took, written, false)+"\n")
}

package restore

import (
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
	"sync/atomic"
	"time"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/clevyr/kubedb/internal/command"
	"github.com/clevyr/kubedb/internal/config"
	"github.com/clevyr/kubedb/internal/consts"
	"github.com/clevyr/kubedb/internal/database/sqlformat"
	"github.com/clevyr/kubedb/internal/kubernetes"
	"github.com/clevyr/kubedb/internal/log"
	"github.com/clevyr/kubedb/internal/notifier"
	"github.com/clevyr/kubedb/internal/progressbar"
	"github.com/clevyr/kubedb/internal/storage"
	"github.com/clevyr/kubedb/internal/tui"
	"github.com/clevyr/kubedb/internal/util"
	"github.com/dustin/go-humanize"
	"github.com/muesli/termenv"
	"github.com/spf13/viper"
	"golang.org/x/sync/errgroup"
)

type Restore struct {
	config.Restore `mapstructure:",squash"`

	Analyze bool
}

func (action Restore) Run(ctx context.Context) error {
	errGroup, ctx := errgroup.WithContext(ctx)

	var f io.ReadCloser
	switch {
	case action.Filename == "-":
		f = os.Stdin
	case storage.IsS3(action.Filename):
		pipe := storage.NewS3DownloadPipe()
		f = pipe
		defer func(pipe *storage.S3DownloadPipe) {
			_ = pipe.Close()
		}(pipe)

		errGroup.Go(func() error {
			return storage.DownloadS3(ctx, pipe, action.Filename)
		})
	case storage.IsGCS(action.Filename):
		var err error
		if f, err = storage.DownloadGCS(ctx, action.Filename); err != nil {
			return err
		}
		defer func(f io.ReadCloser) {
			_ = f.Close()
		}(f)
	default:
		var err error
		if f, err = os.Open(action.Filename); err != nil {
			return err
		}
		defer func(f io.ReadCloser) {
			_ = f.Close()
		}(f)
	}

	actionLog := slog.With(
		"file", action.Filename,
		"namespace", action.Client.Namespace,
		"pod", action.DBPod.Name,
	)

	actionLog.Info("Ready to restore database")

	startTime := time.Now()
	bar := progressbar.New(os.Stderr, -1, "uploading", action.Progress, action.Spinner)
	defer bar.Close()

	pr, pw := io.Pipe()
	errGroup.Go(func() error {
		// Connect to pod and begin piping from io.PipeReader
		defer func(pr io.ReadCloser) {
			_ = pr.Close()
		}(pr)
		return action.runInDatabasePod(ctx, pr, bar.Logger(), bar.Logger(), action.Format)
	})

	var written atomic.Int64
	errGroup.Go(func() error {
		defer func(pw io.WriteCloser) {
			_ = pw.Close()
		}(pw)

		w := io.MultiWriter(pw, bar)

		// Clean database
		if action.Clean && action.Format != sqlformat.Custom {
			if db, ok := action.Dialect.(config.DBDatabaseDropper); ok {
				dropQuery := db.DatabaseDropQuery(action.Database)
				actionLog.Info("Cleaning existing data")
				n, err := action.copy(w, strings.NewReader(dropQuery))
				written.Add(n)
				if err != nil {
					return err
				}
			}
		}

		// Main restore
		actionLog.Info("Restoring database")
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

			n, err := io.Copy(w, f)
			written.Add(n)
			if err != nil {
				return err
			}
		case sqlformat.Plain, sqlformat.Custom:
			n, err := action.copy(w, f)
			written.Add(n)
			if err != nil {
				return err
			}
		}

		// Analyze query
		if action.Analyze {
			if db, ok := action.Dialect.(config.DBAnalyzer); ok {
				analyzeQuery := db.AnalyzeQuery()
				if action.Format == sqlformat.Custom {
					defer func() {
						pr, pw := io.Pipe()

						errGroup.Go(func() error {
							defer func(pr io.ReadCloser) {
								_ = pr.Close()
							}(pr)
							return action.runInDatabasePod(ctx, pr, bar.Logger(), bar.Logger(), sqlformat.Gzip)
						})

						errGroup.Go(func() error {
							defer func() {
								_ = pw.Close()
							}()
							n, err := action.copy(pw, strings.NewReader(analyzeQuery))
							written.Add(n)
							return err
						})
					}()
				} else {
					n, err := action.copy(w, strings.NewReader(analyzeQuery))
					written.Add(n)
					if err != nil {
						return err
					}
				}
			}
		}

		if err := pw.Close(); err != nil {
			return err
		}

		bar.Describe("uploaded")
		return nil
	})

	util.OnFinalize(func(err error) {
		action.printSummary(err, time.Since(startTime).Truncate(10*time.Millisecond), written.Load())
	})

	if err := errGroup.Wait(); err != nil {
		return err
	}

	_ = bar.Finish()

	actionLog.Info("Restore complete",
		"took", time.Since(startTime).Truncate(10*time.Millisecond),
		"size", written.Load(),
	)

	if handler, ok := notifier.FromContext(ctx); ok {
		if logger, ok := handler.(notifier.Logs); ok {
			logger.SetLog(action.summary(nil, time.Since(startTime).Truncate(10*time.Millisecond), written.Load(), true))
		}
	}
	return nil
}

func (action Restore) buildCommand(inputFormat sqlformat.Format) (*command.Builder, error) {
	db, ok := action.Dialect.(config.DBRestorer)
	if !ok {
		return nil, fmt.Errorf("%w: %s", util.ErrNoRestore, action.Dialect.Name())
	}

	cmd := db.RestoreCommand(action.Restore, inputFormat)
	if opts := viper.GetString(consts.OptsKey); opts != "" {
		cmd.Push(command.Split(opts))
	}
	cmd.Unshift(command.Raw("{"))
	cmd.Push(command.Raw("|| { cat >/dev/null; kill $$; }; }"))

	if action.RemoteGzip {
		cmd.Unshift("gunzip", "--force", command.Pipe)
	}
	slog.Log(context.Background(), log.LevelTrace, "Finished building command", "cmd", cmd)
	return cmd, nil
}

func (action Restore) copy(w io.Writer, r io.Reader) (int64, error) {
	if action.RemoteGzip {
		gzw := gzip.NewWriter(w)
		n, err := io.Copy(gzw, r)
		if err != nil {
			return n, err
		}
		return n, gzw.Close()
	}

	n, err := io.Copy(w, r)
	return n, err
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

func (action Restore) Table(r *lipgloss.Renderer) *tui.Table {
	return tui.MinimalTable(r).
		RowIfNotEmpty("Context", action.Context).
		Row("Namespace", tui.NamespaceStyle(r, action.Namespace).Render()).
		Row("Pod", action.DBPod.Name).
		RowIfNotEmpty("Username", action.Username).
		RowIfNotEmpty("Database", action.Database)
}

func (action Restore) Confirm() (bool, error) {
	table := action.Table(nil)
	var description string
	if action.Filename != "-" && !strings.Contains(action.Filename, action.Namespace) {
		warnStyle := tui.WarnStyle(nil)
		table.Row("File", warnStyle.Render(tui.InPath(action.Filename, nil)))

		description = lipgloss.JoinVertical(lipgloss.Left,
			table.Render(),
			warnStyle.Render("WARNING: ")+
				tui.TextStyle(nil).Render("Filename does not contain the current namespace."),
			"Please verify you are restoring to the correct namespace.",
		)
	} else {
		description = table.Row("File", tui.InPath(action.Filename, nil)).Render()
	}

	theme := huh.ThemeCharm()
	theme.Focused.Description = tui.TextStyle(nil)

	var response bool
	err := tui.NewForm(huh.NewGroup(
		huh.NewConfirm().
			Title("Ready to restore?").
			Description(description).
			Value(&response),
	)).WithTheme(theme).Run()
	return response, err
}

func (action Restore) summary(err error, took time.Duration, written int64, plain bool) string {
	var r *lipgloss.Renderer
	if plain {
		r = lipgloss.NewRenderer(os.Stdout, termenv.WithTTY(false))
		r.SetColorProfile(termenv.Ascii)
		r.SetHasDarkBackground(lipgloss.HasDarkBackground())
	}

	t := action.Table(r).
		Row("File", tui.InPath(action.Filename, r)).
		Row("Took", took.String())
	if err != nil {
		t.Row("Error", tui.ErrStyle(r).Render(err.Error()))
	} else {
		t.Row("Size", humanize.IBytes(uint64(written))) //nolint:gosec
	}

	if plain {
		t.Border(lipgloss.NormalBorder())
	}

	return lipgloss.JoinVertical(lipgloss.Center,
		tui.HeaderStyle(nil).PaddingTop(1).Render("Restore Summary"),
		t.Render(),
	)
}

func (action Restore) printSummary(err error, took time.Duration, written int64) {
	out := os.Stdout
	if action.Filename == "-" {
		out = os.Stderr
	}
	_, _ = io.WriteString(out, "\n"+action.summary(err, took, written, false)+"\n")
}

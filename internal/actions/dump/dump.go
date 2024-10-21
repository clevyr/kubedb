package dump

import (
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/clevyr/kubedb/internal/command"
	"github.com/clevyr/kubedb/internal/config"
	"github.com/clevyr/kubedb/internal/consts"
	"github.com/clevyr/kubedb/internal/database/sqlformat"
	"github.com/clevyr/kubedb/internal/github"
	"github.com/clevyr/kubedb/internal/kubernetes"
	"github.com/clevyr/kubedb/internal/log"
	"github.com/clevyr/kubedb/internal/notifier"
	"github.com/clevyr/kubedb/internal/progressbar"
	"github.com/clevyr/kubedb/internal/storage"
	"github.com/clevyr/kubedb/internal/tui"
	"github.com/clevyr/kubedb/internal/util"
	"github.com/muesli/termenv"
	"github.com/spf13/viper"
	"golang.org/x/sync/errgroup"
)

type Dump struct {
	config.Dump `mapstructure:",squash"`
}

func (action Dump) Run(ctx context.Context) error {
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
			defer func() {
				_ = pr.Close()
			}()
			return storage.UploadS3(ctx, pr, action.Filename)
		})
	case storage.IsGCS(action.Filename):
		var err error
		if f, err = storage.UploadGCS(ctx, action.Filename); err != nil {
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

		var err error
		if f, err = os.Create(action.Filename); err != nil {
			return err
		}
		defer func(f io.WriteCloser) {
			_ = f.Close()
		}(f)
	}

	actionLog := slog.With(
		"namespace", action.Client.Namespace,
		"pod", action.DBPod.Name,
		"file", action.Filename,
	)

	actionLog.Info("Exporting database")

	if err := github.SetOutput("filename", action.Filename); err != nil {
		return err
	}

	startTime := time.Now()
	bar := progressbar.New(os.Stderr, -1, "downloading", action.Spinner)
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

	sizeW := &util.SizeWriter{}
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

		if _, err := io.Copy(io.MultiWriter(f, bar, sizeW), r); err != nil {
			return err
		}
		return f.Close()
	})

	util.OnFinalize(func(err error) {
		action.printSummary(err, time.Since(startTime).Truncate(10*time.Millisecond), sizeW)
	})

	if err := errGroup.Wait(); err != nil {
		return err
	}

	_ = bar.Finish()

	actionLog.Info("Dump complete",
		"took", time.Since(startTime).Truncate(10*time.Millisecond),
		"size", sizeW,
	)

	if handler, ok := notifier.FromContext(ctx); ok {
		if logger, ok := handler.(notifier.Logs); ok {
			logger.SetLog(action.summary(nil, time.Since(startTime).Truncate(10*time.Millisecond), sizeW, true))
		}
	}
	return nil
}

func (action Dump) buildCommand() (*command.Builder, error) {
	db, ok := action.Dialect.(config.DBDumper)
	if !ok {
		return nil, fmt.Errorf("%w: %s", util.ErrNoDump, action.Dialect.Name())
	}

	cmd := db.DumpCommand(action.Dump)
	if opts := viper.GetString(consts.OptsKey); opts != "" {
		cmd.Push(command.Split(opts))
	}
	cmd.Unshift(command.Raw("{"))
	cmd.Push(command.Raw("|| kill $$; }"))

	if action.RemoteGzip && action.Format != sqlformat.Custom {
		cmd.Push(command.Pipe, "gzip", "--force")
	}
	slog.Log(context.Background(), log.LevelTrace, "Finished building command", "cmd", cmd)
	return cmd, nil
}

func (action Dump) summary(err error, took time.Duration, size *util.SizeWriter, plain bool) string {
	var r *lipgloss.Renderer
	if plain {
		r = lipgloss.NewRenderer(os.Stdout, termenv.WithTTY(false))
		r.SetColorProfile(termenv.Ascii)
		r.SetHasDarkBackground(lipgloss.HasDarkBackground())
	}

	t := tui.MinimalTable(r).
		RowIfNotEmpty("Context", action.Context).
		Row("Namespace", tui.NamespaceStyle(r, action.Namespace).Render()).
		Row("Pod", action.DBPod.Name).
		RowIfNotEmpty("Username", action.Username).
		RowIfNotEmpty("Database", action.Database).
		Row("File", tui.OutPath(action.Filename, r)).
		Row("Took", took.String())
	if err != nil {
		t.Row("Error", tui.ErrStyle(r).Render(err.Error()))
	} else if size != nil {
		t.Row("Size", size.String())
	}

	if plain {
		t.Border(lipgloss.NormalBorder())
	}

	return lipgloss.JoinVertical(lipgloss.Center,
		tui.HeaderStyle(r).Render("Dump Summary"),
		t.Render(),
	)
}

func (action Dump) printSummary(err error, took time.Duration, size *util.SizeWriter) {
	out := os.Stdout
	if action.Filename == "-" {
		out = os.Stderr
	}
	_, _ = io.WriteString(out, "\n"+action.summary(err, took, size, false)+"\n")
}

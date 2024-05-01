package dump

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/clevyr/kubedb/internal/command"
	"github.com/clevyr/kubedb/internal/config"
	"github.com/clevyr/kubedb/internal/consts"
	"github.com/clevyr/kubedb/internal/database/sqlformat"
	"github.com/clevyr/kubedb/internal/github"
	"github.com/clevyr/kubedb/internal/kubernetes"
	"github.com/clevyr/kubedb/internal/notifier"
	"github.com/clevyr/kubedb/internal/progressbar"
	"github.com/clevyr/kubedb/internal/storage"
	"github.com/clevyr/kubedb/internal/tui"
	"github.com/clevyr/kubedb/internal/util"
	gzip "github.com/klauspost/pgzip"
	"github.com/muesli/termenv"
	"github.com/rs/zerolog/log"
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
			return storage.CreateS3Upload(ctx, pr, action.Filename)
		})
	case storage.IsGCS(action.Filename):
		var err error
		if f, err = storage.CreateGCSUpload(ctx, action.Filename); err != nil {
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

	actionLog := log.With().
		Str("namespace", action.Client.Namespace).
		Str("pod", action.DBPod.Name).
		Str("file", action.Filename).
		Logger()

	actionLog.Info().Msg("exporting database")

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

	sizeW := &util.SizeWriter{}

	// Begin copying export to local file
	errGroup.Go(func() error {
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

		if err := f.Close(); err != nil && !errors.Is(err, os.ErrClosed) {
			return err
		}

		return nil
	})

	util.OnFinalize(func(err error) {
		action.printSummary(err, time.Since(startTime).Truncate(10*time.Millisecond), sizeW)
	})

	if err := errGroup.Wait(); err != nil {
		return err
	}

	_ = bar.Finish()

	actionLog.Info().
		Stringer("took", time.Since(startTime).Truncate(10*time.Millisecond)).
		Stringer("size", sizeW).
		Msg("dump complete")

	if handler, ok := notifier.FromContext(ctx); ok {
		if logger, ok := handler.(notifier.Logs); ok {
			logger.SetLog(action.summary(nil, time.Since(startTime).Truncate(10*time.Millisecond), sizeW, true))
		}
	}
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
	cmd.Unshift(command.Raw("{"))
	cmd.Push(command.Raw("|| kill $$; }"))

	if action.RemoteGzip && action.Format != sqlformat.Custom {
		cmd.Push(command.Pipe, "gzip", "--force")
	}
	sanitized := strings.ReplaceAll(cmd.String(), action.Password, "***")
	log.Trace().Str("cmd", sanitized).Msg("finished building command")
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
		Row("Context", action.Context).
		Row("Namespace", action.Namespace).
		Row("Pod", action.DBPod.Name)
	if action.Username != "" {
		t.Row("Username", action.Username)
	}
	if action.Database != "" {
		t.Row("Database", action.Database)
	}
	t.Row("File", tui.OutPath(action.Filename, r)).
		Row("Took", took.String())
	if err != nil {
		t.Row("Error", lipgloss.NewStyle().Renderer(r).Foreground(lipgloss.Color("1")).Render(err.Error()))
	} else if size != nil {
		t.Row("Size", size.String())
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

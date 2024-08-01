package restore

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/clevyr/kubedb/internal/command"
	"github.com/clevyr/kubedb/internal/config"
	"github.com/clevyr/kubedb/internal/consts"
	"github.com/clevyr/kubedb/internal/database/sqlformat"
	"github.com/clevyr/kubedb/internal/kubernetes"
	"github.com/clevyr/kubedb/internal/progressbar"
	"github.com/clevyr/kubedb/internal/tui"
	"github.com/clevyr/kubedb/internal/util"
	gzip "github.com/klauspost/pgzip"
	"github.com/rs/zerolog/log"
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

	actionLog := log.With().
		Str("file", action.Filename).
		Str("namespace", action.Client.Namespace).
		Str("pod", action.DBPod.Name).
		Logger()

	actionLog.Info().Msg("ready to restore database")

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

	sizeW := &util.SizeWriter{}

	errGroup.Go(func() error {
		defer func(pw io.WriteCloser) {
			_ = pw.Close()
		}(pw)

		w := io.MultiWriter(pw, sizeW, bar)

		// Clean database
		if action.Clean && action.Format != sqlformat.Custom {
			if db, ok := action.Dialect.(config.DBDatabaseDropper); ok {
				dropQuery := db.DatabaseDropQuery(action.Database)
				actionLog.Info().Msg("cleaning existing data")
				if err := action.copy(w, strings.NewReader(dropQuery)); err != nil {
					return err
				}
			}
		}

		// Main restore
		actionLog.Info().Msg("restoring database")
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
			if err := action.copy(w, f); err != nil {
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
							return action.runInDatabasePod(ctx, pr, errLog, errLog, sqlformat.Gzip)
						})

						errGroup.Go(func() error {
							defer func() {
								_ = pw.Close()
							}()
							return action.copy(pw, strings.NewReader(analyzeQuery))
						})
					}()
				} else {
					if err := action.copy(w, strings.NewReader(analyzeQuery)); err != nil {
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
		action.printSummary(err, time.Since(startTime).Truncate(10*time.Millisecond), sizeW)
	})

	if err := errGroup.Wait(); err != nil {
		return err
	}

	_ = bar.Finish()

	actionLog.Info().
		Stringer("took", time.Since(startTime).Truncate(10*time.Millisecond)).
		Stringer("size", sizeW).
		Msg("restore complete")
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
	sanitized := strings.ReplaceAll(cmd.String(), action.Password, "***")
	log.Trace().Str("cmd", sanitized).Msg("finished building command")
	return cmd, nil
}

func (action Restore) copy(w io.Writer, r io.Reader) error {
	if action.RemoteGzip {
		gzw := gzip.NewWriter(w)
		if _, err := io.Copy(gzw, r); err != nil {
			return err
		}
		return gzw.Close()
	}

	_, err := io.Copy(w, r)
	return err
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

func (action Restore) Table() *tui.Table {
	return tui.MinimalTable(nil).
		RowIfNotEmpty("Context", action.Context).
		Row("Namespace", tui.NamespaceStyle(nil, action.Namespace).Render()).
		Row("Pod", action.DBPod.Name).
		RowIfNotEmpty("Username", action.Username).
		RowIfNotEmpty("Database", action.Database)
}

func (action Restore) Confirm() (bool, error) {
	table := action.Table()
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

func (action Restore) printSummary(err error, took time.Duration, size *util.SizeWriter) {
	t := action.Table().
		Row("File", tui.InPath(action.Filename, nil)).
		Row("Took", took.String())
	if err != nil {
		t.Row("Error", tui.ErrStyle(nil).Render(err.Error()))
	} else if size != nil {
		t.Row("Size", size.String())
	}

	fmt.Println( //nolint:forbidigo
		lipgloss.JoinVertical(lipgloss.Center,
			tui.HeaderStyle(nil).PaddingTop(1).Render("Restore Summary"),
			t.Render(),
		),
	)
}

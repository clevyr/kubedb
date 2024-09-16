package progressbar

import (
	"context"
	"io"
	"log/slog"
	"os"
	"sync"
	"time"

	"github.com/clevyr/kubedb/internal/config/flags"
	spinner "github.com/gabe565/go-spinners"
	"github.com/schollz/progressbar/v3"
	"k8s.io/kubectl/pkg/util/term"
)

func New(w io.Writer, total int64, label string, spinnerKey string) *ProgressBar {
	s, ok := spinner.Map[spinnerKey]
	if !ok {
		slog.Warn("Invalid spinner", "spinner", spinnerKey)
		s = spinner.Map[flags.DefaultSpinner]
	}

	options := []progressbar.Option{
		progressbar.OptionSetDescription(label),
		progressbar.OptionSetWriter(io.Discard),
		progressbar.OptionShowBytes(true),
		progressbar.OptionShowCount(),
		progressbar.OptionSpinnerCustom(s.Frames),
	}

	throttle := 2 * time.Second
	if (term.TTY{Out: w}).IsTerminalOut() {
		throttle = 65 * time.Millisecond
		options = append(options,
			progressbar.OptionSetRenderBlankState(true),
			progressbar.OptionOnCompletion(func() {
				_, _ = io.WriteString(os.Stderr, "\r\x1B[K")
			}),
		)
	}
	options = append(options, progressbar.OptionThrottle(throttle))

	ctx, cancel := context.WithCancel(context.Background())

	bar := &ProgressBar{
		ProgressBar: progressbar.NewOptions64(total, options...),
		cancel:      cancel,
	}
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(throttle):
				if bar.IsFinished() {
					return
				}
				if bar.mu.TryLock() {
					if !bar.logger.canOverwrite && time.Since(bar.logger.lastWrite) > 10*time.Millisecond {
						_, _ = os.Stderr.WriteString("\n")
						bar.logger.canOverwrite = true
					}
					if bar.logger.canOverwrite {
						_ = bar.RenderBlank()
						_, _ = os.Stderr.WriteString(bar.String())
					}
					bar.mu.Unlock()
				}
			}
		}
	}()

	bar.logger = NewBarSafeLogger(w, bar)
	return bar
}

type ProgressBar struct {
	*progressbar.ProgressBar
	mu     sync.Mutex
	cancel context.CancelFunc
	logger *BarSafeLogger
}

func (p *ProgressBar) Finish() error {
	defer func() {
		p.Close()
	}()
	return p.ProgressBar.Finish()
}

func (p *ProgressBar) Close() {
	p.cancel()
}

func (p *ProgressBar) Logger() io.Writer {
	return p.logger
}

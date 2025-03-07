package progressbar

import (
	"context"
	"io"
	"log/slog"
	"sync"
	"time"

	spinner "gabe565.com/spinners"
	"gabe565.com/utils/termx"
	"github.com/clevyr/kubedb/internal/consts"
	"github.com/schollz/progressbar/v3"
)

func New(w io.Writer, total int64, label string, enabled bool, spinnerKey string) *ProgressBar {
	s, ok := spinner.Map[spinnerKey]
	if !ok {
		slog.Warn("Invalid spinner", "spinner", spinnerKey)
		s = spinner.Map[consts.DefaultSpinner]
	}

	options := []progressbar.Option{
		progressbar.OptionSetDescription(label),
		progressbar.OptionSetWriter(io.Discard),
		progressbar.OptionShowBytes(true),
		progressbar.OptionShowCount(),
		progressbar.OptionSpinnerCustom(s.Frames),
	}

	throttle := 2 * time.Second
	if termx.IsTerminal(w) {
		throttle = 65 * time.Millisecond
		options = append(options,
			progressbar.OptionSetRenderBlankState(true),
			progressbar.OptionOnCompletion(func() {
				_, _ = io.WriteString(w, "\r\x1B[K")
			}),
		)
	}
	options = append(options, progressbar.OptionThrottle(throttle))

	ctx, cancel := context.WithCancel(context.Background())

	bar := &ProgressBar{
		ProgressBar: progressbar.NewOptions64(total, options...),
		cancel:      cancel,
		enabled:     enabled,
	}
	bar.logger = NewBarSafeLogger(w, bar)
	if enabled {
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
							_, _ = io.WriteString(w, "\n")
							bar.logger.canOverwrite = true
						}
						if bar.logger.canOverwrite {
							_ = bar.RenderBlank()
							_, _ = io.WriteString(w, bar.String())
						}
						bar.mu.Unlock()
					}
				}
			}
		}()
	}

	return bar
}

type ProgressBar struct {
	*progressbar.ProgressBar
	mu      sync.Mutex
	cancel  context.CancelFunc
	logger  *BarSafeLogger
	enabled bool
}

func (p *ProgressBar) Finish() error {
	defer func() {
		p.Close()
	}()
	return p.ProgressBar.Finish()
}

func (p *ProgressBar) Close() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.cancel()
}

func (p *ProgressBar) Logger() io.Writer {
	return p.logger
}

package progressbar

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"github.com/clevyr/kubedb/internal/config/flags"
	"github.com/gabe565/go-spinners"
	"github.com/mattn/go-isatty"
	"github.com/schollz/progressbar/v3"
	log "github.com/sirupsen/logrus"
)

func New(w io.Writer, max int64, label string, spinnerKey string) (*ProgressBar, *BarSafeLogger) {
	s, ok := spinner.Map[spinnerKey]
	if !ok {
		log.WithField("spinner", spinnerKey).Warn("invalid spinner")
		s = spinner.Map[flags.DefaultSpinner]
	}

	options := []progressbar.Option{
		progressbar.OptionSetDescription(label),
		progressbar.OptionSetWriter(io.Discard),
		progressbar.OptionShowBytes(true),
		progressbar.OptionSetWidth(10),
		progressbar.OptionShowCount(),
		progressbar.OptionSpinnerCustom(s.Frames),
		progressbar.OptionFullWidth(),
	}

	var throttle time.Duration
	if isatty.IsTerminal(os.Stdout.Fd()) || isatty.IsCygwinTerminal(os.Stdout.Fd()) {
		throttle = 65 * time.Millisecond
		options = append(options,
			progressbar.OptionSetRenderBlankState(true),
			progressbar.OptionOnCompletion(func() {
				_, _ = fmt.Fprint(os.Stderr, "\r\x1B[K")
			}),
		)
	} else {
		throttle = 2 * time.Second
	}
	options = append(options,
		progressbar.OptionThrottle(throttle),
	)

	cancelChan := make(chan struct{})
	bar := &ProgressBar{
		ProgressBar: progressbar.NewOptions64(max, options...),
		cancelChan:  cancelChan,
	}
	go func() {
		for {
			select {
			case <-cancelChan:
				return
			case <-time.After(throttle):
				if bar.IsFinished() {
					return
				}
				if bar.mu.TryLock() {
					if !bar.logger.canOverwrite && time.Since(bar.logger.lastWrite) > time.Millisecond {
						_, _ = os.Stderr.Write([]byte("\n"))
						bar.logger.canOverwrite = true
					}
					if bar.logger.canOverwrite {
						_ = bar.RenderBlank()
						_, _ = os.Stderr.Write([]byte(bar.String()))
					}
					bar.mu.Unlock()
				}
			}
		}
	}()

	logger := NewBarSafeLogger(w, bar)
	log.SetOutput(logger)
	return bar, logger
}

type ProgressBar struct {
	*progressbar.ProgressBar
	mu         sync.Mutex
	cancelChan chan struct{}
	cancelOnce sync.Once
	logger     BarSafeLogger
}

func (p *ProgressBar) Finish() error {
	defer func() {
		p.Close()
		log.SetOutput(os.Stderr)
	}()
	return p.ProgressBar.Finish()
}

func (p *ProgressBar) Close() {
	p.cancelOnce.Do(func() {
		close(p.cancelChan)
	})
}

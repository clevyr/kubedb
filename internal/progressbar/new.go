package progressbar

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/clevyr/kubedb/internal/config/flags"
	"github.com/gabe565/go-spinners"
	"github.com/mattn/go-isatty"
	"github.com/schollz/progressbar/v3"
	log "github.com/sirupsen/logrus"
)

func New(max int64, label string, spinnerKey string) *progressbar.ProgressBar {
	s, ok := spinner.Map[spinnerKey]
	if !ok {
		log.WithField("spinner", spinnerKey).Warn("invalid spinner")
		s = spinner.Map[flags.DefaultSpinner]
	}

	options := []progressbar.Option{
		progressbar.OptionSetDescription(label),
		progressbar.OptionSetWriter(os.Stderr),
		progressbar.OptionShowBytes(true),
		progressbar.OptionSetWidth(10),
		progressbar.OptionShowCount(),
		progressbar.OptionSpinnerCustom(s.Frames),
		progressbar.OptionFullWidth(),
	}

	if isatty.IsTerminal(os.Stdout.Fd()) || isatty.IsCygwinTerminal(os.Stdout.Fd()) {
		options = append(options,
			progressbar.OptionThrottle(65*time.Millisecond),
			progressbar.OptionSetRenderBlankState(true),
			progressbar.OptionOnCompletion(func() {
				_, _ = fmt.Fprint(os.Stderr, "\r\x1B[K")
			}),
		)
	} else {
		options = append(options,
			progressbar.OptionThrottle(2*time.Second),
		)
	}

	return progressbar.NewOptions64(max, options...)
}

func NewBarSafeLogger(w io.Writer, bar *progressbar.ProgressBar) *BarSafeLogger {
	return &BarSafeLogger{
		out: w,
		bar: bar,
	}
}

type BarSafeLogger struct {
	out io.Writer
	bar *progressbar.ProgressBar
	buf bytes.Buffer
}

func (l *BarSafeLogger) Write(p []byte) (int, error) {
	if l.bar.IsFinished() {
		return l.out.Write(p)
	}

	l.buf.Write([]byte("\r\x1B[K"))
	l.buf.Write(p)
	l.buf.WriteString(l.bar.String())
	if n, err := io.Copy(l.out, &l.buf); err != nil {
		return int(n), err
	}
	return len(p), nil
}

package progressbar

import (
	"bytes"
	"io"
	"os"
	"time"

	"github.com/mattn/go-isatty"
	"github.com/schollz/progressbar/v3"
)

func New(max int64, label string) *progressbar.ProgressBar {
	if isatty.IsTerminal(os.Stdout.Fd()) || isatty.IsCygwinTerminal(os.Stdout.Fd()) {
		return progressbar.DefaultBytes(max, label)
	}

	return progressbar.NewOptions64(
		max,
		progressbar.OptionSetDescription(label),
		progressbar.OptionSetWriter(os.Stderr),
		progressbar.OptionShowBytes(true),
		progressbar.OptionSetWidth(10),
		progressbar.OptionThrottle(2*time.Second),
		progressbar.OptionShowCount(),
		progressbar.OptionSpinnerType(14),
	)
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
}

func (l *BarSafeLogger) Write(p []byte) (int, error) {
	if l.bar.IsFinished() {
		return l.out.Write(p)
	}

	buf := bytes.NewBuffer([]byte("\r\x1B[K"))
	buf.Write(p)
	buf.WriteString(l.bar.String())
	if b, err := l.out.Write(buf.Bytes()); err != nil {
		return b, err
	}
	return len(p), nil
}

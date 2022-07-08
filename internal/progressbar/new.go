package progressbar

import (
	"bytes"
	"github.com/clevyr/kubedb/internal/terminal"
	"github.com/schollz/progressbar/v3"
	"io"
	"os"
	"time"
)

func New(max int64, label string) *progressbar.ProgressBar {
	if terminal.IsTTY() {
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

	buf := bytes.NewBuffer([]byte("\x1b[1K\r"))
	buf.Write(p)
	buf.WriteString(l.bar.String())
	if b, err := l.out.Write(buf.Bytes()); err != nil {
		return b, err
	}
	return len(p), nil
}

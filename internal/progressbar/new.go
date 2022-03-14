package progressbar

import (
	"bytes"
	"github.com/clevyr/kubedb/internal/terminal"
	"github.com/schollz/progressbar/v3"
	"io"
	"os"
	"time"
)

func New(max int64) *progressbar.ProgressBar {
	if !terminal.IsTTY() {
		return progressbar.NewOptions64(
			max,
			progressbar.OptionSetWriter(os.Stderr),
			progressbar.OptionShowBytes(true),
			progressbar.OptionSetWidth(10),
			progressbar.OptionThrottle(2*time.Second),
			progressbar.OptionShowCount(),
			progressbar.OptionSpinnerType(14),
		)
	}
	return progressbar.DefaultBytes(max)
}

func NewBarSafeLogger(w io.Writer) *BarSafeLogger {
	return &BarSafeLogger{
		out: w,
	}
}

type BarSafeLogger struct {
	out io.Writer
}

func (l *BarSafeLogger) Write(p []byte) (int, error) {
	buf := bytes.NewBuffer([]byte("\r"))
	buf.Write(p)
	return l.out.Write(buf.Bytes())
}

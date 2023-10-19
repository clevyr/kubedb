package progressbar

import (
	"bytes"
	"io"
)

func NewBarSafeLogger(w io.Writer, bar *ProgressBar) *BarSafeLogger {
	return &BarSafeLogger{
		out: w,
		bar: bar,
	}
}

type BarSafeLogger struct {
	out io.Writer
	bar *ProgressBar
	buf bytes.Buffer
}

func (l *BarSafeLogger) Write(p []byte) (int, error) {
	if l.bar.IsFinished() {
		return l.out.Write(p)
	}

	l.buf.Write([]byte("\r\x1B[K"))
	l.buf.Write(p)
	if p[len(p)-1] == '\n' {
		l.buf.WriteString(l.bar.String())
	}
	l.bar.mu.Lock()
	defer l.bar.mu.Unlock()
	if n, err := io.Copy(l.out, &l.buf); err != nil {
		return int(n), err
	}
	return len(p), nil
}

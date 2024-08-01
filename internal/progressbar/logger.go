package progressbar

import (
	"bytes"
	"io"
	"time"
)

func NewBarSafeLogger(w io.Writer, bar *ProgressBar) *BarSafeLogger {
	return &BarSafeLogger{
		out:          w,
		bar:          bar,
		canOverwrite: true,
	}
}

type BarSafeLogger struct {
	out          io.Writer
	bar          *ProgressBar
	canOverwrite bool
	lastWrite    time.Time
}

func (l *BarSafeLogger) Write(p []byte) (int, error) {
	if l.bar.IsFinished() {
		return l.out.Write(p)
	}

	l.bar.mu.Lock()
	defer l.bar.mu.Unlock()

	if l.canOverwrite {
		if _, err := io.WriteString(l.out, "\r\x1B[K"); err != nil {
			return 0, err
		}
	}

	n, err := l.out.Write(p)
	if err != nil {
		return n, err
	}

	if bytes.HasSuffix(p, []byte("\n")) {
		if _, err := io.WriteString(l.out, l.bar.String()); err != nil {
			return n, err
		}
		l.canOverwrite = true
	} else {
		l.canOverwrite = false
	}

	l.lastWrite = time.Now()
	return n, nil
}

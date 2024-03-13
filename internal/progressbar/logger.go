package progressbar

import (
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
		if _, err := l.out.Write([]byte("\r\x1B[K")); err != nil {
			return 0, err
		}
	}

	n, err := l.out.Write(p)
	if err != nil {
		return n, err
	}

	if p[len(p)-1] == '\n' {
		if _, err := l.out.Write([]byte(l.bar.String())); err != nil {
			return n, err
		}
		l.canOverwrite = true
	} else {
		l.canOverwrite = false
	}

	l.lastWrite = time.Now()
	return n, nil
}

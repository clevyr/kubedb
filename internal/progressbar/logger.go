package progressbar

import (
	"io"
)

func NewBarSafeLogger(w io.Writer, bar *ProgressBar) *BarSafeLogger {
	return &BarSafeLogger{
		out: w,
		bar: bar,
	}
}

type BarSafeLogger struct {
	out     io.Writer
	bar     *ProgressBar
	atStart bool
}

func (l *BarSafeLogger) Write(p []byte) (int, error) {
	if l.bar.IsFinished() {
		return l.out.Write(p)
	}

	l.bar.mu.Lock()
	defer l.bar.mu.Unlock()

	if !l.atStart {
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
		l.atStart = false
	} else {
		l.atStart = true
	}

	return n, nil
}

package log

import (
	"bytes"

	"github.com/rs/zerolog"
)

func NewWriter(logger zerolog.Logger, level zerolog.Level) Writer {
	return Writer{
		logger: logger,
		level:  level,
	}
}

type Writer struct {
	logger zerolog.Logger
	level  zerolog.Level
}

func (l Writer) Write(p []byte) (int, error) {
	l.logger.WithLevel(l.level).Msg(string(bytes.TrimRight(p, "\n")))
	return len(p), nil
}

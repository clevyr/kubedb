package log

import (
	"bytes"
	"context"
	"log/slog"
)

func NewWriter(logger *slog.Logger, level slog.Level) Writer {
	return Writer{
		logger: logger,
		level:  level,
	}
}

type Writer struct {
	logger *slog.Logger
	level  slog.Level
}

func (l Writer) Write(p []byte) (int, error) {
	l.logger.Log(context.Background(), l.level, string(bytes.TrimSpace(p)))
	return len(p), nil
}

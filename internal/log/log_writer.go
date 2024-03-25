package log

import log "github.com/sirupsen/logrus"

func NewWriter(logger *log.Logger, level log.Level) Writer {
	return Writer{
		logger: logger,
		level:  level,
	}
}

type Writer struct {
	logger *log.Logger
	level  log.Level
}

func (l Writer) Write(p []byte) (int, error) {
	l.logger.Log(l.level, string(p))
	return len(p), nil
}

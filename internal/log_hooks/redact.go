package log_hooks

import (
	"fmt"
	"strings"

	log "github.com/sirupsen/logrus"
)

// Redact will redact a secret from log output
type Redact string

func (r Redact) Levels() []log.Level {
	return log.AllLevels
}

func (r Redact) Fire(entry *log.Entry) error {
	entry.Message = strings.ReplaceAll(entry.Message, string(r), "***")

	for i, field := range entry.Data {
		switch field := field.(type) {
		case string:
			entry.Data[i] = strings.ReplaceAll(field, string(r), "***")
		default:
			if field, ok := field.(fmt.Stringer); ok {
				entry.Data[i] = strings.ReplaceAll(field.String(), string(r), "***")
			}
		}
	}
	return nil
}

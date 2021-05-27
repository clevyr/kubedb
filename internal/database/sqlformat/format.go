package sqlformat

import (
	"errors"
	"strings"
)

const (
	Unknown = iota
	Gzip
	Plain
	Custom
)

var UnknownFormatError = errors.New("unknown format specified")

func ParseFormat(format string) (uint8, error) {
	format = strings.ToLower(format)
	switch format {
	case "gzip", "gz", "g":
		return Gzip, nil
	case "plain", "sql", "p":
		return Plain, nil
	case "custom", "c":
		return Custom, nil
	default:
		return Unknown, UnknownFormatError
	}
}

func ParseFilename(filename string) (uint8, error) {
	filename = strings.ToLower(filename)
	switch {
	case strings.HasSuffix(filename, ".sql.gz"):
		return Gzip, nil
	case strings.HasSuffix(filename, ".dmp"):
		return Custom, nil
	case strings.HasSuffix(filename, ".sql"):
		return Plain, nil
	}
	return Unknown, UnknownFormatError
}

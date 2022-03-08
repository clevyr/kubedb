package sqlformat

import (
	"errors"
	"fmt"
	"strings"
)

type Format uint8

const (
	Unknown = iota
	Gzip
	Plain
	Custom
)

var UnknownFileFormat = errors.New("unknown file format")

func ParseFormat(format string) (Format, error) {
	format = strings.ToLower(format)
	switch format {
	case "gzip", "gz", "g":
		return Gzip, nil
	case "plain", "sql", "p":
		return Plain, nil
	case "custom", "c":
		return Custom, nil
	default:
		return Unknown, fmt.Errorf("%w: %s", UnknownFileFormat, format)
	}
}

func ParseFilename(filename string) (Format, error) {
	filename = strings.ToLower(filename)
	switch {
	case strings.HasSuffix(filename, ".sql.gz"):
		return Gzip, nil
	case strings.HasSuffix(filename, ".dmp"):
		return Custom, nil
	case strings.HasSuffix(filename, ".sql"):
		return Plain, nil
	}
	return Unknown, fmt.Errorf("%w: %s", UnknownFileFormat, filename)
}

func WriteExtension(outputFormat Format) (extension string, err error) {
	switch outputFormat {
	case Gzip:
		return ".sql.gz", nil
	case Plain:
		return ".sql", nil
	case Custom:
		return ".dmp", nil
	}
	return "", fmt.Errorf("%w: %d", UnknownFileFormat, outputFormat)
}

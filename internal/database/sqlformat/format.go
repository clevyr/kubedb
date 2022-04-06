package sqlformat

import (
	"errors"
	"fmt"
	"strings"
)

//go:generate stringer -type Format -linecomment

type Format uint8

const (
	Unknown Format = iota // unknown
	Gzip                  // gzip
	Plain                 // plain
	Custom                // custom
)

func (i *Format) Type() string {
	return "string"
}

func (i *Format) Set(s string) (err error) {
	*i, err = ParseFormat(s)
	return err
}

var UnknownFileFormat = errors.New("unknown file format")

func ParseFormat(format string) (Format, error) {
	format = strings.ToLower(format)
	switch format {
	case Gzip.String(), "gz", "g":
		return Gzip, nil
	case Plain.String(), "sql", "p":
		return Plain, nil
	case Custom.String(), "c":
		return Custom, nil
	}
	return Unknown, fmt.Errorf("%w: %s", UnknownFileFormat, format)
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

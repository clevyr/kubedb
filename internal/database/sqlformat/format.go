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

var ErrUnknown = errors.New("unknown file format")

func ParseFormat(format string) (Format, error) {
	format = strings.ToLower(format)
	switch format {
	case Gzip.String(), "archive.gz", "gz", "g":
		return Gzip, nil
	case Plain.String(), "archive", "sql", "p":
		return Plain, nil
	case Custom.String(), "c":
		return Custom, nil
	}
	return Unknown, fmt.Errorf("%w: %s", ErrUnknown, format)
}

package notifier

import (
	"errors"
	"fmt"
	"strings"
)

type Status uint8

const (
	StatusSuccess Status = iota
	StatusFailure
	StatusStart
)

var (
	ErrInvalidResponse = errors.New("invalid http response")
	ErrUnknownHandler  = errors.New("unknown handler")
	ErrEmptyUrl        = errors.New("url must be set")
)

type Notifier interface {
	Started() error
	Finished(err error) error
}

func New(handler, url string) (Notifier, error) {
	switch strings.ToLower(handler) {
	case "healthchecks":
		return NewHealthchecks(url)
	default:
		return nil, fmt.Errorf("%w: %s", ErrUnknownHandler, handler)
	}
}

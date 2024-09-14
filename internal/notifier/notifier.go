package notifier

import (
	"context"
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
	ErrEmptyURL        = errors.New("url must be set")
)

type Notifier interface {
	Started(ctx context.Context) error
	Finished(ctx context.Context, err error) error
}

type Logs interface {
	SetLog(log string)
}

func New(handler, url string) (Notifier, error) {
	switch strings.ToLower(handler) {
	case "healthchecks":
		return NewHealthchecks(url)
	default:
		return nil, fmt.Errorf("%w: %s", ErrUnknownHandler, handler)
	}
}

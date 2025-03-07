package notifier

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

var _ Logs = &Healthchecks{}

func NewHealthchecks(url string) (Notifier, error) {
	if url == "" {
		return nil, fmt.Errorf("healthchecks %w", ErrEmptyURL)
	}

	return &Healthchecks{url: url}, nil
}

type Healthchecks struct {
	PingBodyLimit int
	RunID         string

	url string
	log string
}

func (h *Healthchecks) SendStatus(ctx context.Context, status Status, log string) error {
	u, err := url.Parse(h.url)
	if err != nil {
		return err
	}

	switch status {
	case StatusStart:
		u.Path = path.Join(u.Path, "start")
	case StatusFailure:
		u.Path = path.Join(u.Path, "fail")
	}

	if h.RunID == "" {
		if u, err := uuid.NewRandom(); err == nil {
			h.RunID = u.String()
		}
	}
	q := u.Query()
	q.Set("rid", h.RunID)
	u.RawQuery = q.Encode()

	client := &http.Client{Timeout: 10 * time.Second}

	method := http.MethodHead
	if log != "" {
		method = http.MethodPost
	}

	if h.PingBodyLimit != 0 && len(log) > h.PingBodyLimit {
		log = log[len(log)-h.PingBodyLimit:]
	}

	req, err := http.NewRequestWithContext(ctx, method, u.String(), strings.NewReader(log))
	if err != nil {
		return err
	}

	var res *http.Response
	for i := range 5 {
		if res, err = client.Do(req); err == nil {
			_, _ = io.Copy(io.Discard, res.Body)
			_ = res.Body.Close()

			if res.StatusCode < 300 {
				if h.PingBodyLimit == 0 {
					if limitStr := res.Header.Get("Ping-Body-Limit"); limitStr != "" {
						if limit, err := strconv.Atoi(limitStr); err == nil {
							h.PingBodyLimit = limit
						}
					}
				}
				return nil
			}
		}

		backoff := time.Duration(i+1) * time.Duration(i+1) * time.Second
		slog.Debug("Healthchecks ping failed", "try", i+1, "backoff", backoff)
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(backoff):
		}
	}
	switch {
	case err != nil:
		return err
	case res != nil:
		return fmt.Errorf("%w: %s", ErrInvalidResponse, res.Status)
	default:
		return ErrRetriesExhausted
	}
}

func (h *Healthchecks) Started(ctx context.Context) error {
	slog.Info("Pinging Healthchecks start")
	return h.SendStatus(ctx, StatusStart, "")
}

func (h *Healthchecks) SetLog(log string) {
	h.log = log
}

func (h *Healthchecks) Finished(ctx context.Context, err error) error {
	slog.Info("Pinging Healthchecks finish")
	if err == nil {
		return h.SendStatus(ctx, StatusSuccess, h.log)
	}

	msg := "Error: " + err.Error()
	if h.log != "" {
		msg += "\n\n" + h.log
	}
	return h.SendStatus(ctx, StatusFailure, msg)
}

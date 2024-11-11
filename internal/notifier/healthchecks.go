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

	url string
	log string
}

func (h *Healthchecks) SendStatus(ctx context.Context, status Status, log string) error {
	var statusStr string
	switch status {
	case StatusStart:
		statusStr = "start"
	case StatusFailure:
		statusStr = "fail"
	}

	u, err := url.Parse(h.url)
	if err != nil {
		return err
	}

	u.Path = path.Join(u.Path, statusStr)

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

	var resp *http.Response
	for i := range 5 {
		resp, err = client.Do(req)
		_, _ = io.Copy(io.Discard, resp.Body)
		_ = resp.Body.Close()

		if err == nil && resp.StatusCode < 300 {
			if h.PingBodyLimit == 0 {
				if limitStr := resp.Header.Get("Ping-Body-Limit"); limitStr != "" {
					if limit, err := strconv.Atoi(limitStr); err == nil {
						h.PingBodyLimit = limit
					}
				}
			}
			return nil
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
	case resp != nil:
		return fmt.Errorf("%w: %s", ErrInvalidResponse, resp.Status)
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

package notifier

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
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

	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u.String(), strings.NewReader(log))
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	_, _ = io.Copy(io.Discard, resp.Body)
	_ = resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
	case http.StatusCreated:
	default:
		return fmt.Errorf("%w: %s", ErrInvalidResponse, resp.Status)
	}
	return nil
}

func (h *Healthchecks) Started(ctx context.Context) error {
	return h.SendStatus(ctx, StatusStart, "")
}

func (h *Healthchecks) SetLog(log string) {
	h.log = log
}

func (h *Healthchecks) Finished(ctx context.Context, err error) error {
	if err == nil {
		return h.SendStatus(ctx, StatusSuccess, h.log)
	}

	msg := "Error: " + err.Error()
	if h.log != "" {
		msg += "\n\n" + h.log
	}
	return h.SendStatus(ctx, StatusFailure, msg)
}

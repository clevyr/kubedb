package notifier

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"
)

func NewHealthchecks(url string) (Notifier, error) {
	if url == "" {
		return nil, fmt.Errorf("healthchecks %w", ErrEmptyUrl)
	}

	return &Healthchecks{
		url: url,
	}, nil
}

type Healthchecks struct {
	url string
}

func (h Healthchecks) SendStatus(status Status, log string) error {
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

	req, err := http.NewRequest(http.MethodPost, u.String(), strings.NewReader(log))
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

func (h Healthchecks) Started() error {
	return h.SendStatus(StatusStart, "")
}

func (h Healthchecks) Finished(err error) error {
	if err == nil {
		return h.SendStatus(StatusSuccess, "")
	} else {
		return h.SendStatus(StatusFailure, "Error: "+err.Error())
	}
}

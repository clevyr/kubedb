package util

import (
	"net/http"
	"os"
	"path/filepath"
	"runtime"
)

func NewUserAgentTransport() *UserAgentTransport {
	ua := filepath.Base(os.Args[0])
	if version := GetVersion(); version != "" {
		ua += "/v" + version
		if commit := GetCommit(); commit != "" {
			ua += "-" + commit
		}
	} else if commit := GetCommit(); commit != "" {
		ua += commit
	}
	ua += " (" + runtime.GOOS + "/" + runtime.GOARCH + ")"

	return &UserAgentTransport{
		transport: http.DefaultTransport,
		userAgent: ua,
	}
}

type UserAgentTransport struct {
	transport http.RoundTripper
	userAgent string
}

func (u *UserAgentTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	r.Header.Set("User-Agent", u.userAgent)
	return u.transport.RoundTrip(r)
}

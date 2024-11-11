package util

import (
	"os"
	"path/filepath"
	"runtime"

	"gabe565.com/utils/httpx"
)

func NewUserAgentTransport() *httpx.UserAgentTransport {
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

	return httpx.NewUserAgentTransport(nil, ua)
}

package kubernetes

import (
	"fmt"
	"strconv"
	"strings"
)

func (c *KubeClient) MinServerVersion(wantMajor, wantMinor int) (bool, error) {
	serverVersion, err := c.Discovery.ServerVersion()
	if err != nil {
		return false, err
	}

	vers := strings.TrimPrefix(serverVersion.GitVersion, "v")
	majorStr, minorStr, found := strings.Cut(vers, ".")
	if !found {
		return false, fmt.Errorf("invalid version: %s", serverVersion.GitVersion)
	}
	minorStr, _, _ = strings.Cut(minorStr, ".")

	major, err := strconv.Atoi(majorStr)
	if err != nil {
		return false, err
	}

	minor, err := strconv.Atoi(minorStr)
	if err != nil {
		return false, err
	}

	return wantMajor <= major && wantMinor <= minor, nil
}

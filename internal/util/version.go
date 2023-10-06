package util

import (
	"strings"

	"k8s.io/client-go/pkg/version"
)

func GetVersion() string {
	v := version.Get()
	result, _, _ := strings.Cut(v.GitVersion[1:], "-")
	return result
}

func GetCommit() string {
	commit := version.Get().GitCommit
	if commit == "$Format:%H$" {
		return ""
	}
	return commit
}

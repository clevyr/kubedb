package util

import "strings"

func FilterExts(exts []string, path string) bool {
	for _, ext := range exts {
		if strings.HasSuffix(path, ext) {
			return true
		}
	}
	return false
}

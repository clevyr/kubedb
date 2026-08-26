package tui

import (
	"os"
	"strings"

	"charm.land/lipgloss/v2"
)

func InPath(path string) string {
	if path == "-" {
		return "stdin"
	}
	return lipgloss.NewStyle().Italic(true).Render(CleanPath(path))
}

func OutPath(path string) string {
	if path == "-" {
		return "stdout"
	}
	return lipgloss.NewStyle().Italic(true).Render(CleanPath(path))
}

func CleanPath(path string) string {
	if cwd, err := os.Getwd(); err == nil {
		path = strings.Replace(path, cwd, ".", 1)
	}
	if home, err := os.UserHomeDir(); err == nil {
		path = strings.Replace(path, home, "~", 1)
	}
	return path
}

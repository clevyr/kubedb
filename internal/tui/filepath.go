package tui

import (
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func InPath(path string, r *lipgloss.Renderer) string {
	style := lipgloss.NewStyle().Renderer(r).Italic(true)
	if path == "-" {
		return "stdin"
	}
	return style.Render(CleanPath(path))
}

func OutPath(path string, r *lipgloss.Renderer) string {
	style := lipgloss.NewStyle().Renderer(r).Italic(true)
	if path == "-" {
		return "stdout"
	}
	return style.Render(CleanPath(path))
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

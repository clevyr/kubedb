package tui

import (
	"regexp"

	"github.com/charmbracelet/lipgloss"
)

const (
	ColorRed     = lipgloss.Color("1")
	ColorGreen   = lipgloss.Color("2")
	ColorYellow  = lipgloss.Color("3")
	ColorHiBlack = lipgloss.Color("8")
)

func HeaderStyle(r *lipgloss.Renderer) lipgloss.Style {
	if r == nil {
		r = Renderer
	}
	return lipgloss.NewStyle().Renderer(r).
		Bold(true).
		Foreground(lipgloss.AdaptiveColor{Light: "#5A56E0", Dark: "#7571F9"})
}

func NamespaceStyle(r *lipgloss.Renderer, colors map[string]string, namespace string) lipgloss.Style {
	if r == nil {
		r = Renderer
	}
	style := lipgloss.NewStyle().Renderer(r).SetString(namespace)

	for k, v := range colors {
		if regexp.MustCompile(k).MatchString(namespace) {
			style = style.Foreground(lipgloss.Color(v))
			break
		}
	}

	return style
}

func WarnStyle(r *lipgloss.Renderer) lipgloss.Style {
	if r == nil {
		r = Renderer
	}
	return lipgloss.NewStyle().Renderer(r).Foreground(ColorYellow)
}

func ErrStyle(r *lipgloss.Renderer) lipgloss.Style {
	if r == nil {
		r = Renderer
	}
	return lipgloss.NewStyle().Renderer(r).Foreground(ColorRed)
}

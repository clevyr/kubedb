package tui

import (
	"regexp"

	"charm.land/lipgloss/v2"
)

const (
	ColorRed     = lipgloss.Red
	ColorGreen   = lipgloss.Green
	ColorYellow  = lipgloss.Yellow
	ColorHiBlack = lipgloss.BrightBlack
)

func HeaderStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Bold(true).
		Foreground(lightDark(lipgloss.Color("#5A56E0"), lipgloss.Color("#7571F9")))
}

func NamespaceStyle(colors map[string]string, namespace string) lipgloss.Style {
	style := lipgloss.NewStyle().SetString(namespace)

	for k, v := range colors {
		if regexp.MustCompile(k).MatchString(namespace) {
			style = style.Foreground(lipgloss.Color(v))
			break
		}
	}

	return style
}

func WarnStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(ColorYellow)
}

func ErrStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(ColorRed)
}

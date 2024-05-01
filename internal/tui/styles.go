package tui

import "github.com/charmbracelet/lipgloss"

func HeaderStyle(r *lipgloss.Renderer) lipgloss.Style {
	return lipgloss.NewStyle().Renderer(r).
		Bold(true).
		Foreground(lipgloss.AdaptiveColor{Light: "#5A56E0", Dark: "#7571F9"})
}

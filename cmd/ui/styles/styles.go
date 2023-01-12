package styles

import "github.com/charmbracelet/lipgloss"

var (
	App = lipgloss.NewStyle().Padding(1, 2)

	Title = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FC8F02")).
		Background(lipgloss.Color("#0E2B49")).
		Padding(0, 2)

	StatusMessage = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#FF9102", Dark: "#FF9102"}).
			Render

	Err = lipgloss.NewStyle().Padding(0, 2).Foreground(lipgloss.Color("#ED567A")).Render
)

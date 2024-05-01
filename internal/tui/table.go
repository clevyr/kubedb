package tui

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
)

func BorderStyle(r *lipgloss.Renderer) lipgloss.Style {
	return lipgloss.NewStyle().Renderer(r).
		Foreground(lipgloss.AdaptiveColor{Light: "", Dark: "243"})
}

func MinimalTable(r *lipgloss.Renderer) *table.Table {
	colStyle := BorderStyle(r).Padding(0, 1)
	firstColStyle := colStyle.Copy().Align(lipgloss.Right).Bold(true)

	return table.New().
		BorderStyle(BorderStyle(r)).
		StyleFunc(func(_, col int) lipgloss.Style {
			if col == 0 {
				return firstColStyle
			}
			return colStyle
		})
}

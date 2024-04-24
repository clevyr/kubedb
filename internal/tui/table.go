package tui

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
)

func BorderStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "", Dark: "243"})
}

func MinimalTable() *table.Table {
	colStyle := BorderStyle().Padding(0, 1)
	firstColStyle := colStyle.Copy().Align(lipgloss.Right).Bold(true)

	return table.New().
		BorderStyle(BorderStyle()).
		StyleFunc(func(_, col int) lipgloss.Style {
			if col == 0 {
				return firstColStyle
			}
			return colStyle
		})
}

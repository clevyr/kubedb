package tui

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
)

type Table struct {
	*table.Table
}

func (t *Table) Row(row ...string) *Table {
	t.Table.Row(row...)
	return t
}

func (t *Table) RowIfNotEmpty(row ...string) *Table {
	for _, cell := range row {
		if cell == "" {
			return t
		}
	}
	t.Table.Row(row...)
	return t
}

func BorderStyle(r *lipgloss.Renderer) lipgloss.Style {
	return lipgloss.NewStyle().Renderer(r).
		Foreground(lipgloss.AdaptiveColor{Light: "", Dark: "243"})
}

func MinimalTable(r *lipgloss.Renderer) *Table {
	colStyle := BorderStyle(r).Padding(0, 1)
	firstColStyle := colStyle.Align(lipgloss.Right).Bold(true)

	return &Table{
		Table: table.New().
			BorderStyle(BorderStyle(r)).
			StyleFunc(func(_, col int) lipgloss.Style {
				if col == 0 {
					return firstColStyle
				}
				return colStyle
			}),
	}
}

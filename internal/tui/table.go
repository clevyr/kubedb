package tui

import (
	"slices"

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
	if slices.Contains(row, "") {
		return t
	}
	t.Table.Row(row...)
	return t
}

func BorderStyle(r *lipgloss.Renderer) lipgloss.Style {
	if r == nil {
		r = Renderer
	}
	return lipgloss.NewStyle().Renderer(r).
		Foreground(lipgloss.AdaptiveColor{Light: "246", Dark: "241"})
}

func TextStyle(r *lipgloss.Renderer) lipgloss.Style {
	if r == nil {
		r = Renderer
	}
	return lipgloss.NewStyle().Renderer(r).
		Foreground(lipgloss.AdaptiveColor{Light: "234", Dark: "250"})
}

func MinimalTable(r *lipgloss.Renderer) *Table {
	if r == nil {
		r = Renderer
	}
	colStyle := TextStyle(r).Padding(0, 1)
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

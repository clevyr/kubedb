package tui

import (
	"slices"

	"charm.land/lipgloss/v2"
	"charm.land/lipgloss/v2/table"
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

func BorderStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(lightDark(lipgloss.Color("246"), lipgloss.Color("241")))
}

func TextStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(lightDark(lipgloss.Color("234"), lipgloss.Color("250")))
}

func MinimalTable() *Table {
	colStyle := TextStyle().Padding(0, 1)
	firstColStyle := colStyle.Align(lipgloss.Right).Bold(true)

	return &Table{
		Table: table.New().
			BorderStyle(BorderStyle()).
			StyleFunc(func(_, col int) lipgloss.Style {
				if col == 0 {
					return firstColStyle
				}
				return colStyle
			}),
	}
}

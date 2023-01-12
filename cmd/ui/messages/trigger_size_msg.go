package messages

import (
	tea "github.com/charmbracelet/bubbletea"
)

func TriggerSizeMsg(w, h int) tea.Cmd {
	return func() tea.Msg {
		return tea.WindowSizeMsg{
			Width:  w,
			Height: h,
		}
	}
}

package messages

import (
	tea "github.com/charmbracelet/bubbletea"
	"time"
)

func DispatchKey(key tea.KeyType, sleep time.Duration) tea.Cmd {
	return func() tea.Msg {
		time.Sleep(sleep)
		return tea.KeyMsg{Type: key}
	}
}

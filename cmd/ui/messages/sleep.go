package messages

import (
	tea "github.com/charmbracelet/bubbletea"
	"time"
)

func Sleep(duration time.Duration, msg tea.Msg) tea.Cmd {
	return func() tea.Msg {
		time.Sleep(duration)
		return msg
	}
}

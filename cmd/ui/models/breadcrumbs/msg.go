package breadcrumbs

import tea "github.com/charmbracelet/bubbletea"

func NewActiveMsg(num int) tea.Msg {
	return ActiveMsg{Num: num}
}

type ActiveMsg struct {
	Num int
}

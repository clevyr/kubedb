package messages

import tea "github.com/charmbracelet/bubbletea"

type ViewDir int

const (
	DirNextView ViewDir = iota
	DirPrevView
)

func ChangeView(dir ViewDir, times int) tea.Cmd {
	return func() tea.Msg {
		return ChangeViewMsg{Direction: dir, Times: times}
	}
}

func NextView() tea.Cmd {
	return ChangeView(DirNextView, 1)
}

func PrevView() tea.Cmd {
	return ChangeView(DirPrevView, 1)
}

type ChangeViewMsg struct {
	Direction ViewDir
	Times     int
}

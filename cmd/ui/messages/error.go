package messages

import tea "github.com/charmbracelet/bubbletea"

func NewErrMsg(title string, err error) tea.Msg {
	return ErrMsg{
		title: title,
		error: err,
	}
}

type ErrMsg struct {
	title string
	error
}

func (e ErrMsg) Title() string { return e.title }

func (e ErrMsg) Error() string { return e.error.Error() }

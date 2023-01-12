package error_view

import (
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/clevyr/kubedb/cmd/ui/messages"
	"github.com/clevyr/kubedb/cmd/ui/styles"
	"github.com/muesli/reflow/wordwrap"
)

func NewModel(title string, err error) Model {
	return Model{
		title: title,
		err:   err,
		keys:  keys,
		help:  help.New(),
	}
}

type Model struct {
	title string
	err   error
	keys  keyMap
	help  help.Model
	w, h  int
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.w, m.h = msg.Width, msg.Height

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.quit):
			return m, tea.Quit
		case key.Matches(msg, m.keys.back):
			return m, messages.PrevView()
		}
	}
	return m, nil
}

func (m Model) View() string {
	s := styles.Err(wordwrap.String(m.title+".\n\n"+m.err.Error(), m.w))
	s += "\n\n"
	s += m.help.View(m.keys)
	return s
}

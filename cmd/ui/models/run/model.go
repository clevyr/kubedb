package run

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/clevyr/kubedb/cmd/ui/config"
)

func NewModel(conf *config.Config) tea.Model {
	return Model{
		conf: conf,
	}
}

type Model struct {
	conf *config.Config
}

func (m Model) Init() tea.Cmd {
	m.conf.Run = true
	return tea.Quit
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m, tea.Quit
}

func (m Model) View() string {
	return ""
}

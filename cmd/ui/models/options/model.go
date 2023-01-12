package options

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/clevyr/kubedb/cmd/ui/config"
	"github.com/clevyr/kubedb/cmd/ui/delegate"
	"github.com/clevyr/kubedb/cmd/ui/messages"
	"github.com/clevyr/kubedb/cmd/ui/styles"
	zone "github.com/lrstanley/bubblezone"
	"github.com/spf13/pflag"
	"time"
)

func NewModel(conf *config.Config) tea.Model {
	delegateKeys := delegate.NewKeyMap()

	var flags []list.Item
	conf.RootCmd.Flags().VisitAll(func(flag *pflag.Flag) {
		if !flag.Hidden && flag.Name != "help" && flag.Name != "kubeconfig" && flag.Name != "context" && flag.Name != "log-format" && flag.Name != "log-level" && flag.Name != "namespace" {
			flags = append(flags, Item{Flag: flag})
		}
	})

	d := delegate.NewDelegate(delegateKeys)
	l := list.New(flags, d, 0, 0)
	l.Title = "Choose options"
	l.Styles.Title = styles.Title
	l.SetStatusBarItemName("option", "options")

	return Model{
		conf:         conf,
		list:         l,
		delegateKeys: delegateKeys,
	}
}

type Model struct {
	conf         *config.Config
	list         list.Model
	delegateKeys *delegate.KeyMap
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(tea.EnterAltScreen)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetSize(msg.Width, msg.Height)

	case tea.KeyMsg:
		if m.list.FilterState() == list.Filtering {
			break
		}

		switch {
		case key.Matches(msg, m.delegateKeys.Back):
			return m, messages.PrevView()
		case key.Matches(msg, m.delegateKeys.Choose):
			//m.conf.Cmd = m.list.SelectedItem().(Item).Command
			return m, messages.NextView()
		}

	case tea.MouseMsg:
		if msg.Type == tea.MouseWheelUp {
			m.list.CursorUp()
			return m, nil
		}

		if msg.Type == tea.MouseWheelDown {
			m.list.CursorDown()
			return m, nil
		}

		if msg.Type == tea.MouseLeft {
			for i, listItem := range m.list.VisibleItems() {
				item, _ := listItem.(Item)
				if zone.Get(item.Name).InBounds(msg) {
					m.list.Select(i)
					return m, messages.DispatchKey(tea.KeyEnter, 200*time.Millisecond)
				}
			}
		}

		return m, nil
	}

	newListModel, cmd := m.list.Update(msg)
	m.list = newListModel
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	return m.list.View()
}

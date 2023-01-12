package namespaces

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/clevyr/kubedb/cmd/ui/config"
	"github.com/clevyr/kubedb/cmd/ui/delegate"
	"github.com/clevyr/kubedb/cmd/ui/messages"
	"github.com/clevyr/kubedb/cmd/ui/styles"
	zone "github.com/lrstanley/bubblezone"
	"time"
)

func NewModel(conf *config.Config) tea.Model {
	delegateKeys := delegate.NewKeyMap()

	d := delegate.NewDelegate(delegateKeys)
	d.ShowDescription = false

	l := list.New([]list.Item{}, d, 0, 0)
	l.Title = "Fetching namespaces..."
	l.Styles.Title = styles.Title
	l.SetStatusBarItemName("namespace", "namespaces")
	l.StartSpinner()

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
	return tea.Batch(tea.EnterAltScreen, m.list.StartSpinner(), List(m.conf))
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case GetNamespaceMsg:
		items := make([]list.Item, 0, len(msg.Namespaces))
		for _, ns := range msg.Namespaces {
			items = append(items, Item(ns.Name))
		}
		cmds = append(cmds, m.list.SetItems(items))
		m.list.StopSpinner()
		m.list.Title = "Choose a namespace"

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
			if err := m.conf.RootCmd.Flags().Set("namespace", string(m.list.SelectedItem().(Item))); err != nil {
				panic(err)
			}
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
				if zone.Get(string(item)).InBounds(msg) {
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

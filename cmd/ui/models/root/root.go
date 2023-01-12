package root

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/clevyr/kubedb/cmd/ui/config"
	"github.com/clevyr/kubedb/cmd/ui/messages"
	"github.com/clevyr/kubedb/cmd/ui/models/actions"
	"github.com/clevyr/kubedb/cmd/ui/models/breadcrumbs"
	"github.com/clevyr/kubedb/cmd/ui/models/contexts"
	"github.com/clevyr/kubedb/cmd/ui/models/error_view"
	"github.com/clevyr/kubedb/cmd/ui/models/namespaces"
	"github.com/clevyr/kubedb/cmd/ui/models/options"
	"github.com/clevyr/kubedb/cmd/ui/models/run"
	"github.com/clevyr/kubedb/cmd/ui/styles"
	"github.com/clevyr/kubedb/internal/util"
	zone "github.com/lrstanley/bubblezone"
)

var ViewFuncs = []func(conf *config.Config) tea.Model{
	actions.NewModel,
	contexts.NewModel,
	namespaces.NewModel,
	options.NewModel,
	run.NewModel,
}

func NewModel(conf *config.Config) *Model {
	m := &Model{
		conf:      conf,
		prevViews: util.NewStack[tea.Model](0, len(ViewFuncs)),
		crumbs:    breadcrumbs.NewModel("Actions", "Contexts", "Namespaces", "Options", "Run"),
	}
	m.view = ViewFuncs[m.prevViews.Len()](conf)
	return m
}

type Model struct {
	conf *config.Config
	w, h int

	prevViews util.Stack[tea.Model]
	view      tea.Model

	crumbs tea.Model
}

func (m *Model) Init() tea.Cmd {
	return tea.Batch(tea.EnterAltScreen)
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case messages.ErrMsg:
		m.view = error_view.NewModel(msg.Title(), msg)
		cmds = append(cmds, m.view.Init(), messages.TriggerSizeMsg(m.w, m.h))
		return m, tea.Batch(cmds...)

	case tea.WindowSizeMsg:
		m.w, m.h = msg.Width, msg.Height
		h, v := styles.App.GetFrameSize()
		msg.Width -= h
		msg.Height -= v
		msg.Height -= breadcrumbs.Height
		m.view, cmd = m.view.Update(msg)
		cmds = append(cmds, cmd)
		m.crumbs, cmd = m.crumbs.Update(msg)
		cmds = append(cmds, cmd)
		return m, tea.Batch(cmds...)

	case messages.ChangeViewMsg:
		switch msg.Direction {
		case messages.DirNextView:
			m.prevViews.Push(m.view)
			if m.prevViews.Len() > len(ViewFuncs)-1 {
				return m, tea.Quit
			}
			m.view = ViewFuncs[m.prevViews.Len()](m.conf)
			cmds = append(cmds, m.view.Init())

		case messages.DirPrevView:
			var view tea.Model
			for i := 0; i < msg.Times; i += 1 {
				var ok bool
				view, ok = m.prevViews.Pop()
				if !ok {
					return m, tea.Quit
				}
			}
			m.view = view
		}
		cmds = append(cmds, messages.TriggerSizeMsg(m.w, m.h))

		m.crumbs, cmd = m.crumbs.Update(breadcrumbs.NewActiveMsg(m.prevViews.Len()))
		cmds = append(cmds, cmd)

		return m, tea.Batch(cmds...)
	}

	m.view, cmd = m.view.Update(msg)
	cmds = append(cmds, cmd)

	m.crumbs, cmd = m.crumbs.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m *Model) View() string {
	return zone.Scan(styles.App.Render(lipgloss.JoinVertical(
		lipgloss.Top,
		m.crumbs.View(),
		m.view.View(),
	)))
}

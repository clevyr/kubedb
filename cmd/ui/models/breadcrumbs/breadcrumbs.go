package breadcrumbs

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/clevyr/kubedb/cmd/ui/messages"
	zone "github.com/lrstanley/bubblezone"
)

const Height = 4

var (
	activeTabBorder = lipgloss.Border{
		Top:         "─",
		Bottom:      " ",
		Left:        "│",
		Right:       "│",
		TopLeft:     "╭",
		TopRight:    "╮",
		BottomLeft:  "┘",
		BottomRight: "└",
	}

	tabBorder = lipgloss.Border{
		Top:         "─",
		Bottom:      "─",
		Left:        "│",
		Right:       "│",
		TopLeft:     "╭",
		TopRight:    "╮",
		BottomLeft:  "┴",
		BottomRight: "┴",
	}

	tab = lipgloss.NewStyle().
		Border(tabBorder, true).
		BorderForeground(lipgloss.Color("#E97724")).
		Padding(0, 1)

	disabledTab = tab.Copy().
			Foreground(lipgloss.AdaptiveColor{Light: "#757575", Dark: "#616161"})

	activeTab = tab.Copy().Border(activeTabBorder, true)

	tabGap = tab.Copy().
		BorderTop(false).
		BorderLeft(false).
		BorderRight(false).
		Align(lipgloss.Right)
)

func NewModel(crumbs ...string) tea.Model {
	return Model{
		crumbs: crumbs,
	}
}

type Model struct {
	crumbs []string
	active int
	width  int
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width

	case tea.MouseMsg:
		if msg.Type == tea.MouseLeft {
			for i, crumb := range m.crumbs {
				if zone.Get(crumb).InBounds(msg) {
					return m, messages.ChangeView(messages.DirPrevView, m.active-i)
				}
			}
		}

	case ActiveMsg:
		m.active = msg.Num
	}

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	tabs := make([]string, 0, len(m.crumbs))
	for i, crumb := range m.crumbs {
		if i == m.active {
			tabs = append(tabs, activeTab.Render(crumb))
		} else if i < m.active {
			tabs = append(tabs, zone.Mark(crumb, tab.Render(crumb)))
		} else {
			tabs = append(tabs, disabledTab.Render(crumb))
		}
	}
	row := lipgloss.JoinHorizontal(lipgloss.Top, tabs...)
	title := activeTab.Render("KubeDB")
	gap := tabGap.Width(max(0, m.width-lipgloss.Width(row)) - lipgloss.Width(title)).Render("")
	return lipgloss.JoinHorizontal(lipgloss.Bottom, row, gap, title) + "\n"
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

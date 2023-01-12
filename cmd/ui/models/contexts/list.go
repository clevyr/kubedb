package contexts

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/clevyr/kubedb/cmd/ui/config"
	"github.com/clevyr/kubedb/cmd/ui/messages"
	"github.com/clevyr/kubedb/internal/kubernetes"
	"sort"
	"strings"
)

func List(conf *config.Config) tea.Cmd {
	return func() tea.Msg {
		configLoader := kubernetes.NewConfigLoader(conf.Kubeconfig, "")

		raw, err := configLoader.RawConfig()
		if err != nil {
			return messages.NewErrMsg("Failed to list contexts", err)
		}

		items := make([]Item, 0, len(raw.Contexts))
		for name, context := range raw.Contexts {
			items = append(items, Item{name, context})
		}

		sort.Slice(items, func(i, j int) bool {
			return strings.ToLower(items[i].Name) < strings.ToLower(items[j].Name)
		})

		return GetContextMsg{Contexts: items}
	}
}

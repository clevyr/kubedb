package contexts

import (
	zone "github.com/lrstanley/bubblezone"
	"k8s.io/client-go/tools/clientcmd/api"
)

type Item struct {
	Name    string
	Context *api.Context
}

func (i Item) Title() string { return zone.Mark(i.Name, i.Name) }

func (i Item) Description() string {
	if i.Context != nil {
		return "Cluster: " + i.Context.Cluster
	}
	return ""
}

func (i Item) FilterValue() string { return i.Title() }

package namespaces

import (
	"context"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/clevyr/kubedb/cmd/ui/config"
	"github.com/clevyr/kubedb/cmd/ui/messages"
	kubernetes2 "github.com/clevyr/kubedb/internal/kubernetes"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"sort"
	"strings"
)

func List(conf *config.Config) tea.Cmd {
	return func() tea.Msg {
		configLoader := kubernetes2.NewConfigLoader(conf.Kubeconfig, conf.Context)

		clientConfig, err := configLoader.ClientConfig()
		if err != nil {
			return messages.NewErrMsg("Failed to create client config", err)
		}

		clientSet, err := kubernetes.NewForConfig(clientConfig)
		if err != nil {
			return messages.NewErrMsg("Failed to create clientset", err)
		}

		namespaces, err := clientSet.CoreV1().Namespaces().List(context.Background(), metav1.ListOptions{})
		if err != nil {
			return messages.NewErrMsg("Failed to list namespaces", err)
		}

		items := namespaces.Items
		sort.Slice(items, func(i, j int) bool {
			return strings.ToLower(items[i].Name) < strings.ToLower(items[j].Name)
		})

		return GetNamespaceMsg{Namespaces: items}
	}
}

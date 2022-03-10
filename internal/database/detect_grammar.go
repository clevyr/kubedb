package database

import (
	"errors"
	"github.com/clevyr/kubedb/internal/config"
	"github.com/clevyr/kubedb/internal/database/grammar"
	"github.com/clevyr/kubedb/internal/kubernetes"
	v1 "k8s.io/api/core/v1"
)

var ErrDatabaseNotFound = errors.New("could not detect a database")

func DetectGrammar(client kubernetes.KubeClient) (config.Databaser, []v1.Pod, error) {
	grammars := []config.Databaser{
		grammar.Postgres{},
		grammar.MariaDB{},
	}

	pods, err := client.GetNamespacedPods()
	if err != nil {
		return nil, []v1.Pod{}, err
	}

	for _, g := range grammars {
		pods, err := client.FilterPodList(pods, g.PodLabels())
		if err == nil {
			return g, pods, nil
		}
	}
	return nil, []v1.Pod{}, ErrDatabaseNotFound
}

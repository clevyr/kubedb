package database

import (
	"errors"
	"github.com/clevyr/kubedb/internal/config"
	"github.com/clevyr/kubedb/internal/database/dialect"
	"github.com/clevyr/kubedb/internal/kubernetes"
	v1 "k8s.io/api/core/v1"
)

var ErrDatabaseNotFound = errors.New("could not detect a database")

func DetectDialect(client kubernetes.KubeClient) (config.Databaser, []v1.Pod, error) {
	dialects := []config.Databaser{
		dialect.Postgres{},
		dialect.MariaDB{},
	}

	pods, err := client.GetNamespacedPods()
	if err != nil {
		return nil, []v1.Pod{}, err
	}

	for _, g := range dialects {
		pods, err := client.FilterPodList(pods, g.PodLabels())
		if err == nil {
			return g, pods, nil
		}
	}
	return nil, []v1.Pod{}, ErrDatabaseNotFound
}

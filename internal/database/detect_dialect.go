package database

import (
	"context"
	"errors"

	"github.com/clevyr/kubedb/internal/config"
	"github.com/clevyr/kubedb/internal/database/dialect"
	"github.com/clevyr/kubedb/internal/kubernetes"
	v1 "k8s.io/api/core/v1"
)

var ErrDatabaseNotFound = errors.New("could not detect a database")

func DetectDialect(ctx context.Context, client kubernetes.KubeClient) (config.Databaser, []v1.Pod, error) {
	pods, err := client.GetNamespacedPods(ctx)
	if err != nil {
		return nil, []v1.Pod{}, err
	}

	for _, g := range dialect.All() {
		pods, err := client.FilterPodList(pods, g.PodLabels())
		if err == nil {
			return g, pods, nil
		}
	}
	return nil, []v1.Pod{}, ErrDatabaseNotFound
}

func DetectDialectFromPod(pod v1.Pod) (config.Databaser, error) {
	for _, g := range dialect.All() {
		for _, v := range g.PodLabels() {
			if v.Matches(pod) {
				return g, nil
			}
		}
	}
	return nil, ErrDatabaseNotFound
}

package database

import (
	"context"
	"errors"

	"github.com/clevyr/kubedb/internal/config"
	"github.com/clevyr/kubedb/internal/kubernetes"
	v1 "k8s.io/api/core/v1"
)

var ErrDatabaseNotFound = errors.New("could not detect a database")

func DetectDialect(ctx context.Context, client kubernetes.KubeClient) (config.Database, []v1.Pod, error) {
	podList, err := client.GetNamespacedPods(ctx)
	if err != nil {
		return nil, []v1.Pod{}, err
	}

	for _, db := range All() {
		pods := kubernetes.FilterPodList(podList.Items, db.PodLabels())
		if len(pods) != 0 {
			return db, pods, nil
		}
	}
	return nil, []v1.Pod{}, ErrDatabaseNotFound
}

func DetectDialectFromPod(pod v1.Pod) (config.Database, error) {
	for _, db := range All() {
		if db.PodLabels().Matches(pod) {
			return db, nil
		}
	}
	return nil, ErrDatabaseNotFound
}

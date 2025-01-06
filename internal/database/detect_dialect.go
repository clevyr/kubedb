package database

import (
	"context"
	"errors"

	"github.com/clevyr/kubedb/internal/config"
	"github.com/clevyr/kubedb/internal/kubernetes"
	corev1 "k8s.io/api/core/v1"
)

var ErrDatabaseNotFound = errors.New("could not detect a database")

type DetectResult map[config.Database][]corev1.Pod

func DetectDialect(ctx context.Context, client kubernetes.KubeClient) (DetectResult, error) {
	podList, err := client.GetNamespacedPods(ctx)
	if err != nil {
		return nil, err
	}

	result := make(DetectResult)
	for _, db := range All() {
		pods := kubernetes.FilterPodList(podList.Items, db.PodFilters())
		if len(pods) != 0 {
			result[db] = pods
		}
	}
	if len(result) == 0 {
		return nil, ErrDatabaseNotFound
	}
	if len(result) > 1 {
		// Find the highest priority dialects
		var maxPriority uint8
		for dialect := range result {
			if dbPriority, ok := dialect.(config.DBOrderer); ok {
				priority := dbPriority.Priority()
				if maxPriority < priority {
					maxPriority = priority
				}
			}
		}
		if maxPriority != 0 {
			// Filter out dialects that are lower than the max
			for dialect := range result {
				if dbPriority, ok := dialect.(config.DBOrderer); ok {
					priority := dbPriority.Priority()
					if priority < maxPriority {
						delete(result, dialect)
					}
				} else {
					delete(result, dialect)
				}
			}
		}
	}
	return result, nil
}

func DetectDialectFromPod(pod corev1.Pod) (config.Database, error) {
	for _, db := range All() {
		if db.PodFilters().Matches(pod) {
			return db, nil
		}
	}
	return nil, ErrDatabaseNotFound
}

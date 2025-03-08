package database

import (
	"context"
	"errors"
	"slices"

	"github.com/clevyr/kubedb/internal/config/conftypes"
	"github.com/clevyr/kubedb/internal/kubernetes"
	corev1 "k8s.io/api/core/v1"
)

var ErrDatabaseNotFound = errors.New("could not detect a database")

type DetectResult struct {
	Dialect conftypes.Database
	Pods    []corev1.Pod
}

func DetectDialect(ctx context.Context, client kubernetes.KubeClient) ([]DetectResult, error) {
	podList, err := client.GetNamespacedPods(ctx)
	if err != nil {
		return nil, err
	}

	dialects := All()
	result := make([]DetectResult, 0, len(dialects))
	var maxPriority uint8
	for _, db := range dialects {
		if pods := kubernetes.FilterPodList(podList.Items, db.PodFilters()); len(pods) != 0 {
			result = append(result, DetectResult{db, pods})

			// Find the highest priority dialects
			if dbPriority, ok := db.(conftypes.DBOrderer); ok {
				if priority := dbPriority.Priority(); maxPriority < priority {
					maxPriority = priority
				}
			}
		}
	}

	switch len(result) {
	case 0:
		return nil, ErrDatabaseNotFound
	case 1:
	default:
		if maxPriority != 0 {
			// Filter out dialects that are lower than the max
			result = slices.DeleteFunc(result, func(v DetectResult) bool {
				if dbPriority, ok := v.Dialect.(conftypes.DBOrderer); ok {
					return dbPriority.Priority() < maxPriority
				}
				return true
			})
		}
	}
	return result, nil
}

func DetectDialectFromPod(pod corev1.Pod) (conftypes.Database, error) {
	for _, db := range All() {
		if db.PodFilters().Matches(pod) {
			return db, nil
		}
	}
	return nil, ErrDatabaseNotFound
}

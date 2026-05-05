package discovery

import (
	"context"
	"log/slog"
	"strings"

	"github.com/clevyr/kubedb/internal/config/conftypes"
	"github.com/clevyr/kubedb/internal/database"
	"github.com/clevyr/kubedb/internal/kubernetes"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Result struct {
	Dialect conftypes.Database
	Pods    []corev1.Pod
}

func Discover(ctx context.Context, client kubernetes.KubeClient, podName, dialectName string) ([]Result, error) {
	var pods []corev1.Pod
	if podName != "" {
		if slashIdx := strings.IndexRune(podName, '/'); slashIdx != 0 && slashIdx+1 < len(podName) {
			podName = podName[slashIdx+1:]
		}
		pod, err := client.Pods().Get(ctx, podName, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}
		pods = []corev1.Pod{*pod}
	}

	var results []Result
	switch {
	case dialectName != "":
		dialect, err := database.New(dialectName)
		if err != nil {
			return nil, err
		}
		slog.Debug("Configured database", "dialect", dialect.Name())
		if len(pods) == 0 {
			pods, err = client.GetPodsFiltered(ctx, dialect.PodFilters())
			if err != nil {
				return nil, err
			}
			if len(pods) == 0 {
				return nil, kubernetes.ErrPodNotFound
			}
		}
		results = []Result{{Dialect: dialect, Pods: pods}}
	case len(pods) == 0:
		detected, err := database.DetectDialect(ctx, client)
		if err != nil {
			return nil, err
		}
		results = make([]Result, len(detected))
		for i, v := range detected {
			results[i] = Result{Dialect: v.Dialect, Pods: v.Pods}
		}
	default:
		dialect, err := database.DetectDialectFromPod(pods[0])
		if err != nil {
			return nil, err
		}
		results = []Result{{Dialect: dialect, Pods: pods}}
	}

	for i, r := range results {
		if len(r.Pods) > 1 {
			if db, ok := r.Dialect.(conftypes.DBFilterer); ok && podName == "" {
				filtered, err := db.FilterPods(ctx, client, r.Pods)
				if err != nil {
					slog.Warn("Could not query primary instance", "error", err)
				} else if len(filtered) != 0 {
					results[i].Pods = filtered
				}
			}
		}
	}

	return results, nil
}

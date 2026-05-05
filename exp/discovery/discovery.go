package discovery

import (
	"context"

	internaldiscovery "github.com/clevyr/kubedb/internal/discovery"
	"github.com/clevyr/kubedb/internal/kubernetes"
	corev1 "k8s.io/api/core/v1"
	k8s "k8s.io/client-go/kubernetes"
)

type Result struct {
	Dialect    string
	PrettyName string
	Pods       []corev1.Pod
}

type Options struct {
	Dialect string
	PodName string
}

// Discover finds database pods in the given namespace.
// When multiple database types are found and no Dialect is specified,
// multiple results are returned.
func Discover(ctx context.Context, clientset k8s.Interface, namespace string, opts Options) ([]Result, error) {
	client := kubernetes.KubeClient{
		ClientSet: clientset,
		Namespace: namespace,
	}

	internal, err := internaldiscovery.Discover(ctx, client, opts.PodName, opts.Dialect)
	if err != nil {
		return nil, err
	}

	results := make([]Result, len(internal))
	for i, r := range internal {
		results[i] = Result{
			Dialect:    r.Dialect.Name(),
			PrettyName: r.Dialect.PrettyName(),
			Pods:       r.Pods,
		}
	}
	return results, nil
}

package conftypes

import "github.com/clevyr/kubedb/internal/kubernetes"

type Kubernetes struct {
	Kubeconfig string                `koanf:"kubeconfig"`
	Context    string                `koanf:"context"`
	Namespace  string                `koanf:"namespace"`
	Client     kubernetes.KubeClient `koanf:"-"`
}

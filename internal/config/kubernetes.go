package config

import "github.com/clevyr/kubedb/internal/kubernetes"

type Kubernetes struct {
	Kubeconfig string
	Context    string
	Namespace  string
	Client     kubernetes.KubeClient
}

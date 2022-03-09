package config

import (
	"github.com/clevyr/kubedb/internal/kubernetes"
	v1 "k8s.io/api/core/v1"
)

type Global struct {
	Client    kubernetes.KubeClient
	Databaser Databaser
	Pod       v1.Pod
	Database  string
	Username  string
	Password  string
}

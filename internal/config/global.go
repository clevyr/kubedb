package config

import (
	"github.com/clevyr/kubedb/internal/kubernetes"
	v1 "k8s.io/api/core/v1"
)

type Global struct {
	Kubernetes
	Client     kubernetes.KubeClient
	Dialect    Databaser `mapstructure:"-"`
	Pod        v1.Pod    `mapstructure:"-"`
	Database   string
	Username   string
	Password   string
	Quiet      bool
	RemoteGzip bool `mapstructure:"remote-gzip"`
}

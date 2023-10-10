package config

import (
	"github.com/clevyr/kubedb/internal/kubernetes"
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
)

type Global struct {
	Kubernetes
	Client  kubernetes.KubeClient
	Dialect Databaser `mapstructure:"-"`

	Job    *batchv1.Job
	JobPod v1.Pod `mapstructure:"-"`
	DbPod  v1.Pod `mapstructure:"-"`

	Host       string
	Port       uint16
	Database   string
	Username   string
	Password   string
	Quiet      bool
	RemoteGzip bool `mapstructure:"remote-gzip"`
}

package config

import (
	"github.com/clevyr/kubedb/internal/kubernetes"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
)

type Global struct {
	Kubernetes
	Client  kubernetes.KubeClient
	Dialect Database `mapstructure:"-"`

	Job    *batchv1.Job
	JobPod corev1.Pod `mapstructure:"-"`
	DBPod  corev1.Pod `mapstructure:"-"`

	Host       string
	Port       uint16
	Database   string
	Username   string
	Password   string
	Quiet      bool
	RemoteGzip bool `mapstructure:"remote-gzip"`

	Progress bool
}

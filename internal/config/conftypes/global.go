package conftypes

import (
	"gabe565.com/utils/slogx"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
)

type Log struct {
	Level  slogx.Level  `koanf:"log-level"`
	Format slogx.Format `koanf:"log-format"`
	Mask   bool         `koanf:"log-mask"`
}

type Global struct {
	Log        Log  `koanf:",squash"`
	SkipSurvey bool `koanf:"-"`

	Kubernetes  `koanf:",squash"`
	DialectName string   `koanf:"dialect"`
	Dialect     Database `koanf:"-"`

	CreateJob           bool              `koanf:"create-job"`
	CreateNetworkPolicy bool              `koanf:"create-network-policy"`
	PodName             string            `koanf:"pod"`
	Job                 *batchv1.Job      `koanf:"-"`
	JobPod              corev1.Pod        `koanf:"-"`
	JobPodLabels        map[string]string `koanf:"job-pod-labels"`
	DBPod               corev1.Pod        `koanf:"-"`

	Host       string `koanf:"-"`
	Port       uint16 `koanf:"port"`
	Database   string `koanf:"dbname"`
	Username   string `koanf:"username"`
	Password   string `koanf:"password"`
	Quiet      bool   `koanf:"quiet"`
	RemoteGzip bool   `koanf:"remote-gzip"`

	Progress            bool   `koanf:"progress"`
	HealthchecksPingURL string `koanf:"healthchecks-ping-url"`

	NamespaceColors map[string]string `koanf:"namespace-colors"`
}

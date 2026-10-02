package conftypes

import (
	"gabe565.com/utils/slogx"
	"github.com/clevyr/kubedb/internal/command"
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
	Replica             bool              `koanf:"replica"`
	Job                 *batchv1.Job      `koanf:"-"`
	JobSecret           *corev1.Secret    `koanf:"-"`
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
	Opts       string `koanf:"opts"`
	Spinner    string `koanf:"spinner"`

	// PasswordSource references the Secret or ConfigMap key that holds the discovered password.
	PasswordSource *corev1.EnvVarSource `koanf:"-"`
	// PasswordRef is a shell expression that expands to the password within JobPod.
	// When set, it is used instead of embedding the password in commands.
	PasswordRef string `koanf:"-"`

	Progress            bool   `koanf:"progress"`
	HealthchecksPingURL string `koanf:"healthchecks-ping-url"`

	NamespaceColors map[string]string `koanf:"namespace-colors"`
}

// PasswordArg returns a `key=password` shell word. If the password is available within JobPod,
// it is referenced instead of embedded so that it does not leak into the exec request.
func (g *Global) PasswordArg(key string) command.Quoter {
	if g.PasswordRef != "" {
		return command.Raw(key + "=" + g.PasswordRef)
	}
	return command.NewEnv(key, g.Password)
}

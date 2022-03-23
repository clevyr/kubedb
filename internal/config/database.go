package config

import (
	"github.com/clevyr/kubedb/internal/database/sqlformat"
	"github.com/clevyr/kubedb/internal/kubernetes"
	v1 "k8s.io/api/core/v1"
)

type Databaser interface {
	Name() string
	DefaultPort() uint16

	DatabaseEnvNames() []string
	DefaultDatabase() string

	UserEnvNames() []string
	DefaultUser() string

	DropDatabaseQuery(database string) string
	AnalyzeQuery() string
	PodLabels() []kubernetes.LabelQueryable
	FilterPods(client kubernetes.KubeClient, pods []v1.Pod) ([]v1.Pod, error)
	PasswordEnvNames() []string

	ExecCommand(conf Exec) []string
	DumpCommand(conf Dump) []string
	RestoreCommand(conf Restore, inputFormat sqlformat.Format) []string
}

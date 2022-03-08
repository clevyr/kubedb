package database

import (
	"github.com/clevyr/kubedb/internal/config"
	"github.com/clevyr/kubedb/internal/database/sqlformat"
	"github.com/clevyr/kubedb/internal/kubernetes"
	v1core "k8s.io/api/core/v1"
)

type Databaser interface {
	Name() string
	DefaultDatabase() string
	DefaultUser() string
	DropDatabaseQuery(database string) string
	AnalyzeQuery() string

	GetPod(client kubernetes.KubeClient) (v1core.Pod, error)
	GetSecret(client kubernetes.KubeClient) (string, error)

	ExecCommand(conf config.Exec) []string
	DumpCommand(conf config.Dump) []string
	RestoreCommand(conf config.Restore, inputFormat sqlformat.Format) []string
}

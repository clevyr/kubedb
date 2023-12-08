package config

import (
	"context"

	"github.com/clevyr/kubedb/internal/command"
	"github.com/clevyr/kubedb/internal/database/sqlformat"
	"github.com/clevyr/kubedb/internal/kubernetes"
	v1 "k8s.io/api/core/v1"
)

type Database interface {
	Name() string
	DefaultPort() uint16

	PortEnvNames() []string
	DatabaseEnvNames() []string
	ListDatabasesQuery() string
	ListTablesQuery() string

	UserEnvNames() []string
	DefaultUser() string

	DropDatabaseQuery(database string) string
	AnalyzeQuery() string
	PodLabels() []kubernetes.LabelQueryable
	FilterPods(ctx context.Context, client kubernetes.KubeClient, pods []v1.Pod) ([]v1.Pod, error)
	PasswordEnvNames(conf Global) []string

	ExecCommand(conf Exec) *command.Builder
	DumpCommand(conf Dump) *command.Builder
	RestoreCommand(conf Restore, inputFormat sqlformat.Format) *command.Builder

	Formats() map[sqlformat.Format]string
	FormatFromFilename(filename string) sqlformat.Format
	DumpExtension(format sqlformat.Format) string
}

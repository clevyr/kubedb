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
	PodLabels() []kubernetes.LabelQueryable
}

type DatabaseAliases interface {
	Aliases() []string
}

type DatabaseDump interface {
	Database
	DumpCommand(conf Dump) *command.Builder
	DatabaseFile
}

type DatabaseExec interface {
	Database
	ExecCommand(conf Exec) *command.Builder
}

type DatabaseRestore interface {
	Database
	RestoreCommand(conf Restore, inputFormat sqlformat.Format) *command.Builder
	DatabaseFile
}

type DatabaseFilter interface {
	FilterPods(ctx context.Context, client kubernetes.KubeClient, pods []v1.Pod) ([]v1.Pod, error)
}

type DatabaseFile interface {
	Formats() map[sqlformat.Format]string
	FormatFromFilename(filename string) sqlformat.Format
	DumpExtension(format sqlformat.Format) string
}

type DatabaseUsername interface {
	UserEnvNames() kubernetes.ConfigFinders
	DefaultUser() string
}

type DatabasePort interface {
	DefaultPort() uint16
	PortEnvNames() kubernetes.ConfigFinders
}

type DatabasePassword interface {
	PasswordEnvNames(conf Global) kubernetes.ConfigFinders
}

type DatabaseDb interface {
	DatabaseEnvNames() kubernetes.ConfigFinders
}

type DatabaseDbList interface {
	ListDatabasesQuery() string
}

type DatabaseDbDrop interface {
	DropDatabaseQuery(database string) string
}

type DatabaseTables interface {
	ListTablesQuery() string
}

type DatabaseAnalyze interface {
	AnalyzeQuery() string
}

type DatabaseDisableJob interface {
	DisableJob() bool
}

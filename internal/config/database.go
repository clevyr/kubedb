package config

import (
	"context"

	"github.com/clevyr/kubedb/internal/command"
	"github.com/clevyr/kubedb/internal/database/sqlformat"
	"github.com/clevyr/kubedb/internal/kubernetes"
	"github.com/clevyr/kubedb/internal/kubernetes/filter"
	v1 "k8s.io/api/core/v1"
)

type Database interface {
	Name() string
	PodFilters() filter.Filter
}

type DatabaseAliases interface {
	Aliases() []string
}

type DatabasePriority interface {
	Priority() uint8
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
}

type DatabaseUsername interface {
	UserEnvNames() kubernetes.ConfigLookups
	DefaultUser() string
}

type DatabasePort interface {
	DefaultPort() uint16
	PortEnvNames() kubernetes.ConfigLookups
}

type DatabasePassword interface {
	PasswordEnvNames(conf Global) kubernetes.ConfigLookups
}

type DatabaseDB interface {
	DatabaseEnvNames() kubernetes.ConfigLookups
}

type DatabaseDBList interface {
	ListDatabasesQuery() string
}

type DatabaseDBDrop interface {
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

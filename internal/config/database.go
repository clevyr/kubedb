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
	PrettyName() string
	PodFilters() filter.Filter
}

type DBAliaser interface {
	Aliases() []string
}

type DBOrderer interface {
	Priority() uint8
}

type DBDumper interface {
	Database
	DumpCommand(conf Dump) *command.Builder
	DBFiler
}

type DBExecer interface {
	Database
	ExecCommand(conf Exec) *command.Builder
}

type DBRestorer interface {
	Database
	RestoreCommand(conf Restore, inputFormat sqlformat.Format) *command.Builder
	DBFiler
}

type DBFilterer interface {
	FilterPods(ctx context.Context, client kubernetes.KubeClient, pods []v1.Pod) ([]v1.Pod, error)
}

type DBFiler interface {
	Formats() map[sqlformat.Format]string
}

type DBHasUser interface {
	UserEnvs() kubernetes.ConfigLookups
	UserDefault() string
}

type DBHasPort interface {
	PortEnvs() kubernetes.ConfigLookups
	PortDefault() uint16
}

type DBHasPassword interface {
	PasswordEnvs(conf Global) kubernetes.ConfigLookups
}

type DBHasDatabase interface {
	DatabaseEnvs() kubernetes.ConfigLookups
}

type DBDatabaseLister interface {
	DatabaseListQuery() string
}

type DBDatabaseDropper interface {
	DatabaseDropQuery(database string) string
}

type DBTableLister interface {
	TableListQuery() string
}

type DBAnalyzer interface {
	AnalyzeQuery() string
}

type DBCanDisableJob interface {
	DisableJob() bool
}

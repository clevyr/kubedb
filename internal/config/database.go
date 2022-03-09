package config

import (
	"github.com/clevyr/kubedb/internal/database/sqlformat"
	"github.com/clevyr/kubedb/internal/kubernetes"
)

type Databaser interface {
	Name() string

	DatabaseEnvNames() []string
	DefaultDatabase() string

	UserEnvNames() []string
	DefaultUser() string

	DropDatabaseQuery(database string) string
	AnalyzeQuery() string
	PodLabels() []kubernetes.LabelQueryable
	PasswordEnvNames() []string

	ExecCommand(conf Exec) []string
	DumpCommand(conf Dump) []string
	RestoreCommand(conf Restore, inputFormat sqlformat.Format) []string
}

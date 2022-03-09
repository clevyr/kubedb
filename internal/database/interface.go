package database

import (
	"github.com/clevyr/kubedb/internal/config"
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

	ExecCommand(conf config.Exec) []string
	DumpCommand(conf config.Dump) []string
	RestoreCommand(conf config.Restore, inputFormat sqlformat.Format) []string
}

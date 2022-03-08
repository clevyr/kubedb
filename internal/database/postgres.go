package database

import (
	"github.com/clevyr/kubedb/internal/config"
	"github.com/clevyr/kubedb/internal/database/sqlformat"
	"github.com/clevyr/kubedb/internal/kubernetes"
)

type Postgres struct{}

func (Postgres) Name() string {
	return "postgres"
}

func (Postgres) DefaultDatabase() string {
	return "db"
}

func (Postgres) DefaultUser() string {
	return "postgres"
}

func (Postgres) DropDatabaseQuery(database string) string {
	return "drop schema public cascade; create schema public;"
}

func (Postgres) AnalyzeQuery() string {
	return "analyze;"
}

func (Postgres) PodLabels() []kubernetes.LabelQueryable {
	return []kubernetes.LabelQueryable{
		kubernetes.LabelQuery{
			Name:  "app",
			Value: "postgresql",
		},
		kubernetes.LabelQueryAnd{
			{"app.kubernetes.io/name", "postgresql"},
			{"app.kubernetes.io/component", "primary"},
		},
		kubernetes.LabelQueryAnd{
			{"app.kubernetes.io/name", "postgresql-ha"},
			{"app.kubernetes.io/component", "postgresql"},
		},
	}
}

func (Postgres) PasswordEnvNames() []string {
	return []string{"POSTGRES_PASSWORD", "PGPOOL_POSTGRES_PASSWORD"}
}

func (Postgres) ExecCommand(conf config.Exec) []string {
	return []string{"PGPASSWORD=" + conf.Password, "psql", "--host=127.0.0.1", "--username=" + conf.Username, "--dbname=" + conf.Database}
}

func (Postgres) DumpCommand(conf config.Dump) []string {
	cmd := []string{"PGPASSWORD=" + conf.Password, "pg_dump", "--host=127.0.0.1", "--username=" + conf.Username, "--dbname=" + conf.Database}
	if conf.Clean {
		cmd = append(cmd, "--clean")
	}
	if conf.NoOwner {
		cmd = append(cmd, "--no-owner")
	}
	if conf.IfExists {
		cmd = append(cmd, "--if-exists")
	}
	for _, table := range conf.Tables {
		cmd = append(cmd, "--table='"+table+"'")
	}
	for _, table := range conf.ExcludeTable {
		cmd = append(cmd, "--exclude-table='"+table+"'")
	}
	for _, table := range conf.ExcludeTableData {
		cmd = append(cmd, "--exclude-table-data='"+table+"'")
	}
	if conf.OutputFormat == sqlformat.Custom {
		cmd = append(cmd, "--format=c")
	}
	return cmd
}

func (Postgres) RestoreCommand(conf config.Restore, inputFormat sqlformat.Format) []string {
	cmd := []string{"PGPASSWORD=" + conf.Password}
	switch inputFormat {
	case sqlformat.Gzip, sqlformat.Plain:
		cmd = append(cmd, "psql", "--quiet")
	case sqlformat.Custom:
		cmd = append(cmd, "pg_restore", "--format=custom", "--verbose")
		if conf.NoOwner {
			cmd = append(cmd, "--no-owner")
		}
	}
	cmd = append(cmd, "--host=127.0.0.1", "--username="+conf.Username, "--dbname="+conf.Database)
	if conf.SingleTransaction {
		cmd = append(cmd, "--single-transaction")
	}
	return cmd
}

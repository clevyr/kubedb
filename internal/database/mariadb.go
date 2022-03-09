package database

import (
	"github.com/clevyr/kubedb/internal/config"
	"github.com/clevyr/kubedb/internal/database/sqlformat"
	"github.com/clevyr/kubedb/internal/kubernetes"
)

type MariaDB struct{}

func (MariaDB) Name() string {
	return "mariadb"
}
func (MariaDB) DatabaseEnvNames() []string {
	return []string{"MARIADB_DATABASE"}
}

func (MariaDB) DefaultDatabase() string {
	return "db"
}

func (MariaDB) UserEnvNames() []string {
	return []string{"MARIADB_USER"}
}

func (MariaDB) DefaultUser() string {
	return "mariadb"
}

func (MariaDB) DropDatabaseQuery(database string) string {
	return "set FOREIGN_KEY_CHECKS=0; create or replace database " + database + "; set FOREIGN_KEY_CHECKS=1;"
}

func (MariaDB) AnalyzeQuery() string {
	return ";"
}

func (MariaDB) PodLabels() []kubernetes.LabelQueryable {
	return []kubernetes.LabelQueryable{
		kubernetes.LabelQuery{
			Name:  "app",
			Value: "mariadb",
		},
		kubernetes.LabelQueryAnd{
			{"app.kubernetes.io/name", "mariadb"},
			{"app.kubernetes.io/component", "primary"},
		},
	}
}

func (MariaDB) PasswordEnvNames() []string {
	return []string{"MARIADB_PASSWORD"}
}

func (MariaDB) ExecCommand(conf config.Exec) []string {
	return []string{"MYSQL_PWD=" + conf.Password, "mysql", "--host=127.0.0.1", "--user=" + conf.Username, "--database=" + conf.Database}
}

func (MariaDB) DumpCommand(conf config.Dump) []string {
	cmd := []string{"MYSQL_PWD=" + conf.Password, "mysqldump", "--host=127.0.0.1", "--user=" + conf.Username, conf.Database}
	if conf.Clean {
		cmd = append(cmd, "--add-drop-table")
	}
	for _, table := range conf.Tables {
		cmd = append(cmd, table)
	}
	for _, table := range conf.ExcludeTable {
		cmd = append(cmd, "--ignore-table='"+table+"'")
	}
	return cmd
}

func (MariaDB) RestoreCommand(conf config.Restore, inputFormat sqlformat.Format) []string {
	return []string{"MYSQL_PWD=" + conf.Password, "mysql", "--host=127.0.0.1", "--user=" + conf.Username, conf.Database}
}

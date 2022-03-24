package grammar

import (
	"github.com/clevyr/kubedb/internal/command"
	"github.com/clevyr/kubedb/internal/config"
	"github.com/clevyr/kubedb/internal/database/sqlformat"
	"github.com/clevyr/kubedb/internal/kubernetes"
	v1 "k8s.io/api/core/v1"
)

type MariaDB struct{}

func (MariaDB) Name() string {
	return "mariadb"
}

func (MariaDB) DefaultPort() uint16 {
	return 3306
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
	return ""
}

func (MariaDB) PodLabels() []kubernetes.LabelQueryable {
	return []kubernetes.LabelQueryable{
		kubernetes.LabelQuery{Name: "app", Value: "mariadb"},
		kubernetes.LabelQueryAnd{
			{Name: "app.kubernetes.io/name", Value: "mariadb"},
			{Name: "app.kubernetes.io/component", Value: "primary"},
		},
	}
}

func (MariaDB) FilterPods(client kubernetes.KubeClient, pods []v1.Pod) ([]v1.Pod, error) {
	return pods, nil
}

func (MariaDB) PasswordEnvNames() []string {
	return []string{"MARIADB_PASSWORD"}
}

func (MariaDB) ExecCommand(conf config.Exec) *command.Builder {
	return command.NewBuilder(
		command.NewEnv("MYSQL_PWD", conf.Password),
		"mysql", "--host=127.0.0.1", "--user="+conf.Username, "--database="+conf.Database,
	)
}

func (MariaDB) DumpCommand(conf config.Dump) *command.Builder {
	cmd := command.NewBuilder(
		command.NewEnv("MYSQL_PWD", conf.Password),
		"mysqldump", "--host=127.0.0.1", "--user="+conf.Username, conf.Database,
	)
	if conf.Clean {
		cmd.Push("--add-drop-table")
	}
	for _, table := range conf.Tables {
		cmd.Push(table)
	}
	for _, table := range conf.ExcludeTable {
		cmd.Push("--ignore-table=" + table)
	}
	return cmd
}

func (MariaDB) RestoreCommand(conf config.Restore, inputFormat sqlformat.Format) *command.Builder {
	return command.NewBuilder(
		command.NewEnv("MYSQL_PWD", conf.Password),
		"mysql", "--host=127.0.0.1", "--user="+conf.Username, conf.Database,
	)
}

package dialect

import (
	"context"
	"strconv"
	"strings"

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

func (MariaDB) PortEnvNames() []string {
	return []string{"MARIADB_PORT_NUMBER", "MYSQL_PORT_NUMBER"}
}

func (MariaDB) DefaultPort() uint16 {
	return 3306
}

func (MariaDB) DatabaseEnvNames() []string {
	return []string{"MARIADB_DATABASE", "MYSQL_DATABASE"}
}

func (MariaDB) ListDatabasesQuery() string {
	return "show databases"
}

func (MariaDB) ListTablesQuery() string {
	return "show tables"
}

func (MariaDB) UserEnvNames() []string {
	return []string{"MARIADB_USER", "MYSQL_USER"}
}

func (MariaDB) DefaultUser() string {
	return "root"
}

func (MariaDB) DropDatabaseQuery(database string) string {
	return "set FOREIGN_KEY_CHECKS=0; create or replace database " + database + "; set FOREIGN_KEY_CHECKS=1; use " + database + ";"
}

func (MariaDB) AnalyzeQuery() string {
	return ""
}

func (MariaDB) PodLabels() []kubernetes.LabelQueryable {
	return []kubernetes.LabelQueryable{
		kubernetes.LabelQueryAnd{
			{Name: "app.kubernetes.io/name", Value: "mariadb"},
			{Name: "app.kubernetes.io/component", Value: "primary"},
		},
		kubernetes.LabelQueryAnd{
			{Name: "app.kubernetes.io/name", Value: "mariadb-galera"},
		},
		kubernetes.LabelQueryAnd{
			{Name: "app.kubernetes.io/name", Value: "mysql"},
			{Name: "app.kubernetes.io/component", Value: "primary"},
		},
		kubernetes.LabelQuery{Name: "app", Value: "mariadb"},
		kubernetes.LabelQuery{Name: "app", Value: "mysql"},
	}
}

func (MariaDB) FilterPods(ctx context.Context, client kubernetes.KubeClient, pods []v1.Pod) ([]v1.Pod, error) {
	return pods, nil
}

func (db MariaDB) PasswordEnvNames(c config.Global) []string {
	if c.Username == db.DefaultUser() {
		return []string{"MARIADB_ROOT_PASSWORD", "MYSQL_ROOT_PASSWORD"}
	}
	return []string{"MARIADB_PASSWORD", "MYSQL_PASSWORD"}
}

func (MariaDB) ExecCommand(conf config.Exec) *command.Builder {
	cmd := command.NewBuilder(
		command.NewEnv("MYSQL_PWD", conf.Password),
		"exec", "mysql", "--host="+conf.Host, "--user="+conf.Username,
	)
	if conf.Port != 0 {
		cmd.Push("--port=" + strconv.Itoa(int(conf.Port)))
	}
	if conf.Database != "" {
		cmd.Push("--database=" + conf.Database)
	}
	if conf.DisableHeaders {
		cmd.Push("--skip-column-names")
	}
	if conf.Command != "" {
		cmd.Push("--execute=" + conf.Command)
	}
	return cmd
}

func (MariaDB) DumpCommand(conf config.Dump) *command.Builder {
	cmd := command.NewBuilder(
		command.NewEnv("MYSQL_PWD", conf.Password),
		"mysqldump", "--host="+conf.Host, "--user="+conf.Username, conf.Database,
	)
	if conf.Port != 0 {
		cmd.Push("--port=" + strconv.Itoa(int(conf.Port)))
	}
	if conf.Clean {
		cmd.Push("--add-drop-table")
	}
	for _, table := range conf.Tables {
		cmd.Push(table)
	}
	for _, table := range conf.ExcludeTable {
		cmd.Push("--ignore-table=" + table)
	}
	if !conf.Quiet {
		cmd.Push("--verbose")
	}
	return cmd
}

func (MariaDB) RestoreCommand(conf config.Restore, inputFormat sqlformat.Format) *command.Builder {
	cmd := command.NewBuilder(
		command.NewEnv("MYSQL_PWD", conf.Password),
		"mysql", "--host="+conf.Host, "--user="+conf.Username, "--database="+conf.Database,
	)
	if conf.Port != 0 {
		cmd.Push("--port=" + strconv.Itoa(int(conf.Port)))
	}
	return cmd
}

func (MariaDB) Formats() map[sqlformat.Format]string {
	return map[sqlformat.Format]string{
		sqlformat.Plain: ".sql",
		sqlformat.Gzip:  ".sql.gz",
	}
}

func (db MariaDB) FormatFromFilename(filename string) sqlformat.Format {
	for format, ext := range db.Formats() {
		if strings.HasSuffix(filename, ext) {
			return format
		}
	}
	return sqlformat.Unknown
}

func (db MariaDB) DumpExtension(format sqlformat.Format) string {
	ext, ok := db.Formats()[format]
	if ok {
		return ext
	}
	return ""
}

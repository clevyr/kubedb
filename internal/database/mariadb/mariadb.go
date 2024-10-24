package mariadb

import (
	"strconv"
	"strings"

	"github.com/clevyr/kubedb/internal/command"
	"github.com/clevyr/kubedb/internal/config"
	"github.com/clevyr/kubedb/internal/database/sqlformat"
	"github.com/clevyr/kubedb/internal/kubernetes"
	"github.com/clevyr/kubedb/internal/kubernetes/filter"
)

var (
	_ config.DBAliaser         = MariaDB{}
	_ config.DBOrderer         = MariaDB{}
	_ config.DBDumper          = MariaDB{}
	_ config.DBExecer          = MariaDB{}
	_ config.DBRestorer        = MariaDB{}
	_ config.DBHasUser         = MariaDB{}
	_ config.DBHasPort         = MariaDB{}
	_ config.DBHasPassword     = MariaDB{}
	_ config.DBHasDatabase     = MariaDB{}
	_ config.DBDatabaseLister  = MariaDB{}
	_ config.DBDatabaseDropper = MariaDB{}
	_ config.DBTableLister     = MariaDB{}
)

type MariaDB struct{}

func (MariaDB) Name() string { return "mariadb" }

func (MariaDB) PrettyName() string { return "MariaDB" }

func (MariaDB) Aliases() []string { return []string{"maria", "mysql"} }

func (MariaDB) Priority() uint8 { return 255 }

func (MariaDB) PortEnvs(_ config.Global) kubernetes.ConfigLookups {
	return kubernetes.ConfigLookups{kubernetes.LookupEnv{"MARIADB_PORT_NUMBER", "MYSQL_PORT_NUMBER"}}
}

func (MariaDB) PortDefault() uint16 {
	return 3306
}

func (MariaDB) DatabaseEnvs(_ config.Global) kubernetes.ConfigLookups {
	return kubernetes.ConfigLookups{kubernetes.LookupEnv{"MARIADB_DATABASE", "MYSQL_DATABASE"}}
}

func (MariaDB) DatabaseListQuery() string { return "show databases" }

func (MariaDB) TableListQuery() string { return "show tables" }

func (MariaDB) UserEnvs(_ config.Global) kubernetes.ConfigLookups {
	return kubernetes.ConfigLookups{kubernetes.LookupEnv{"MARIADB_USER", "MYSQL_USER"}}
}

func (MariaDB) UserDefault() string { return "root" }

func (db MariaDB) DatabaseDropQuery(database string) string {
	database = db.quoteIdentifier(database)
	return "set FOREIGN_KEY_CHECKS=0; create or replace database " + database + "; set FOREIGN_KEY_CHECKS=1; use " + database + ";"
}

func (MariaDB) PodFilters() filter.Filter {
	return filter.Or{
		filter.And{
			filter.Label{Name: "app.kubernetes.io/name", Value: "mariadb"},
			filter.Label{Name: "app.kubernetes.io/component", Value: "primary"},
		},
		filter.And{
			filter.Label{Name: "app.kubernetes.io/name", Value: "mariadb-galera"},
		},
		filter.And{
			filter.Label{Name: "app.kubernetes.io/name", Value: "mysql"},
			filter.Label{Name: "app.kubernetes.io/component", Value: "primary"},
		},
		filter.Label{Name: "app", Value: "mariadb"},
		filter.Label{Name: "app", Value: "mysql"},
	}
}

func (db MariaDB) PasswordEnvs(c config.Global) kubernetes.ConfigLookups {
	if c.Username == db.UserDefault() {
		return kubernetes.ConfigLookups{
			kubernetes.LookupEnv{"MARIADB_ROOT_PASSWORD", "MYSQL_ROOT_PASSWORD"},
		}
	}
	return kubernetes.ConfigLookups{
		kubernetes.LookupEnv{"MARIADB_PASSWORD", "MYSQL_PASSWORD"},
	}
}

func (MariaDB) ExecCommand(conf config.Exec) *command.Builder {
	cmd := command.NewBuilder(
		command.NewEnv("MYSQL_PWD", conf.Password),
		"exec", command.Raw(`"$(which mariadb || which mysql)"`), "--host="+conf.Host, "--user="+conf.Username,
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
		command.Raw(`"$(which mariadb-dump || which mysqldump)"`), "--host="+conf.Host, "--user="+conf.Username, conf.Database,
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

func (MariaDB) RestoreCommand(conf config.Restore, _ sqlformat.Format) *command.Builder {
	cmd := command.NewBuilder(
		command.NewEnv("MYSQL_PWD", conf.Password),
		command.Raw(`"$(which mariadb || which mysql)"`), "--host="+conf.Host, "--user="+conf.Username, "--database="+conf.Database,
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

func (db MariaDB) quoteIdentifier(param string) string {
	param = strings.ReplaceAll(param, "`", "``")
	param = "`" + param + "`"
	return param
}

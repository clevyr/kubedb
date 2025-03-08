package mariadb

import (
	"strconv"
	"strings"

	"github.com/clevyr/kubedb/internal/command"
	"github.com/clevyr/kubedb/internal/config/conftypes"
	"github.com/clevyr/kubedb/internal/database/sqlformat"
	"github.com/clevyr/kubedb/internal/kubernetes"
	"github.com/clevyr/kubedb/internal/kubernetes/filter"
)

var (
	_ conftypes.DBAliaser         = MariaDB{}
	_ conftypes.DBOrderer         = MariaDB{}
	_ conftypes.DBDumper          = MariaDB{}
	_ conftypes.DBExecer          = MariaDB{}
	_ conftypes.DBRestorer        = MariaDB{}
	_ conftypes.DBHasUser         = MariaDB{}
	_ conftypes.DBHasPort         = MariaDB{}
	_ conftypes.DBHasPassword     = MariaDB{}
	_ conftypes.DBHasDatabase     = MariaDB{}
	_ conftypes.DBDatabaseLister  = MariaDB{}
	_ conftypes.DBDatabaseDropper = MariaDB{}
	_ conftypes.DBTableLister     = MariaDB{}
)

type MariaDB struct{}

func (MariaDB) Name() string { return "mariadb" }

func (MariaDB) PrettyName() string { return "MariaDB" }

func (MariaDB) Aliases() []string { return []string{"maria", "mysql"} }

func (MariaDB) Priority() uint8 { return 255 }

func (MariaDB) PortEnvs(_ *conftypes.Global) kubernetes.ConfigLookups {
	return kubernetes.ConfigLookups{kubernetes.LookupEnv{"MARIADB_PORT_NUMBER", "MYSQL_PORT_NUMBER"}}
}

func (MariaDB) PortDefault() uint16 {
	return 3306
}

func (MariaDB) DatabaseEnvs(_ *conftypes.Global) kubernetes.ConfigLookups {
	return kubernetes.ConfigLookups{kubernetes.LookupEnv{"MARIADB_DATABASE", "MYSQL_DATABASE"}}
}

func (MariaDB) DatabaseListQuery() string { return "show databases" }

func (MariaDB) TableListQuery() string { return "show tables" }

func (MariaDB) UserEnvs(_ *conftypes.Global) kubernetes.ConfigLookups {
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

func (db MariaDB) PasswordEnvs(c *conftypes.Global) kubernetes.ConfigLookups {
	if c.Username == db.UserDefault() {
		return kubernetes.ConfigLookups{
			kubernetes.LookupEnv{"MARIADB_ROOT_PASSWORD", "MYSQL_ROOT_PASSWORD"},
			kubernetes.LookupSecretVolume{Name: "mariadb-credentials", Key: "mariadb-root-password"},
		}
	}
	return kubernetes.ConfigLookups{
		kubernetes.LookupEnv{"MARIADB_PASSWORD", "MYSQL_PASSWORD"},
		kubernetes.LookupSecretVolume{Name: "mariadb-credentials", Key: "mariadb-password"},
	}
}

func (MariaDB) ExecCommand(conf *conftypes.Exec) *command.Builder {
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

func (MariaDB) DumpCommand(conf *conftypes.Dump) *command.Builder {
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
	for _, table := range conf.Table {
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

func (MariaDB) RestoreCommand(conf *conftypes.Restore, _ sqlformat.Format) *command.Builder {
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

//nolint:goconst
package mongodb

import (
	"strconv"

	"github.com/clevyr/kubedb/internal/command"
	"github.com/clevyr/kubedb/internal/config"
	"github.com/clevyr/kubedb/internal/database/sqlformat"
	"github.com/clevyr/kubedb/internal/kubernetes"
	"github.com/clevyr/kubedb/internal/kubernetes/filter"
)

var (
	_ config.DatabaseAliases  = MongoDB{}
	_ config.DatabaseDump     = MongoDB{}
	_ config.DatabaseExec     = MongoDB{}
	_ config.DatabaseRestore  = MongoDB{}
	_ config.DatabaseUsername = MongoDB{}
	_ config.DatabasePort     = MongoDB{}
	_ config.DatabasePassword = MongoDB{}
	_ config.DatabaseDb       = MongoDB{}
	_ config.DatabaseDbList   = MongoDB{}
	_ config.DatabaseTables   = MongoDB{}
)

type MongoDB struct{}

func (MongoDB) Name() string {
	return "mongodb"
}

func (MongoDB) Aliases() []string {
	return []string{"mongo"}
}

func (MongoDB) PortEnvNames() kubernetes.ConfigLookups {
	return kubernetes.ConfigLookups{kubernetes.LookupEnv{"MONGODB_PORT_NUMBER"}}
}

func (MongoDB) DefaultPort() uint16 {
	return 27017
}

func (MongoDB) DatabaseEnvNames() kubernetes.ConfigLookups {
	return kubernetes.ConfigLookups{kubernetes.LookupEnv{"MONGODB_EXTRA_DATABASES"}}
}

func (MongoDB) ListDatabasesQuery() string {
	return "db.getMongo().getDBNames().forEach(function(db){ print(db) })"
}

func (MongoDB) ListTablesQuery() string {
	return "db.getCollectionNames().forEach(function(collection){ print(collection) })"
}

func (MongoDB) UserEnvNames() kubernetes.ConfigLookups {
	return kubernetes.ConfigLookups{kubernetes.LookupEnv{"MONGODB_EXTRA_USERNAMES", "MONGODB_ROOT_USER"}}
}

func (MongoDB) DefaultUser() string {
	return "root"
}

func (MongoDB) PodFilters() filter.Filter {
	return filter.Or{
		filter.And{
			filter.Label{Name: "app.kubernetes.io/name", Value: "mongodb"},
			filter.Label{Name: "app.kubernetes.io/component", Value: "mongodb"},
		},
		filter.And{
			filter.Label{Name: "app.kubernetes.io/name", Value: "mongodb-sharded"},
			filter.Label{Name: "app.kubernetes.io/component", Value: "mongos"},
		},
		filter.Label{Name: "app", Value: "mongodb"},
		filter.Label{Name: "app", Value: "mongodb-replicaset"},
	}
}

func (db MongoDB) PasswordEnvNames(c config.Global) kubernetes.ConfigLookups {
	if c.Username == db.DefaultUser() {
		return kubernetes.ConfigLookups{kubernetes.LookupEnv{"MONGODB_ROOT_PASSWORD"}}
	}
	return kubernetes.ConfigLookups{kubernetes.LookupEnv{"MONGODB_EXTRA_PASSWORDS"}}
}

func (db MongoDB) AuthenticationDatabase(c config.Global) string {
	if c.Username == db.DefaultUser() {
		return "admin"
	}
	return c.Database
}

func (db MongoDB) ExecCommand(conf config.Exec) *command.Builder {
	cmd := command.NewBuilder(
		"exec", command.Raw(`"$(which mongosh || which mongo)"`),
		"--host="+conf.Host,
		"--username="+conf.Username,
		"--password="+conf.Password,
		"--authenticationDatabase="+db.AuthenticationDatabase(conf.Global),
	)
	if conf.Port != 0 {
		cmd.Push("--port=" + strconv.Itoa(int(conf.Port)))
	}
	if conf.DisableHeaders {
		cmd.Push("--quiet")
	}
	if conf.Command != "" {
		cmd.Push("--eval=" + conf.Command)
	}
	if conf.Database != "" {
		cmd.Push(conf.Database)
	}
	return cmd
}

func (db MongoDB) DumpCommand(conf config.Dump) *command.Builder {
	cmd := command.NewBuilder(
		"mongodump",
		"--archive",
		"--host="+conf.Host,
		"--username="+conf.Username,
		"--password="+conf.Password,
		"--authenticationDatabase="+db.AuthenticationDatabase(conf.Global),
	)
	if conf.Port != 0 {
		cmd.Push("--port=" + strconv.Itoa(int(conf.Port)))
	}
	if conf.Database != "" {
		cmd.Push("--db=" + conf.Database)
	}
	for _, table := range conf.Tables {
		cmd.Push("--collection=" + table)
	}
	for _, table := range conf.ExcludeTable {
		cmd.Push("--excludeCollection=" + table)
	}
	if conf.Quiet {
		cmd.Push("--quiet")
	}
	return cmd
}

func (db MongoDB) RestoreCommand(conf config.Restore, _ sqlformat.Format) *command.Builder {
	cmd := command.NewBuilder(
		"mongorestore",
		"--archive",
		"--host="+conf.Host,
		"--username="+conf.Username,
		"--password="+conf.Password,
		"--authenticationDatabase="+db.AuthenticationDatabase(conf.Global),
	)
	if conf.Port != 0 {
		cmd.Push("--port=" + strconv.Itoa(int(conf.Port)))
	}
	if conf.Database != "" {
		if conf.Clean {
			cmd.Push("--drop")
		}
		cmd.Push("--db=" + conf.Database)
	}
	if conf.Quiet {
		cmd.Push("--quiet")
	}
	return cmd
}

func (MongoDB) Formats() map[sqlformat.Format]string {
	return map[sqlformat.Format]string{
		sqlformat.Plain: ".archive",
		sqlformat.Gzip:  ".archive.gz",
	}
}

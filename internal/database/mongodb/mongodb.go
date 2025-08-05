package mongodb

import (
	"strconv"

	"github.com/clevyr/kubedb/internal/command"
	"github.com/clevyr/kubedb/internal/config/conftypes"
	"github.com/clevyr/kubedb/internal/database/sqlformat"
	"github.com/clevyr/kubedb/internal/kubernetes"
	"github.com/clevyr/kubedb/internal/kubernetes/filter"
	"k8s.io/apimachinery/pkg/selection"
)

var (
	_ conftypes.DBAliaser        = MongoDB{}
	_ conftypes.DBOrderer        = MongoDB{}
	_ conftypes.DBDumper         = MongoDB{}
	_ conftypes.DBExecer         = MongoDB{}
	_ conftypes.DBRestorer       = MongoDB{}
	_ conftypes.DBHasUser        = MongoDB{}
	_ conftypes.DBHasPort        = MongoDB{}
	_ conftypes.DBHasPassword    = MongoDB{}
	_ conftypes.DBHasDatabase    = MongoDB{}
	_ conftypes.DBDatabaseLister = MongoDB{}
	_ conftypes.DBTableLister    = MongoDB{}
)

type MongoDB struct{}

func (MongoDB) Name() string {
	return "mongodb"
}

func (MongoDB) PrettyName() string { return "MongoDB" }

func (MongoDB) Aliases() []string { return []string{"mongo"} }

func (MongoDB) Priority() uint8 { return 255 }

func (MongoDB) PortEnvs(_ *conftypes.Global) kubernetes.ConfigLookups {
	return kubernetes.ConfigLookups{kubernetes.LookupEnv{"MONGODB_PORT_NUMBER"}}
}

func (MongoDB) PortDefault() uint16 { return 27017 }

func (MongoDB) DatabaseEnvs(_ *conftypes.Global) kubernetes.ConfigLookups {
	return kubernetes.ConfigLookups{kubernetes.LookupEnv{
		"MONGO_INITDB_DATABASE",
		"MONGO_DATABASE",
		"MONGODB_DATABASE",
		"MONGODB_EXTRA_DATABASES",
	}}
}

func (MongoDB) DatabaseListQuery() string {
	return "db.getMongo().getDBNames().forEach(function(db){ print(db) })"
}

func (MongoDB) TableListQuery() string {
	return "db.getCollectionNames().forEach(function(collection){ print(collection) })"
}

func (MongoDB) UserEnvs(_ *conftypes.Global) kubernetes.ConfigLookups {
	return kubernetes.ConfigLookups{kubernetes.LookupEnv{
		"MONGODB_EXTRA_USERNAMES",
		"MONGO_INITDB_USERNAME",
		"MONGODB_USERNAME",
		"MONGO_USERNAME",
		"MONGO_INITDB_ROOT_USERNAME",
		"MONGODB_ROOT_USER",
	}}
}

func (MongoDB) UserDefault() string { return "root" }

func (MongoDB) PodFilters() filter.Filter {
	return filter.Or{
		filter.And{
			filter.Label{
				Name:     "app.kubernetes.io/name",
				Operator: selection.In,
				Values:   []string{"mongodb", "mongo"},
			},
			filter.Label{
				Name:     "app.kubernetes.io/component",
				Operator: selection.NotEquals,
				Value:    "arbiter",
			},
		},
		filter.And{
			filter.Label{Name: "app.kubernetes.io/name", Value: "mongodb-sharded"},
			filter.Label{Name: "app.kubernetes.io/component", Value: "mongos"},
		},
		filter.Label{
			Name:     "app",
			Operator: selection.In,
			Values:   []string{"mongodb", "mongodb-replicaset"},
		},
	}
}

func (db MongoDB) PasswordEnvs(c *conftypes.Global) kubernetes.ConfigLookups {
	if c.Username == db.UserDefault() {
		return kubernetes.ConfigLookups{kubernetes.LookupEnv{
			"MONGODB_ROOT_PASSWORD",
			"MONGO_INITDB_ROOT_PASSWORD",
			"MONGO_ROOT_PASSWORD",
		}}
	}
	return kubernetes.ConfigLookups{kubernetes.LookupEnv{
		"MONGODB_EXTRA_PASSWORDS",
		"MONGO_INITDB_PASSWORD",
		"MONGODB_PASSWORD",
		"MONGO_PASSWORD",
	}}
}

func (db MongoDB) AuthenticationDatabase(c *conftypes.Global) string {
	if c.Username == db.UserDefault() {
		return "admin"
	}
	return c.Database
}

func (db MongoDB) newCmd(conf *conftypes.Global, p ...any) *command.Builder {
	cmd := command.NewBuilder(p...)
	if conf.Host != "" {
		cmd.Push("--host=" + conf.Host)
	}
	if conf.Port != 0 {
		cmd.Push("--port=" + strconv.Itoa(int(conf.Port)))
	}
	if authDB := db.AuthenticationDatabase(conf); authDB != "" {
		cmd.Push("--authenticationDatabase=" + authDB)
	}
	if conf.Username != "" {
		cmd.Push("--username=" + conf.Username)
	}
	if conf.Password != "" {
		cmd.Push("--password=" + conf.Password)
	}
	return cmd
}

func (db MongoDB) ExecCommand(conf *conftypes.Exec) *command.Builder {
	cmd := db.newCmd(conf.Global, "exec", command.Raw(`"$(which mongosh || which mongo)"`))
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

func (db MongoDB) DumpCommand(conf *conftypes.Dump) *command.Builder {
	cmd := db.newCmd(conf.Global, "mongodump", "--archive")
	if conf.Database != "" {
		cmd.Push("--db=" + conf.Database)
	}
	for _, table := range conf.Table {
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

func (db MongoDB) RestoreCommand(conf *conftypes.Restore, _ sqlformat.Format) *command.Builder {
	cmd := db.newCmd(conf.Global, "mongorestore", "--archive")
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

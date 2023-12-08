package mongodb

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

type MongoDB struct{}

func (MongoDB) Name() string {
	return "mongodb"
}

func (MongoDB) PortEnvNames() []string {
	return []string{"MONGODB_PORT_NUMBER"}
}

func (MongoDB) DefaultPort() uint16 {
	return 27017
}

func (MongoDB) DatabaseEnvNames() []string {
	return []string{"MONGODB_EXTRA_DATABASES"}
}

func (MongoDB) ListDatabasesQuery() string {
	return "db.getMongo().getDBNames().forEach(function(db){ print(db) })"
}

func (MongoDB) ListTablesQuery() string {
	return "db.getCollectionNames().forEach(function(collection){ print(collection) })"
}

func (MongoDB) UserEnvNames() []string {
	return []string{"MONGODB_EXTRA_USERNAMES", "MONGODB_ROOT_USER"}
}

func (MongoDB) DefaultUser() string {
	return "root"
}

func (MongoDB) DropDatabaseQuery(database string) string {
	return ""
}

func (MongoDB) AnalyzeQuery() string {
	return ""
}

func (MongoDB) PodLabels() []kubernetes.LabelQueryable {
	return []kubernetes.LabelQueryable{
		kubernetes.LabelQueryAnd{
			{Name: "app.kubernetes.io/name", Value: "mongodb"},
			{Name: "app.kubernetes.io/component", Value: "mongodb"},
		},
		kubernetes.LabelQueryAnd{
			{Name: "app.kubernetes.io/name", Value: "mongodb-sharded"},
			{Name: "app.kubernetes.io/component", Value: "mongos"},
		},
		kubernetes.LabelQuery{Name: "app", Value: "mongodb"},
		kubernetes.LabelQuery{Name: "app", Value: "mongodb-replicaset"},
	}
}

func (MongoDB) FilterPods(ctx context.Context, client kubernetes.KubeClient, pods []v1.Pod) ([]v1.Pod, error) {
	return pods, nil
}

func (db MongoDB) PasswordEnvNames(c config.Global) []string {
	if c.Username == db.DefaultUser() {
		return []string{"MONGODB_ROOT_PASSWORD"}
	}
	return []string{"MONGODB_EXTRA_PASSWORDS"}
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

func (db MongoDB) RestoreCommand(conf config.Restore, inputFormat sqlformat.Format) *command.Builder {
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

func (db MongoDB) FormatFromFilename(filename string) sqlformat.Format {
	for format, ext := range db.Formats() {
		if strings.HasSuffix(filename, ext) {
			return format
		}
	}
	return sqlformat.Unknown
}

func (db MongoDB) DumpExtension(format sqlformat.Format) string {
	ext, ok := db.Formats()[format]
	if ok {
		return ext
	}
	return ""
}

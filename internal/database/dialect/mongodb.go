package dialect

import (
	"github.com/clevyr/kubedb/internal/command"
	"github.com/clevyr/kubedb/internal/config"
	"github.com/clevyr/kubedb/internal/database/sqlformat"
	"github.com/clevyr/kubedb/internal/kubernetes"
	v1 "k8s.io/api/core/v1"
	"strings"
)

type MongoDB struct{}

func (MongoDB) Name() string {
	return "mongodb"
}

func (MongoDB) DefaultPort() uint16 {
	return 27017
}

func (MongoDB) DatabaseEnvNames() []string {
	return []string{}
}

func (MongoDB) ListDatabasesQuery() string {
	return "db.getMongo().getDBNames().forEach(function(db){ print(db) })"
}

func (MongoDB) ListTablesQuery() string {
	return "show collections"
}

func (MongoDB) UserEnvNames() []string {
	return []string{"MONGODB_ROOT_USER"}
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
		kubernetes.LabelQuery{Name: "app", Value: "mongodb"},
	}
}

func (MongoDB) FilterPods(client kubernetes.KubeClient, pods []v1.Pod) ([]v1.Pod, error) {
	return pods, nil
}

func (MongoDB) PasswordEnvNames() []string {
	return []string{"MONGODB_ROOT_PASSWORD"}
}

func (MongoDB) ExecCommand(conf config.Exec) *command.Builder {
	cmd := command.NewBuilder(
		"mongosh",
		"--host=127.0.0.1",
		"--username="+conf.Username,
		"--password="+conf.Password,
	)
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

func (MongoDB) DumpCommand(conf config.Dump) *command.Builder {
	cmd := command.NewBuilder(
		"mongodump",
		"--archive",
		"--host=127.0.0.1",
		"--username="+conf.Username,
		"--password="+conf.Password,
	)
	if conf.Database != "" {
		cmd.Push("--db=" + conf.Database)
	}
	for _, table := range conf.Tables {
		cmd.Push("--collection=" + table)
	}
	for _, table := range conf.ExcludeTable {
		cmd.Push("--excludeCollection=" + table)
	}
	return cmd
}

func (MongoDB) RestoreCommand(conf config.Restore, inputFormat sqlformat.Format) *command.Builder {
	cmd := command.NewBuilder(
		"mongorestore",
		"--archive",
		"--host=127.0.0.1",
		"--username="+conf.Username,
		"--password="+conf.Password,
	)
	if conf.Database != "" {
		cmd.Push("--db=" + conf.Database)
		if conf.Clean {
			cmd.Push("--drop")
		}
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

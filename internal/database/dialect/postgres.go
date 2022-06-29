package dialect

import (
	"encoding/csv"
	"github.com/clevyr/kubedb/internal/command"
	"github.com/clevyr/kubedb/internal/config"
	"github.com/clevyr/kubedb/internal/database/sqlformat"
	"github.com/clevyr/kubedb/internal/kubernetes"
	log "github.com/sirupsen/logrus"
	"io"
	v1 "k8s.io/api/core/v1"
	"os"
	"strings"
)

type Postgres struct{}

func (Postgres) Name() string {
	return "postgres"
}

func (Postgres) DefaultPort() uint16 {
	return 5432
}

func (Postgres) DatabaseEnvNames() []string {
	return []string{"POSTGRES_DB"}
}

func (Postgres) ListDatabasesQuery() string {
	return "SELECT datname FROM pg_database WHERE datistemplate = false"
}

func (Postgres) ListTablesQuery() string {
	return "SELECT table_name FROM information_schema.tables WHERE table_schema='public' AND table_type='BASE TABLE'"
}

func (Postgres) UserEnvNames() []string {
	return []string{"POSTGRES_USER", "PGPOOL_POSTGRES_USERNAME"}
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
		kubernetes.LabelQuery{Name: "app", Value: "postgresql"},
		kubernetes.LabelQueryAnd{
			{Name: "app.kubernetes.io/name", Value: "postgresql"},
			{Name: "app.kubernetes.io/component", Value: "primary"},
		},
		kubernetes.LabelQueryAnd{
			{Name: "app.kubernetes.io/name", Value: "postgresql-ha"},
			{Name: "app.kubernetes.io/component", Value: "postgresql"},
		},
	}
}

func (Postgres) FilterPods(client kubernetes.KubeClient, pods []v1.Pod) ([]v1.Pod, error) {
	if len(pods) > 0 && pods[0].Labels["app.kubernetes.io/name"] == "postgresql-ha" {
		// HA chart. Need to detect primary.
		log.Info("querying for primary instance")
		cmd := command.NewBuilder(
			command.NewEnv("DISABLE_WELCOME_MESSAGE", "true"),
			"/opt/bitnami/scripts/postgresql-repmgr/entrypoint.sh",
			"repmgr", "--config-file=/opt/bitnami/repmgr/conf/repmgr.conf",
			"service", "status", "--csv",
		)

		var buf strings.Builder
		err := client.Exec(pods[0], cmd.String(), strings.NewReader(""), &buf, os.Stderr, false, nil)
		if err != nil {
			return pods, err
		}

		r := csv.NewReader(strings.NewReader(buf.String()))
		for {
			row, err := r.Read()
			if err == io.EOF {
				break
			}
			if err != nil {
				return pods, err
			}
			if row[2] == "primary" {
				for key, pod := range pods {
					if pod.Name == row[1] {
						return pods[key : key+1], nil
					}
				}
			}
		}
	}
	return pods, nil
}

func (Postgres) PasswordEnvNames() []string {
	return []string{"POSTGRES_PASSWORD", "PGPOOL_POSTGRES_PASSWORD"}
}

func (Postgres) ExecCommand(conf config.Exec) *command.Builder {
	cmd := command.NewBuilder(
		command.NewEnv("PGPASSWORD", conf.Password),
		"psql", "--host=127.0.0.1", "--username="+conf.Username,
	)
	if conf.Database != "" {
		cmd.Push("--dbname=" + conf.Database)
	}
	if conf.DisableHeaders {
		cmd.Push("--tuples-only")
	}
	if conf.Command != "" {
		cmd.Push("--command=" + conf.Command)
	}
	return cmd
}

func quoteParam(param string) string {
	param = `"` + param + `"`
	param = strings.ReplaceAll(param, "*", `"*"`)
	return param
}

func (Postgres) DumpCommand(conf config.Dump) *command.Builder {
	cmd := command.NewBuilder(
		command.NewEnv("PGPASSWORD", conf.Password),
		"pg_dump", "--host=127.0.0.1", "--username="+conf.Username, "--dbname="+conf.Database,
	)
	if conf.Clean {
		cmd.Push("--clean")
	}
	if conf.NoOwner {
		cmd.Push("--no-owner")
	}
	if conf.IfExists {
		cmd.Push("--if-exists")
	}
	for _, table := range conf.Tables {
		cmd.Push("--table=" + quoteParam(table))
	}
	for _, table := range conf.ExcludeTable {
		cmd.Push("--exclude-table=" + quoteParam(table))
	}
	for _, table := range conf.ExcludeTableData {
		cmd.Push("--exclude-table-data=" + quoteParam(table))
	}
	if conf.Format == sqlformat.Custom {
		cmd.Push("--format=c")
	}
	return cmd
}

func (Postgres) RestoreCommand(conf config.Restore, inputFormat sqlformat.Format) *command.Builder {
	cmd := command.NewBuilder(
		command.NewEnv("PGPASSWORD", conf.Password),
		command.NewEnv("PGOPTIONS", "-c client_min_messages=WARNING"),
	)
	switch inputFormat {
	case sqlformat.Gzip, sqlformat.Plain, sqlformat.Unknown:
		cmd.Push("psql", "--quiet", "--output=/dev/null")
	case sqlformat.Custom:
		cmd.Push("pg_restore", "--format=custom", "--verbose", "--clean", "--exit-on-error")
		if conf.NoOwner {
			cmd.Push("--no-owner")
		}
	}
	cmd.Push("--host=127.0.0.1", "--username="+conf.Username, "--dbname="+conf.Database)
	if conf.SingleTransaction {
		cmd.Push("--single-transaction")
	}
	return cmd
}

func (Postgres) Formats() map[sqlformat.Format]string {
	return map[sqlformat.Format]string{
		sqlformat.Plain:  ".sql",
		sqlformat.Gzip:   ".sql.gz",
		sqlformat.Custom: ".dmp",
	}
}

func (db Postgres) FormatFromFilename(filename string) sqlformat.Format {
	for format, ext := range db.Formats() {
		if strings.HasSuffix(filename, ext) {
			return format
		}
	}
	return sqlformat.Unknown
}

func (db Postgres) DumpExtension(format sqlformat.Format) string {
	ext, ok := db.Formats()[format]
	if ok {
		return ext
	}
	return ""
}

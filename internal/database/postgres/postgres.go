package postgres

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/clevyr/kubedb/internal/command"
	"github.com/clevyr/kubedb/internal/config"
	"github.com/clevyr/kubedb/internal/database/sqlformat"
	"github.com/clevyr/kubedb/internal/kubernetes"
	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/selection"

	v1 "k8s.io/api/core/v1"
)

type Postgres struct{}

func (Postgres) Name() string {
	return "postgres"
}

func (Postgres) PortEnvNames() kubernetes.ConfigFinders {
	return kubernetes.ConfigFinders{
		kubernetes.ConfigFromEnv{"POSTGRESQL_PORT_NUMBER"},
		kubernetes.ConfigFromVolumeSecret{Name: "app-secret", Key: "port"},
	}
}

func (Postgres) DefaultPort() uint16 {
	return 5432
}

func (Postgres) DatabaseEnvNames() kubernetes.ConfigFinders {
	return kubernetes.ConfigFinders{
		kubernetes.ConfigFromEnv{"POSTGRES_DATABASE", "POSTGRES_DB"},
		kubernetes.ConfigFromVolumeSecret{Name: "app-secret", Key: "dbname"},
	}
}

func (Postgres) ListDatabasesQuery() string {
	return "SELECT datname FROM pg_database WHERE datistemplate = false"
}

func (Postgres) ListTablesQuery() string {
	return "SELECT table_name FROM information_schema.tables WHERE table_schema='public' AND table_type='BASE TABLE'"
}

func (Postgres) UserEnvNames() kubernetes.ConfigFinders {
	return kubernetes.ConfigFinders{
		kubernetes.ConfigFromEnv{"POSTGRES_USER", "PGPOOL_POSTGRES_USERNAME", "PGUSER_SUPERUSER"},
		kubernetes.ConfigFromVolumeSecret{Name: "app-secret", Key: "username"},
	}
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
		kubernetes.LabelQueryAnd{
			{Name: "app.kubernetes.io/name", Value: "postgresql"},
			{Name: "app.kubernetes.io/component", Value: "primary"},
		},
		kubernetes.LabelQueryAnd{
			{Name: "app.kubernetes.io/name", Value: "postgresql-ha"},
			{Name: "app.kubernetes.io/component", Value: "postgresql"},
		},
		kubernetes.LabelQuery{Name: "cnpg.io/cluster", Operator: selection.Exists},
		kubernetes.LabelQuery{Name: "application", Value: "spilo"},
		kubernetes.LabelQuery{Name: "app", Value: "postgresql"},
	}
}

func (Postgres) FilterPods(ctx context.Context, client kubernetes.KubeClient, pods []v1.Pod) ([]v1.Pod, error) {
	if len(pods) == 0 {
		return pods, nil
	}

	var leaderName string
	if pods[0].Labels["app.kubernetes.io/name"] == "postgresql-ha" {
		// HA chart. Need to detect primary.
		log.Debug("querying Bitnami repmgr for primary instance")
		cmd := command.NewBuilder(
			command.NewEnv("DISABLE_WELCOME_MESSAGE", "true"),
			"/opt/bitnami/scripts/postgresql-repmgr/entrypoint.sh",
			"repmgr", "--config-file=/opt/bitnami/repmgr/conf/repmgr.conf",
			"service", "status", "--csv",
		)

		var buf bytes.Buffer
		var errBuf strings.Builder
		if err := client.Exec(ctx, kubernetes.ExecOptions{
			Pod:    pods[0],
			Cmd:    cmd.String(),
			Stdout: &buf,
			Stderr: &errBuf,
		}); err != nil {
			return pods, fmt.Errorf("%w: %s", err, errBuf.String())
		}

		r := csv.NewReader(&buf)
		for {
			row, err := r.Read()
			if err == io.EOF {
				break
			}
			if err != nil {
				return pods, err
			}
			if row[2] == "primary" {
				leaderName = row[1]
				break
			}
		}
	} else if _, ok := pods[0].Labels["cnpg.io/cluster"]; ok {
		log.Debug("filtering CloudNativePG Pods for Leader")

		for key, pod := range pods {
			if role, ok := pod.Labels["cnpg.io/instanceRole"]; ok {
				if role == "primary" {
					return pods[key : key+1], nil
				}
			}
		}
	} else if pods[0].Labels["application"] == "spilo" {
		log.Debug("querying Patroni for primary instance")
		cmd := command.NewBuilder("patronictl", "list", "--format=json")

		var buf bytes.Buffer
		var errBuf strings.Builder
		if err := client.Exec(ctx, kubernetes.ExecOptions{
			Pod:    pods[0],
			Cmd:    cmd.String(),
			Stdout: &buf,
			Stderr: &errBuf,
		}); err != nil {
			return pods, fmt.Errorf("%w: %s", err, errBuf.String())
		}

		var data []map[string]any
		if err := json.NewDecoder(&buf).Decode(&data); err != nil {
			return pods, err
		}

		for _, member := range data {
			if role, ok := member["Role"]; ok && role == "Leader" {
				if name, ok := member["Member"].(string); ok {
					leaderName = name
					break
				}
			}
		}
	}

	if leaderName != "" {
		for key, pod := range pods {
			if pod.Name == leaderName {
				return pods[key : key+1], nil
			}
		}
	}

	return pods, nil
}

func (db Postgres) PasswordEnvNames(c config.Global) kubernetes.ConfigFinders {
	var searchEnvs kubernetes.ConfigFromEnv
	if c.Username == db.DefaultUser() {
		searchEnvs = append(searchEnvs, "POSTGRES_POSTGRES_PASSWORD")
	}
	searchEnvs = append(searchEnvs, "POSTGRES_PASSWORD", "PGPOOL_POSTGRES_PASSWORD", "PGPASSWORD_SUPERUSER")
	return kubernetes.ConfigFinders{
		searchEnvs,
		kubernetes.ConfigFromVolumeSecret{Name: "app-secret", Key: "password"},
	}
}

func (Postgres) ExecCommand(conf config.Exec) *command.Builder {
	cmd := command.NewBuilder(
		command.NewEnv("PGPASSWORD", conf.Password),
		"exec", "psql", "--host="+conf.Host, "--username="+conf.Username,
	)
	if conf.Port != 0 {
		cmd.Push("--port=" + strconv.Itoa(int(conf.Port)))
	}
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

func (Postgres) quoteParam(param string) string {
	param = `"` + param + `"`
	param = strings.ReplaceAll(param, "*", `"*"`)
	return param
}

func (db Postgres) DumpCommand(conf config.Dump) *command.Builder {
	cmd := command.NewBuilder(
		command.NewEnv("PGPASSWORD", conf.Password),
		"pg_dump", "--host="+conf.Host, "--username="+conf.Username,
	)
	if conf.Port != 0 {
		cmd.Push("--port=" + strconv.Itoa(int(conf.Port)))
	}
	if conf.Database != "" {
		cmd.Push("--dbname=" + conf.Database)
	}
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
		cmd.Push("--table=" + db.quoteParam(table))
	}
	for _, table := range conf.ExcludeTable {
		cmd.Push("--exclude-table=" + db.quoteParam(table))
	}
	for _, table := range conf.ExcludeTableData {
		cmd.Push("--exclude-table-data=" + db.quoteParam(table))
	}
	if conf.Format == sqlformat.Custom {
		cmd.Push("--format=c")
	}
	if !conf.Quiet {
		cmd.Push("--verbose")
	}
	return cmd
}

func (Postgres) RestoreCommand(conf config.Restore, inputFormat sqlformat.Format) *command.Builder {
	cmd := command.NewBuilder(
		command.NewEnv("PGPASSWORD", conf.Password),
	)
	if conf.Quiet {
		cmd.Push(command.NewEnv("PGOPTIONS", "-c client_min_messages=WARNING"))
	}
	switch inputFormat {
	case sqlformat.Gzip, sqlformat.Plain, sqlformat.Unknown:
		cmd.Push("psql")
		if conf.Quiet {
			cmd.Push("--quiet", "--output=/dev/null")
		}
	case sqlformat.Custom:
		cmd.Push("pg_restore", "--format=custom", "--clean", "--exit-on-error")
		if conf.NoOwner {
			cmd.Push("--no-owner")
		}
		if !conf.Quiet {
			cmd.Push("--verbose")
		}
	}
	cmd.Push("--host="+conf.Host, "--username="+conf.Username, "--dbname="+conf.Database)
	if conf.Port != 0 {
		cmd.Push("--port=" + strconv.Itoa(int(conf.Port)))
	}
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

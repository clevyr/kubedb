package postgres

import (
	"bytes"
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"strconv"
	"strings"

	"github.com/clevyr/kubedb/internal/command"
	"github.com/clevyr/kubedb/internal/config/conftypes"
	"github.com/clevyr/kubedb/internal/database/sqlformat"
	"github.com/clevyr/kubedb/internal/kubernetes"
	"github.com/clevyr/kubedb/internal/kubernetes/filter"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/selection"
)

var (
	_ conftypes.DBAliaser         = Postgres{}
	_ conftypes.DBOrderer         = Postgres{}
	_ conftypes.DBDumper          = Postgres{}
	_ conftypes.DBExecer          = Postgres{}
	_ conftypes.DBRestorer        = Postgres{}
	_ conftypes.DBFilterer        = Postgres{}
	_ conftypes.DBHasUser         = Postgres{}
	_ conftypes.DBHasPort         = Postgres{}
	_ conftypes.DBHasPassword     = Postgres{}
	_ conftypes.DBHasDatabase     = Postgres{}
	_ conftypes.DBDatabaseLister  = Postgres{}
	_ conftypes.DBDatabaseDropper = Postgres{}
	_ conftypes.DBTableLister     = Postgres{}
	_ conftypes.DBAnalyzer        = Postgres{}
)

type Postgres struct{}

func (Postgres) Name() string { return "postgres" }

func (Postgres) PrettyName() string { return "Postgres" }

func (Postgres) Aliases() []string { return []string{"postgresql", "psql", "pg"} }

func (Postgres) Priority() uint8 { return 255 }

func (db Postgres) PortEnvs(conf *conftypes.Global) kubernetes.ConfigLookups {
	if secret := db.cnpgSecretName(conf); secret != "" {
		return kubernetes.ConfigLookups{kubernetes.LookupNamedSecret{
			Name: secret,
			Key:  "port",
		}}
	}

	return kubernetes.ConfigLookups{
		kubernetes.LookupEnv{"POSTGRESQL_PORT_NUMBER"},
	}
}

func (Postgres) PortDefault() uint16 { return 5432 }

func (db Postgres) DatabaseEnvs(conf *conftypes.Global) kubernetes.ConfigLookups {
	if secret := db.cnpgSecretName(conf); secret != "" {
		cluster, _ := db.cnpgClusterName(conf)
		return kubernetes.ConfigLookups{kubernetes.LookupNamedSecret{
			Name: cluster + "-app", // Always use the app secret since superuser dbname is `*`
			Key:  "dbname",
		}}
	}

	return kubernetes.ConfigLookups{
		kubernetes.LookupEnv{"POSTGRES_DATABASE", "POSTGRES_DB"},
		kubernetes.LookupDefault("postgres"),
	}
}

func (Postgres) DatabaseListQuery() string {
	return "SELECT datname FROM pg_database WHERE datistemplate = false"
}

func (Postgres) TableListQuery() string {
	return "SELECT table_name FROM information_schema.tables WHERE table_schema='public' AND table_type='BASE TABLE'"
}

func (db Postgres) UserEnvs(conf *conftypes.Global) kubernetes.ConfigLookups {
	if secret := db.cnpgSecretName(conf); secret != "" {
		return kubernetes.ConfigLookups{kubernetes.LookupNamedSecret{
			Name: secret,
			Key:  "username",
		}}
	}

	return kubernetes.ConfigLookups{
		kubernetes.LookupEnv{"POSTGRES_USER", "PGPOOL_POSTGRES_USERNAME", "PGUSER_SUPERUSER"},
	}
}

func (Postgres) UserDefault() string { return "postgres" }

func (Postgres) DatabaseDropQuery(_ string) string {
	return "drop schema public cascade; create schema public;"
}

func (Postgres) AnalyzeQuery() string { return "analyze;" }

func (db Postgres) PodFilters() filter.Filter {
	return filter.Or{
		db.query(),
		db.postgresqlHaQuery(),
		db.cnpgQuery(),
		db.zalandoQuery(),
	}
}

func (db Postgres) FilterPods(
	ctx context.Context,
	client kubernetes.KubeClient,
	pods []corev1.Pod,
) ([]corev1.Pod, error) {
	preferred := make([]corev1.Pod, 0, len(pods))

	// bitnami/postgres
	if matched := filter.Pods(pods, db.query()); len(pods) != 0 {
		preferred = append(preferred, matched...)
	}

	logger := slog.With("dialect", db.Name())

	// bitnami/postgresql-ha
	if matched := filter.Pods(pods, db.postgresqlHaQuery()); len(matched) != 0 {
		// HA chart. Need to detect primary.
		logger.Debug("Querying Bitnami repmgr for primary instance")
		cmd := command.NewBuilder(
			command.NewEnv("DISABLE_WELCOME_MESSAGE", "true"),
			"/opt/bitnami/scripts/postgresql-repmgr/entrypoint.sh",
			"repmgr", "--config-file=/opt/bitnami/repmgr/conf/repmgr.conf",
			"service", "status", "--csv",
		)

		var buf bytes.Buffer
		var errBuf strings.Builder
		if err := client.Exec(ctx, kubernetes.ExecOptions{
			Pod:    matched[0],
			Cmd:    cmd.String(),
			Stdout: &buf,
			Stderr: &errBuf,
		}); err != nil {
			return pods, fmt.Errorf("%w: %s", err, errBuf.String())
		}

		var primaryName string
		r := csv.NewReader(&buf)
		for {
			row, err := r.Read()
			if err != nil {
				if errors.Is(err, io.EOF) {
					break
				}
				return pods, err
			}
			if row[2] == "primary" {
				primaryName = row[1]
				break
			}
		}

		for _, pod := range matched {
			if pod.Name == primaryName {
				preferred = append(preferred, pod)
				break
			}
		}
	}

	// CloudNativePG
	if matched := filter.Pods(pods, db.cnpgQuery()); len(matched) != 0 {
		logger.Debug("Finding CloudNativePG Leader")

		for _, pod := range matched {
			if role, ok := pod.Labels["cnpg.io/instanceRole"]; ok && role == "primary" {
				preferred = append(preferred, pod)
			}
		}
	}

	// Zalando Postgres Operator
	if matched := filter.Pods(pods, db.zalandoQuery()); len(matched) != 0 {
		logger.Debug("Finding Zalando Leader")

		for _, pod := range matched {
			if role, ok := pod.Labels["spilo-role"]; ok && role == "master" {
				preferred = append(preferred, pod)
			}
		}
	}

	return preferred, nil
}

func (db Postgres) PasswordEnvs(conf *conftypes.Global) kubernetes.ConfigLookups {
	if secret := db.cnpgSecretName(conf); secret != "" {
		return kubernetes.ConfigLookups{kubernetes.LookupNamedSecret{
			Name: secret,
			Key:  "password",
		}}
	}

	var envs kubernetes.LookupEnv
	if conf.Username == db.UserDefault() {
		envs = append(envs, "POSTGRES_POSTGRES_PASSWORD")
	}
	envs = append(envs, "POSTGRES_PASSWORD", "PGPOOL_POSTGRES_PASSWORD", "PGPASSWORD_SUPERUSER")
	return kubernetes.ConfigLookups{envs}
}

func (Postgres) newCmd(conf *conftypes.Global, p ...any) *command.Builder {
	cmd := command.NewBuilder(p...)
	if conf.Host != "" {
		cmd.Push("--host=" + conf.Host)
	}
	if conf.Port != 0 {
		cmd.Push("--port=" + strconv.Itoa(int(conf.Port)))
	}
	if conf.Username != "" {
		cmd.Push("--username=" + conf.Username)
	}
	if conf.Password != "" {
		cmd.Unshift(command.NewEnv("PGPASSWORD", conf.Password))
	}
	if conf.Database != "" {
		cmd.Push("--dbname=" + conf.Database)
	}
	return cmd
}

func (db Postgres) ExecCommand(conf *conftypes.Exec) *command.Builder {
	cmd := db.newCmd(conf.Global, "exec", "psql")
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

func (db Postgres) DumpCommand(conf *conftypes.Dump) *command.Builder {
	cmd := db.newCmd(conf.Global, "pg_dump")
	if conf.Clean {
		cmd.Push("--clean")
		if conf.IfExists {
			cmd.Push("--if-exists")
		}
	}
	if conf.NoOwner {
		cmd.Push("--no-owner")
	}
	for _, table := range conf.Table {
		cmd.Push("--table=" + db.quoteParam(table))
	}
	for _, table := range conf.ExcludeTable {
		cmd.Push("--exclude-table=" + db.quoteParam(table))
	}
	for _, table := range conf.ExcludeTableData {
		cmd.Push("--exclude-table-data=" + db.quoteParam(table))
	}
	if conf.Format == sqlformat.Custom {
		cmd.Push("--format=custom")
	}
	if !conf.Quiet {
		cmd.Push("--verbose")
	}
	return cmd
}

func (db Postgres) RestoreCommand(conf *conftypes.Restore, inputFormat sqlformat.Format) *command.Builder {
	var cmd *command.Builder
	if inputFormat == sqlformat.Custom {
		cmd = db.newCmd(conf.Global, "pg_restore", "--format=custom")
		if conf.Clean {
			cmd.Push("--clean")
		}
		if conf.HaltOnError {
			cmd.Push("--exit-on-error")
		}
		if conf.NoOwner {
			cmd.Push("--no-owner")
		}
		if !conf.Quiet {
			cmd.Push("--verbose")
		}
	} else {
		cmd = db.newCmd(conf.Global, "psql")
		if conf.Quiet {
			cmd.Push("--quiet", "--output=/dev/null")
		}
		if conf.HaltOnError {
			cmd.Push("--set=ON_ERROR_STOP=1")
		}
	}
	if conf.Quiet {
		cmd.Unshift(command.NewEnv("PGOPTIONS", "-c client_min_messages=WARNING"))
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

func (Postgres) query() filter.Or {
	return filter.Or{
		filter.And{
			filter.Label{
				Name:     "app.kubernetes.io/name",
				Operator: selection.In,
				Values:   []string{"postgresql", "postgres"},
			},
			filter.Label{
				Name:     "app.kubernetes.io/component",
				Operator: selection.NotIn,
				Values:   []string{"read", "replica"},
			},
		},
		filter.Label{Name: "app", Value: "postgresql"},
	}
}

func (Postgres) postgresqlHaQuery() filter.Filter {
	return filter.And{
		filter.Label{Name: "app.kubernetes.io/name", Value: "postgresql-ha"},
		filter.Label{
			Name:     "app.kubernetes.io/component",
			Operator: selection.NotEquals,
			Value:    "pgpool",
		},
	}
}

func (Postgres) cnpgQuery() filter.Filter {
	return filter.Label{Name: "cnpg.io/cluster", Operator: selection.Exists}
}

func (Postgres) zalandoQuery() filter.Filter {
	return filter.Label{Name: "application", Value: "spilo"}
}

func (Postgres) cnpgClusterName(conf *conftypes.Global) (string, bool) {
	v, ok := conf.DBPod.Labels["cnpg.io/cluster"]
	return v, ok
}

func (db Postgres) cnpgSecretName(conf *conftypes.Global) string {
	if cluster, ok := db.cnpgClusterName(conf); ok {
		if conf.Username == db.UserDefault() {
			return cluster + "-superuser"
		}
		return cluster + "-app"
	}
	return ""
}

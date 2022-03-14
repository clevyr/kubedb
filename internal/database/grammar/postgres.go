package grammar

import (
	"encoding/csv"
	"github.com/clevyr/kubedb/internal/config"
	"github.com/clevyr/kubedb/internal/database/sqlformat"
	"github.com/clevyr/kubedb/internal/kubernetes"
	log "github.com/sirupsen/logrus"
	"io"
	v1 "k8s.io/api/core/v1"
	"strings"
)

type Postgres struct{}

func (Postgres) Name() string {
	return "postgres"
}

func (Postgres) DatabaseEnvNames() []string {
	return []string{"POSTGRES_DB"}
}

func (Postgres) DefaultDatabase() string {
	return "db"
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
		cmd := []string{
			"sh", "-c",
			"DISABLE_WELCOME_MESSAGE=true /opt/bitnami/scripts/postgresql-repmgr/entrypoint.sh " +
				"repmgr --config-file=/opt/bitnami/repmgr/conf/repmgr.conf " +
				"service status --csv",
		}

		var buf strings.Builder
		err := client.Exec(pods[0], cmd, strings.NewReader(""), &buf, false)
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

func (Postgres) ExecCommand(conf config.Exec) []string {
	return []string{"PGPASSWORD=" + conf.Password, "psql", "--host=127.0.0.1", "--username=" + conf.Username, "--dbname=" + conf.Database}
}

func (Postgres) DumpCommand(conf config.Dump) []string {
	cmd := []string{"PGPASSWORD=" + conf.Password, "pg_dump", "--host=127.0.0.1", "--username=" + conf.Username, "--dbname=" + conf.Database}
	if conf.Clean {
		cmd = append(cmd, "--clean")
	}
	if conf.NoOwner {
		cmd = append(cmd, "--no-owner")
	}
	if conf.IfExists {
		cmd = append(cmd, "--if-exists")
	}
	for _, table := range conf.Tables {
		cmd = append(cmd, "--table='"+table+"'")
	}
	for _, table := range conf.ExcludeTable {
		cmd = append(cmd, "--exclude-table='"+table+"'")
	}
	for _, table := range conf.ExcludeTableData {
		cmd = append(cmd, "--exclude-table-data='"+table+"'")
	}
	if conf.OutputFormat == sqlformat.Custom {
		cmd = append(cmd, "--format=c")
	}
	return cmd
}

func (Postgres) RestoreCommand(conf config.Restore, inputFormat sqlformat.Format) []string {
	cmd := []string{"PGPASSWORD=" + conf.Password, "PGOPTIONS='-c client_min_messages=WARNING'"}
	switch inputFormat {
	case sqlformat.Gzip, sqlformat.Plain:
		cmd = append(cmd, "psql", "--quiet", "--output=/dev/null")
	case sqlformat.Custom:
		cmd = append(cmd, "pg_restore", "--format=custom", "--verbose", "--clean", "--exit-on-error")
		if conf.NoOwner {
			cmd = append(cmd, "--no-owner")
		}
	}
	cmd = append(cmd, "--host=127.0.0.1", "--username="+conf.Username, "--dbname="+conf.Database)
	if conf.SingleTransaction {
		cmd = append(cmd, "--single-transaction")
	}
	return cmd
}

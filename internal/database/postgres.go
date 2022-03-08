package database

import (
	"context"
	"github.com/clevyr/kubedb/internal/config"
	"github.com/clevyr/kubedb/internal/database/sqlformat"
	"github.com/clevyr/kubedb/internal/kubernetes"
	v1core "k8s.io/api/core/v1"
	v1meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/selection"
)

type Postgres struct{}

func (Postgres) Name() string {
	return "postgres"
}

func (Postgres) DefaultDatabase() string {
	return "db"
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

func (Postgres) GetPod(client kubernetes.KubeClient) (v1core.Pod, error) {
	return kubernetes.GetPodByLabel(client, "app", selection.Equals, []string{"postgresql"})
}

func (Postgres) GetSecret(client kubernetes.KubeClient) (string, error) {
	secret, err := client.Secrets().Get(context.TODO(), "postgresql", v1meta.GetOptions{})
	if err != nil {
		return "", err
	}
	return string(secret.Data["postgresql-password"]), err
}

func (Postgres) ExecCommand(conf config.Exec) []string {
	return []string{"PGPASSWORD=" + conf.Password, "psql", "--username=" + conf.Username, "--dbname=" + conf.Database}
}

func (Postgres) DumpCommand(conf config.Dump) []string {
	cmd := []string{"PGPASSWORD=" + conf.Password, "pg_dump", "--username=" + conf.Username, "--dbname=" + conf.Database}
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
	cmd := []string{"PGPASSWORD=" + conf.Password}
	switch inputFormat {
	case sqlformat.Gzip, sqlformat.Plain:
		cmd = append(cmd, "psql")
	case sqlformat.Custom:
		cmd = append(cmd, "pg_restore", "--format=custom", "--verbose")
		if conf.NoOwner {
			cmd = append(cmd, "--no-owner")
		}
	}
	cmd = append(cmd, "--username="+conf.Username, "--dbname="+conf.Database)
	if conf.SingleTransaction {
		cmd = append(cmd, "--single-transaction")
	}
	return cmd
}

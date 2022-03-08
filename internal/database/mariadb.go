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

type MariaDB struct{}

func (MariaDB) Name() string {
	return "mariadb"
}

func (MariaDB) DefaultDatabase() string {
	return "db"
}

func (MariaDB) DefaultUser() string {
	return "mariadb"
}

func (MariaDB) DropDatabaseQuery(database string) string {
	return "set FOREIGN_KEY_CHECKS=0; create or replace database " + database + "; set FOREIGN_KEY_CHECKS=1;"
}

func (MariaDB) AnalyzeQuery() string {
	return ";"
}

func (MariaDB) GetPod(client kubernetes.KubeClient) (v1core.Pod, error) {
	return kubernetes.GetPodByLabel(client, "app", selection.Equals, []string{"mariadb"})
}

func (MariaDB) GetSecret(client kubernetes.KubeClient) (string, error) {
	secret, err := client.Secrets().Get(context.TODO(), "mariadb", v1meta.GetOptions{})
	if err != nil {
		return "", err
	}
	return string(secret.Data["mariadb-password"]), err
}

func (MariaDB) ExecCommand(conf config.Exec) []string {
	return []string{"MYSQL_PWD=" + conf.Password, "mysql", "--user=" + conf.Username, "--database=" + conf.Database}
}

func (MariaDB) DumpCommand(conf config.Dump) []string {
	cmd := []string{"MYSQL_PWD=" + conf.Password, "mysqldump", "--user=" + conf.Username, conf.Database}
	if conf.Clean {
		cmd = append(cmd, "--add-drop-table")
	}
	for _, table := range conf.Tables {
		cmd = append(cmd, table)
	}
	for _, table := range conf.ExcludeTable {
		cmd = append(cmd, "--ignore-table='"+table+"'")
	}
	return cmd
}

func (MariaDB) RestoreCommand(conf config.Restore, inputFormat sqlformat.Format) []string {
	return []string{"MYSQL_PWD=" + conf.Password, "mysql", "--user=" + conf.Username, conf.Database}
}

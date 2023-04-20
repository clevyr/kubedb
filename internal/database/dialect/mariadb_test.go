package dialect

import (
	"context"
	"testing"

	"github.com/clevyr/kubedb/internal/command"
	"github.com/clevyr/kubedb/internal/config"
	"github.com/clevyr/kubedb/internal/database/sqlformat"
	"github.com/clevyr/kubedb/internal/kubernetes"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
)

func TestMariaDB_AnalyzeQuery(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{"default", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ma := MariaDB{}
			got := ma.AnalyzeQuery()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestMariaDB_DatabaseEnvNames(t *testing.T) {
	tests := []struct {
		name string
		want []string
	}{
		{"default", []string{"MARIADB_DATABASE"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ma := MariaDB{}
			got := ma.DatabaseEnvNames()
			assert.Equal(t, got, tt.want)
		})
	}
}

func TestMariaDB_DefaultPort(t *testing.T) {
	tests := []struct {
		name string
		want uint16
	}{
		{"default", 3306},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ma := MariaDB{}
			got := ma.DefaultPort()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestMariaDB_DefaultUser(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{"default", "root"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ma := MariaDB{}
			got := ma.DefaultUser()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestMariaDB_DropDatabaseQuery(t *testing.T) {
	type args struct {
		database string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"database", args{"database"}, "set FOREIGN_KEY_CHECKS=0; create or replace database database; set FOREIGN_KEY_CHECKS=1; use database;"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ma := MariaDB{}
			got := ma.DropDatabaseQuery(tt.args.database)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestMariaDB_DumpCommand(t *testing.T) {
	type args struct {
		conf config.Dump
	}
	tests := []struct {
		name string
		args args
		want *command.Builder
	}{
		{
			"default",
			args{config.Dump{Global: config.Global{Database: "d", Username: "u"}}},
			command.NewBuilder(command.NewEnv("MYSQL_PWD", ""), "mysqldump", "--host=127.0.0.1", "--user=u", "d", "--verbose"),
		},
		{
			"clean",
			args{config.Dump{Clean: true, Global: config.Global{Database: "d", Username: "u"}}},
			command.NewBuilder(command.NewEnv("MYSQL_PWD", ""), "mysqldump", "--host=127.0.0.1", "--user=u", "d", "--add-drop-table", "--verbose"),
		},
		{
			"tables",
			args{config.Dump{Tables: []string{"table1", "table2"}, Global: config.Global{Database: "d", Username: "u"}}},
			command.NewBuilder(command.NewEnv("MYSQL_PWD", ""), "mysqldump", "--host=127.0.0.1", "--user=u", "d", "table1", "table2", "--verbose"),
		},
		{
			"exclude-table",
			args{config.Dump{ExcludeTable: []string{"table1", "table2"}, Global: config.Global{Database: "d", Username: "u"}}},
			command.NewBuilder(command.NewEnv("MYSQL_PWD", ""), "mysqldump", "--host=127.0.0.1", "--user=u", "d", "--ignore-table=table1", "--ignore-table=table2", "--verbose"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ma := MariaDB{}
			got := ma.DumpCommand(tt.args.conf)
			assert.Equal(t, got, tt.want)
		})
	}
}

func TestMariaDB_ExecCommand(t *testing.T) {
	type args struct {
		conf config.Exec
	}
	tests := []struct {
		name string
		args args
		want *command.Builder
	}{
		{
			"default",
			args{config.Exec{Global: config.Global{Database: "d", Username: "u"}}},
			command.NewBuilder(command.Env{Key: "MYSQL_PWD"}, "mysql", "--host=127.0.0.1", "--user=u", "--database=d"),
		},
		{
			"disable-headers",
			args{config.Exec{DisableHeaders: true, Global: config.Global{Database: "d", Username: "u"}}},
			command.NewBuilder(command.Env{Key: "MYSQL_PWD"}, "mysql", "--host=127.0.0.1", "--user=u", "--database=d", "--skip-column-names"),
		},
		{
			"command",
			args{config.Exec{Command: "show databases", Global: config.Global{Database: "d", Username: "u"}}},
			command.NewBuilder(command.Env{Key: "MYSQL_PWD"}, "mysql", "--host=127.0.0.1", "--user=u", "--database=d", "--execute=show databases"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ma := MariaDB{}
			got := ma.ExecCommand(tt.args.conf)
			assert.Equal(t, got, tt.want)
		})
	}
}

func TestMariaDB_FilterPods(t *testing.T) {
	type args struct {
		client kubernetes.KubeClient
		pods   []v1.Pod
	}
	tests := []struct {
		name    string
		args    args
		want    []v1.Pod
		wantErr bool
	}{
		{"empty", args{kubernetes.KubeClient{}, []v1.Pod{}}, []v1.Pod{}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ma := MariaDB{}
			got, err := ma.FilterPods(context.TODO(), tt.args.client, tt.args.pods)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, got, tt.want)
		})
	}
}

func TestMariaDB_ListDatabasesQuery(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{"default", "show databases"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ma := MariaDB{}
			got := ma.ListDatabasesQuery()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestMariaDB_ListTablesQuery(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{"default", "show tables"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ma := MariaDB{}
			got := ma.ListTablesQuery()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestMariaDB_Name(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{"default", "mariadb"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ma := MariaDB{}
			got := ma.Name()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestMariaDB_PasswordEnvNames(t *testing.T) {
	type args struct {
		c config.Global
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{"default", args{}, []string{"MARIADB_PASSWORD"}},
		{"root", args{config.Global{Username: "root"}}, []string{"MARIADB_ROOT_PASSWORD"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := MariaDB{}
			got := db.PasswordEnvNames(tt.args.c)
			assert.Equal(t, got, tt.want)
		})
	}
}

func TestMariaDB_PodLabels(t *testing.T) {
	tests := []struct {
		name string
		want []kubernetes.LabelQueryable
	}{
		{"default", []kubernetes.LabelQueryable{
			kubernetes.LabelQueryAnd{
				{Name: "app.kubernetes.io/name", Value: "mariadb"},
				{Name: "app.kubernetes.io/component", Value: "primary"},
			},
			kubernetes.LabelQuery{Name: "app", Value: "mariadb"},
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ma := MariaDB{}
			got := ma.PodLabels()
			assert.Equal(t, got, tt.want)
		})
	}
}

func TestMariaDB_RestoreCommand(t *testing.T) {
	type args struct {
		conf        config.Restore
		inputFormat sqlformat.Format
	}
	tests := []struct {
		name string
		args args
		want *command.Builder
	}{
		{
			"gzip",
			args{config.Restore{Global: config.Global{Database: "d", Username: "u"}}, sqlformat.Gzip},
			command.NewBuilder(command.Env{Key: "MYSQL_PWD"}, "mysql", "--host=127.0.0.1", "--user=u", "--database=d"),
		},
		{
			"plain",
			args{config.Restore{Global: config.Global{Database: "d", Username: "u"}}, sqlformat.Plain},
			command.NewBuilder(command.Env{Key: "MYSQL_PWD"}, "mysql", "--host=127.0.0.1", "--user=u", "--database=d"),
		},
		{
			"custom",
			args{config.Restore{Global: config.Global{Database: "d", Username: "u"}}, sqlformat.Custom},
			command.NewBuilder(command.Env{Key: "MYSQL_PWD"}, "mysql", "--host=127.0.0.1", "--user=u", "--database=d"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ma := MariaDB{}
			got := ma.RestoreCommand(tt.args.conf, tt.args.inputFormat)
			assert.Equal(t, got, tt.want)
		})
	}
}

func TestMariaDB_UserEnvNames(t *testing.T) {
	tests := []struct {
		name string
		want []string
	}{
		{"default", []string{"MARIADB_USER"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ma := MariaDB{}
			got := ma.UserEnvNames()
			assert.Equal(t, got, tt.want)
		})
	}
}

func TestMariaDB_DumpExtension(t *testing.T) {
	type args struct {
		format sqlformat.Format
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"plain", args{sqlformat.Plain}, ".sql"},
		{"gzip", args{sqlformat.Gzip}, ".sql.gz"},
		{"custom", args{sqlformat.Custom}, ""},
		{"unknown", args{sqlformat.Unknown}, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			po := MariaDB{}
			got := po.DumpExtension(tt.args.format)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestMariaDB_FormatFromFilename(t *testing.T) {
	type args struct {
		filename string
	}
	tests := []struct {
		name string
		args args
		want sqlformat.Format
	}{
		{"gzip", args{"test.sql.gz"}, sqlformat.Gzip},
		{"plain", args{"test.sql"}, sqlformat.Plain},
		{"unknown", args{"test.png"}, sqlformat.Unknown},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ma := MariaDB{}
			got := ma.FormatFromFilename(tt.args.filename)
			assert.Equal(t, tt.want, got)
		})
	}
}

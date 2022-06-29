package dialect

import (
	"github.com/clevyr/kubedb/internal/command"
	"github.com/clevyr/kubedb/internal/config"
	"github.com/clevyr/kubedb/internal/database/sqlformat"
	"github.com/clevyr/kubedb/internal/kubernetes"
	v1 "k8s.io/api/core/v1"
	"reflect"
	"testing"
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
			if got := ma.AnalyzeQuery(); got != tt.want {
				t.Errorf("AnalyzeQuery() = %v, want %v", got, tt.want)
			}
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
			if got := ma.DatabaseEnvNames(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DatabaseEnvNames() = %v, want %v", got, tt.want)
			}
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
			if got := ma.DefaultPort(); got != tt.want {
				t.Errorf("DefaultPort() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMariaDB_DefaultUser(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{"default", "mariadb"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ma := MariaDB{}
			if got := ma.DefaultUser(); got != tt.want {
				t.Errorf("DefaultUser() = %v, want %v", got, tt.want)
			}
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
		{"database", args{"database"}, "set FOREIGN_KEY_CHECKS=0; create or replace database database; set FOREIGN_KEY_CHECKS=1;"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ma := MariaDB{}
			if got := ma.DropDatabaseQuery(tt.args.database); got != tt.want {
				t.Errorf("DropDatabaseQuery() = %v, want %v", got, tt.want)
			}
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
			command.NewBuilder(command.NewEnv("MYSQL_PWD", ""), "mysqldump", "--host=127.0.0.1", "--user=u", "d"),
		},
		{
			"clean",
			args{config.Dump{Clean: true, Global: config.Global{Database: "d", Username: "u"}}},
			command.NewBuilder(command.NewEnv("MYSQL_PWD", ""), "mysqldump", "--host=127.0.0.1", "--user=u", "d", "--add-drop-table"),
		},
		{
			"tables",
			args{config.Dump{Tables: []string{"table1", "table2"}, Global: config.Global{Database: "d", Username: "u"}}},
			command.NewBuilder(command.NewEnv("MYSQL_PWD", ""), "mysqldump", "--host=127.0.0.1", "--user=u", "d", "table1", "table2"),
		},
		{
			"exclude-table",
			args{config.Dump{ExcludeTable: []string{"table1", "table2"}, Global: config.Global{Database: "d", Username: "u"}}},
			command.NewBuilder(command.NewEnv("MYSQL_PWD", ""), "mysqldump", "--host=127.0.0.1", "--user=u", "d", "--ignore-table=table1", "--ignore-table=table2"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ma := MariaDB{}
			if got := ma.DumpCommand(tt.args.conf); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DumpCommand() = %v, want %v", got, tt.want)
			}
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ma := MariaDB{}
			if got := ma.ExecCommand(tt.args.conf); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ExecCommand() = %v, want %v", got, tt.want)
			}
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
			got, err := ma.FilterPods(tt.args.client, tt.args.pods)
			if (err != nil) != tt.wantErr {
				t.Errorf("FilterPods() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FilterPods() got = %v, want %v", got, tt.want)
			}
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
			if got := ma.ListDatabasesQuery(); got != tt.want {
				t.Errorf("ListDatabasesQuery() = %v, want %v", got, tt.want)
			}
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
			if got := ma.ListTablesQuery(); got != tt.want {
				t.Errorf("ListTablesQuery() = %v, want %v", got, tt.want)
			}
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
			if got := ma.Name(); got != tt.want {
				t.Errorf("Name() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMariaDB_PasswordEnvNames(t *testing.T) {
	tests := []struct {
		name string
		want []string
	}{
		{"default", []string{"MARIADB_PASSWORD"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ma := MariaDB{}
			if got := ma.PasswordEnvNames(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("PasswordEnvNames() = %v, want %v", got, tt.want)
			}
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
			if got := ma.PodLabels(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("PodLabels() = %v, want %v", got, tt.want)
			}
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
			command.NewBuilder(command.Env{Key: "MYSQL_PWD"}, "mysql", "--host=127.0.0.1", "--user=u", "d"),
		},
		{
			"plain",
			args{config.Restore{Global: config.Global{Database: "d", Username: "u"}}, sqlformat.Plain},
			command.NewBuilder(command.Env{Key: "MYSQL_PWD"}, "mysql", "--host=127.0.0.1", "--user=u", "d"),
		},
		{
			"custom",
			args{config.Restore{Global: config.Global{Database: "d", Username: "u"}}, sqlformat.Custom},
			command.NewBuilder(command.Env{Key: "MYSQL_PWD"}, "mysql", "--host=127.0.0.1", "--user=u", "d"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ma := MariaDB{}
			if got := ma.RestoreCommand(tt.args.conf, tt.args.inputFormat); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("RestoreCommand() = %v, want %v", got, tt.want)
			}
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
			if got := ma.UserEnvNames(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("UserEnvNames() = %v, want %v", got, tt.want)
			}
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
			if got := po.DumpExtension(tt.args.format); got != tt.want {
				t.Errorf("DumpFileExt() = %v, want %v", got, tt.want)
			}
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
			if got := ma.FormatFromFilename(tt.args.filename); got != tt.want {
				t.Errorf("FormatFromFilename() = %v, want %v", got, tt.want)
			}
		})
	}
}

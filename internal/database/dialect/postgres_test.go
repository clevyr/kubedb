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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestPostgres_AnalyzeQuery(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{"default", "analyze;"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			po := Postgres{}
			got := po.AnalyzeQuery()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestPostgres_DatabaseEnvNames(t *testing.T) {
	tests := []struct {
		name string
		want []string
	}{
		{"default", []string{"POSTGRES_DATABASE", "POSTGRES_DB"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			po := Postgres{}
			got := po.DatabaseEnvNames()
			assert.Equal(t, got, tt.want)
		})
	}
}

func TestPostgres_DefaultPort(t *testing.T) {
	tests := []struct {
		name string
		want uint16
	}{
		{"default", 5432},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			po := Postgres{}
			got := po.DefaultPort()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestPostgres_DefaultUser(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{"default", "postgres"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			po := Postgres{}
			got := po.DefaultUser()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestPostgres_DropDatabaseQuery(t *testing.T) {
	type args struct {
		database string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"default", args{"database"}, "drop schema public cascade; create schema public;"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			po := Postgres{}
			got := po.DropDatabaseQuery(tt.args.database)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestPostgres_DumpCommand(t *testing.T) {
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
			command.NewBuilder(command.NewEnv("PGPASSWORD", ""), "pg_dump", "--host=127.0.0.1", "--username=u", "--dbname=d", "--verbose"),
		},
		{
			"clean",
			args{config.Dump{Clean: true, Global: config.Global{Database: "d", Username: "u"}}},
			command.NewBuilder(command.NewEnv("PGPASSWORD", ""), "pg_dump", "--host=127.0.0.1", "--username=u", "--dbname=d", "--clean", "--verbose"),
		},
		{
			"no-owner",
			args{config.Dump{NoOwner: true, Global: config.Global{Database: "d", Username: "u"}}},
			command.NewBuilder(command.NewEnv("PGPASSWORD", ""), "pg_dump", "--host=127.0.0.1", "--username=u", "--dbname=d", "--no-owner", "--verbose"),
		},
		{
			"if-exists",
			args{config.Dump{IfExists: true, Global: config.Global{Database: "d", Username: "u"}}},
			command.NewBuilder(command.NewEnv("PGPASSWORD", ""), "pg_dump", "--host=127.0.0.1", "--username=u", "--dbname=d", "--if-exists", "--verbose"),
		},
		{
			"tables",
			args{config.Dump{Tables: []string{"table1", "table2"}, Global: config.Global{Database: "d", Username: "u"}}},
			command.NewBuilder(command.NewEnv("PGPASSWORD", ""), "pg_dump", "--host=127.0.0.1", "--username=u", "--dbname=d", `--table="table1"`, `--table="table2"`, "--verbose"),
		},
		{
			"exclude-table",
			args{config.Dump{ExcludeTable: []string{"table1", "table2"}, Global: config.Global{Database: "d", Username: "u"}}},
			command.NewBuilder(command.NewEnv("PGPASSWORD", ""), "pg_dump", "--host=127.0.0.1", "--username=u", "--dbname=d", `--exclude-table="table1"`, `--exclude-table="table2"`, "--verbose"),
		},
		{
			"exclude-table-data",
			args{config.Dump{ExcludeTableData: []string{"table1", "table2"}, Global: config.Global{Database: "d", Username: "u"}}},
			command.NewBuilder(command.NewEnv("PGPASSWORD", ""), "pg_dump", "--host=127.0.0.1", "--username=u", "--dbname=d", `--exclude-table-data="table1"`, `--exclude-table-data="table2"`, "--verbose"),
		},
		{
			"custom",
			args{config.Dump{Files: config.Files{Format: sqlformat.Custom}, Global: config.Global{Database: "d", Username: "u"}}},
			command.NewBuilder(command.NewEnv("PGPASSWORD", ""), "pg_dump", "--host=127.0.0.1", "--username=u", "--dbname=d", "--format=c", "--verbose"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			po := Postgres{}
			got := po.DumpCommand(tt.args.conf)
			assert.Equal(t, got, tt.want)
		})
	}
}

func TestPostgres_ExecCommand(t *testing.T) {
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
			command.NewBuilder(command.NewEnv("PGPASSWORD", ""), "psql", "--host=127.0.0.1", "--username=u", "--dbname=d"),
		},
		{
			"disable-headers",
			args{config.Exec{DisableHeaders: true, Global: config.Global{Database: "d", Username: "u"}}},
			command.NewBuilder(command.NewEnv("PGPASSWORD", ""), "psql", "--host=127.0.0.1", "--username=u", "--dbname=d", "--tuples-only"),
		},
		{
			"command",
			args{config.Exec{Command: "select true", Global: config.Global{Database: "d", Username: "u"}}},
			command.NewBuilder(command.NewEnv("PGPASSWORD", ""), "psql", "--host=127.0.0.1", "--username=u", "--dbname=d", "--command=select true"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			po := Postgres{}
			got := po.ExecCommand(tt.args.conf)
			assert.Equal(t, got, tt.want)
		})
	}
}

func TestPostgres_FilterPods(t *testing.T) {
	postgresPod := v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{
				"app.kubernetes.io/name": "postgresql",
			},
		},
	}

	//postgresHaPod := v1.Pod{
	//	ObjectMeta: metav1.ObjectMeta{
	//		Labels: map[string]string{
	//			"app.kubernetes.io/name": "postgresql-ha",
	//		},
	//	},
	//}

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
		{
			"postgresql",
			args{
				kubernetes.KubeClient{},
				[]v1.Pod{postgresPod},
			},
			[]v1.Pod{postgresPod},
			false,
		},
		//{
		//	"postgresql-ha",
		//	args{
		//		kubernetes.KubeClient{ClientSet: testclient.NewSimpleClientset(&postgresHaPod)},
		//		[]v1.Pod{postgresHaPod},
		//	},
		//	[]v1.Pod{postgresHaPod},
		//	false,
		//},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			po := Postgres{}
			got, err := po.FilterPods(context.TODO(), tt.args.client, tt.args.pods)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, got, tt.want)
		})
	}
}

func TestPostgres_ListDatabasesQuery(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{"default", "SELECT datname FROM pg_database WHERE datistemplate = false"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			po := Postgres{}
			got := po.ListDatabasesQuery()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestPostgres_ListTablesQuery(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{"default", "SELECT table_name FROM information_schema.tables WHERE table_schema='public' AND table_type='BASE TABLE'"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			po := Postgres{}
			got := po.ListTablesQuery()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestPostgres_Name(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{"default", "postgres"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			po := Postgres{}
			got := po.Name()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestPostgres_PasswordEnvNames(t *testing.T) {
	type args struct {
		c config.Global
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{"default", args{}, []string{"POSTGRES_PASSWORD", "PGPOOL_POSTGRES_PASSWORD"}},
		{"postgres", args{config.Global{Username: "postgres"}}, []string{"POSTGRES_POSTGRES_PASSWORD", "POSTGRES_PASSWORD", "PGPOOL_POSTGRES_PASSWORD"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := Postgres{}
			got := db.PasswordEnvNames(tt.args.c)
			assert.Equal(t, got, tt.want)
		})
	}
}

func TestPostgres_PodLabels(t *testing.T) {
	tests := []struct {
		name string
		want []kubernetes.LabelQueryable
	}{
		{"default", []kubernetes.LabelQueryable{
			kubernetes.LabelQueryAnd{
				{Name: "app.kubernetes.io/name", Value: "postgresql"},
				{Name: "app.kubernetes.io/component", Value: "primary"},
			},
			kubernetes.LabelQueryAnd{
				{Name: "app.kubernetes.io/name", Value: "postgresql-ha"},
				{Name: "app.kubernetes.io/component", Value: "postgresql"},
			},
			kubernetes.LabelQuery{Name: "app", Value: "postgresql"},
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			po := Postgres{}
			got := po.PodLabels()
			assert.Equal(t, got, tt.want)
		})
	}
}

func TestPostgres_RestoreCommand(t *testing.T) {
	pgpassword := command.NewEnv("PGPASSWORD", "")

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
			command.NewBuilder(pgpassword, "psql", "--host=127.0.0.1", "--username=u", "--dbname=d"),
		},
		{
			"plain",
			args{config.Restore{Global: config.Global{Database: "d", Username: "u"}}, sqlformat.Plain},
			command.NewBuilder(pgpassword, "psql", "--host=127.0.0.1", "--username=u", "--dbname=d"),
		},
		{
			"custom",
			args{config.Restore{Global: config.Global{Database: "d", Username: "u"}}, sqlformat.Custom},
			command.NewBuilder(pgpassword, "pg_restore", "--format=custom", "--clean", "--exit-on-error", "--verbose", "--host=127.0.0.1", "--username=u", "--dbname=d"),
		},
		{
			"custom-no-owner",
			args{config.Restore{NoOwner: true, Global: config.Global{Database: "d", Username: "u"}}, sqlformat.Custom},
			command.NewBuilder(pgpassword, "pg_restore", "--format=custom", "--clean", "--exit-on-error", "--no-owner", "--verbose", "--host=127.0.0.1", "--username=u", "--dbname=d"),
		},
		{
			"single-transaction",
			args{config.Restore{SingleTransaction: true, Global: config.Global{Database: "d", Username: "u"}}, sqlformat.Custom},
			command.NewBuilder(pgpassword, "pg_restore", "--format=custom", "--clean", "--exit-on-error", "--verbose", "--host=127.0.0.1", "--username=u", "--dbname=d", "--single-transaction"),
		},
		{
			"sql-quiet",
			args{config.Restore{Global: config.Global{Database: "d", Username: "u", Quiet: true}}, sqlformat.Gzip},
			command.NewBuilder(pgpassword, command.NewEnv("PGOPTIONS", "-c client_min_messages=WARNING"), "psql", "--quiet", "--output=/dev/null", "--host=127.0.0.1", "--username=u", "--dbname=d"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			po := Postgres{}
			got := po.RestoreCommand(tt.args.conf, tt.args.inputFormat)
			assert.Equal(t, got, tt.want)
		})
	}
}

func TestPostgres_UserEnvNames(t *testing.T) {
	tests := []struct {
		name string
		want []string
	}{
		{"default", []string{"POSTGRES_USER", "PGPOOL_POSTGRES_USERNAME"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			po := Postgres{}
			got := po.UserEnvNames()
			assert.Equal(t, got, tt.want)
		})
	}
}

func Test_quoteParam(t *testing.T) {
	type args struct {
		param string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"simple", args{"table"}, `"table"`},
		{"capital", args{"Table"}, `"Table"`},
		{"wildcard-prefix", args{"*Table"}, `""*"Table"`},
		{"wildcard", args{"T*ble"}, `"T"*"ble"`},
		{"wildcard-suffix", args{"Table*"}, `"Table"*""`},
		{"wildcard-both", args{"*Table*"}, `""*"Table"*""`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := quoteParam(tt.args.param)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestPostgres_DumpExtension(t *testing.T) {
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
		{"custom", args{sqlformat.Custom}, ".dmp"},
		{"other", args{sqlformat.Unknown}, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			po := Postgres{}
			got := po.DumpExtension(tt.args.format)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestPostgres_FormatFromFilename(t *testing.T) {
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
		{"custom", args{"test.dmp"}, sqlformat.Custom},
		{"unknown", args{"test.png"}, sqlformat.Unknown},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ma := Postgres{}
			got := ma.FormatFromFilename(tt.args.filename)
			assert.Equal(t, tt.want, got)
		})
	}
}

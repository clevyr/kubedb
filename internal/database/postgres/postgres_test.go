package postgres

import (
	"context"
	"testing"

	"github.com/clevyr/kubedb/internal/command"
	"github.com/clevyr/kubedb/internal/config"
	"github.com/clevyr/kubedb/internal/database/sqlformat"
	"github.com/clevyr/kubedb/internal/kubernetes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func newCNPGPod() v1.Pod {
	return v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: "postgresql-1",
			Labels: map[string]string{
				"cnpg.io/cluster": "postgresql",
			},
		},
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
			args{config.Dump{Global: config.Global{Host: "1.1.1.1", Database: "d", Username: "u"}}},
			command.NewBuilder(command.NewEnv("PGPASSWORD", ""), "pg_dump", "--host=1.1.1.1", "--username=u", "--dbname=d", "--verbose"),
		},
		{
			"clean",
			args{config.Dump{Clean: true, Global: config.Global{Host: "1.1.1.1", Database: "d", Username: "u"}}},
			command.NewBuilder(command.NewEnv("PGPASSWORD", ""), "pg_dump", "--host=1.1.1.1", "--username=u", "--dbname=d", "--clean", "--verbose"),
		},
		{
			"no-owner",
			args{config.Dump{NoOwner: true, Global: config.Global{Host: "1.1.1.1", Database: "d", Username: "u"}}},
			command.NewBuilder(command.NewEnv("PGPASSWORD", ""), "pg_dump", "--host=1.1.1.1", "--username=u", "--dbname=d", "--no-owner", "--verbose"),
		},
		{
			"if-exists",
			args{config.Dump{Clean: true, IfExists: true, Global: config.Global{Host: "1.1.1.1", Database: "d", Username: "u"}}},
			command.NewBuilder(command.NewEnv("PGPASSWORD", ""), "pg_dump", "--host=1.1.1.1", "--username=u", "--dbname=d", "--clean", "--if-exists", "--verbose"),
		},
		{
			"if-exists without clean",
			args{config.Dump{IfExists: true, Global: config.Global{Host: "1.1.1.1", Database: "d", Username: "u"}}},
			command.NewBuilder(command.NewEnv("PGPASSWORD", ""), "pg_dump", "--host=1.1.1.1", "--username=u", "--dbname=d", "--verbose"),
		},
		{
			"tables",
			args{config.Dump{Tables: []string{"table1", "table2"}, Global: config.Global{Host: "1.1.1.1", Database: "d", Username: "u"}}},
			command.NewBuilder(command.NewEnv("PGPASSWORD", ""), "pg_dump", "--host=1.1.1.1", "--username=u", "--dbname=d", `--table="table1"`, `--table="table2"`, "--verbose"),
		},
		{
			"exclude-table",
			args{config.Dump{ExcludeTable: []string{"table1", "table2"}, Global: config.Global{Host: "1.1.1.1", Database: "d", Username: "u"}}},
			command.NewBuilder(command.NewEnv("PGPASSWORD", ""), "pg_dump", "--host=1.1.1.1", "--username=u", "--dbname=d", `--exclude-table="table1"`, `--exclude-table="table2"`, "--verbose"),
		},
		{
			"exclude-table-data",
			args{config.Dump{ExcludeTableData: []string{"table1", "table2"}, Global: config.Global{Host: "1.1.1.1", Database: "d", Username: "u"}}},
			command.NewBuilder(command.NewEnv("PGPASSWORD", ""), "pg_dump", "--host=1.1.1.1", "--username=u", "--dbname=d", `--exclude-table-data="table1"`, `--exclude-table-data="table2"`, "--verbose"),
		},
		{
			"custom",
			args{config.Dump{Files: config.Files{Format: sqlformat.Custom}, Global: config.Global{Host: "1.1.1.1", Database: "d", Username: "u"}}},
			command.NewBuilder(command.NewEnv("PGPASSWORD", ""), "pg_dump", "--host=1.1.1.1", "--username=u", "--dbname=d", "--format=custom", "--verbose"),
		},
		{
			"port",
			args{config.Dump{Global: config.Global{Port: 1234}}},
			command.NewBuilder(command.NewEnv("PGPASSWORD", ""), "pg_dump", "--host=", "--username=", "--port=1234", "--verbose"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			po := Postgres{}
			got := po.DumpCommand(tt.args.conf)
			assert.Equal(t, tt.want, got)
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
			args{config.Exec{Global: config.Global{Host: "1.1.1.1", Database: "d", Username: "u"}}},
			command.NewBuilder(command.NewEnv("PGPASSWORD", ""), "exec", "psql", "--host=1.1.1.1", "--username=u", "--dbname=d"),
		},
		{
			"disable-headers",
			args{config.Exec{DisableHeaders: true, Global: config.Global{Host: "1.1.1.1", Database: "d", Username: "u"}}},
			command.NewBuilder(command.NewEnv("PGPASSWORD", ""), "exec", "psql", "--host=1.1.1.1", "--username=u", "--dbname=d", "--tuples-only"),
		},
		{
			"command",
			args{config.Exec{Command: "select true", Global: config.Global{Host: "1.1.1.1", Database: "d", Username: "u"}}},
			command.NewBuilder(command.NewEnv("PGPASSWORD", ""), "exec", "psql", "--host=1.1.1.1", "--username=u", "--dbname=d", "--command=select true"),
		},
		{
			"default",
			args{config.Exec{Global: config.Global{Port: 1234}}},
			command.NewBuilder(command.NewEnv("PGPASSWORD", ""), "exec", "psql", "--host=", "--username=", "--port=1234"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			po := Postgres{}
			got := po.ExecCommand(tt.args.conf)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestPostgres_FilterPods(t *testing.T) {
	postgresPod := v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{
				"app.kubernetes.io/name":      "postgresql",
				"app.kubernetes.io/component": "primary",
			},
		},
	}

	type args struct {
		client kubernetes.KubeClient
		pods   []v1.Pod
	}
	tests := []struct {
		name    string
		args    args
		want    []v1.Pod
		wantErr require.ErrorAssertionFunc
	}{
		{
			"postgresql",
			args{
				kubernetes.KubeClient{},
				[]v1.Pod{postgresPod},
			},
			[]v1.Pod{postgresPod},
			require.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			po := Postgres{}
			got, err := po.FilterPods(context.TODO(), tt.args.client, tt.args.pods)
			tt.wantErr(t, err)

			assert.Equal(t, tt.want, got)
		})
	}
}

func TestPostgres_PasswordEnvs(t *testing.T) {
	type args struct {
		c config.Global
	}
	tests := []struct {
		name string
		args args
		want kubernetes.ConfigLookups
	}{
		{"default", args{}, kubernetes.ConfigLookups{
			kubernetes.LookupEnv{
				"POSTGRES_PASSWORD",
				"PGPOOL_POSTGRES_PASSWORD",
				"PGPASSWORD_SUPERUSER",
			},
		}},
		{"postgres", args{config.Global{Username: "postgres"}}, kubernetes.ConfigLookups{
			kubernetes.LookupEnv{
				"POSTGRES_POSTGRES_PASSWORD",
				"POSTGRES_PASSWORD",
				"PGPOOL_POSTGRES_PASSWORD",
				"PGPASSWORD_SUPERUSER",
			},
		}},
		{"cnpg", args{config.Global{DBPod: newCNPGPod()}}, kubernetes.ConfigLookups{
			kubernetes.LookupNamedSecret{
				Name: "postgresql-app",
				Key:  "password",
			},
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := Postgres{}
			got := db.PasswordEnvs(tt.args.c)
			assert.Equal(t, tt.want, got)
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
			args{config.Restore{Global: config.Global{Host: "1.1.1.1", Database: "d", Username: "u"}}, sqlformat.Gzip},
			command.NewBuilder(pgpassword, "psql", "--host=1.1.1.1", "--username=u", "--dbname=d"),
		},
		{
			"plain",
			args{config.Restore{Global: config.Global{Host: "1.1.1.1", Database: "d", Username: "u"}}, sqlformat.Plain},
			command.NewBuilder(pgpassword, "psql", "--host=1.1.1.1", "--username=u", "--dbname=d"),
		},
		{
			"custom",
			args{config.Restore{Global: config.Global{Host: "1.1.1.1", Database: "d", Username: "u"}}, sqlformat.Custom},
			command.NewBuilder(pgpassword, "pg_restore", "--format=custom", "--verbose", "--host=1.1.1.1", "--username=u", "--dbname=d"),
		},
		{
			"custom-no-owner",
			args{config.Restore{NoOwner: true, Global: config.Global{Host: "1.1.1.1", Database: "d", Username: "u"}}, sqlformat.Custom},
			command.NewBuilder(pgpassword, "pg_restore", "--format=custom", "--no-owner", "--verbose", "--host=1.1.1.1", "--username=u", "--dbname=d"),
		},
		{
			"single-transaction",
			args{config.Restore{SingleTransaction: true, Global: config.Global{Host: "1.1.1.1", Database: "d", Username: "u"}}, sqlformat.Custom},
			command.NewBuilder(pgpassword, "pg_restore", "--format=custom", "--verbose", "--host=1.1.1.1", "--username=u", "--dbname=d", "--single-transaction"),
		},
		{
			"custom clean",
			args{config.Restore{Clean: true, Global: config.Global{Host: "1.1.1.1", Database: "d", Username: "u"}}, sqlformat.Custom},
			command.NewBuilder(pgpassword, "pg_restore", "--format=custom", "--clean", "--verbose", "--host=1.1.1.1", "--username=u", "--dbname=d"),
		},
		{
			"custom halt",
			args{config.Restore{HaltOnError: true, Global: config.Global{Host: "1.1.1.1", Database: "d", Username: "u"}}, sqlformat.Custom},
			command.NewBuilder(pgpassword, "pg_restore", "--format=custom", "--exit-on-error", "--verbose", "--host=1.1.1.1", "--username=u", "--dbname=d"),
		},
		{
			"sql-quiet",
			args{config.Restore{Global: config.Global{Host: "1.1.1.1", Database: "d", Username: "u", Quiet: true}}, sqlformat.Gzip},
			command.NewBuilder(pgpassword, command.NewEnv("PGOPTIONS", "-c client_min_messages=WARNING"), "psql", "--quiet", "--output=/dev/null", "--host=1.1.1.1", "--username=u", "--dbname=d"),
		},
		{
			"port",
			args{config.Restore{Global: config.Global{Port: 1234}}, sqlformat.Plain},
			command.NewBuilder(pgpassword, "psql", "--host=", "--username=", "--dbname=", "--port=1234"),
		},
		{
			"halt_on_error",
			args{config.Restore{HaltOnError: true}, sqlformat.Plain},
			command.NewBuilder(pgpassword, "psql", "--set=ON_ERROR_STOP=1", "--host=", "--username=", "--dbname="),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			po := Postgres{}
			got := po.RestoreCommand(tt.args.conf, tt.args.inputFormat)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestPostgres_quoteParam(t *testing.T) {
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
			po := Postgres{}
			got := po.quoteParam(tt.args.param)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestPostgres_cnpgSecretName(t *testing.T) {
	type args struct {
		conf config.Global
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"default", args{config.Global{DBPod: newCNPGPod()}}, "postgresql-app"},
		{"postgres", args{config.Global{Username: "postgres", DBPod: newCNPGPod()}}, "postgresql-superuser"},
		{"other", args{}, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := Postgres{}
			assert.Equal(t, tt.want, db.cnpgSecretName(tt.args.conf))
		})
	}
}

func TestPostgres_PortEnvs(t *testing.T) {
	type args struct {
		conf config.Global
	}
	tests := []struct {
		name string
		args args
		want kubernetes.ConfigLookups
	}{
		{"default", args{}, kubernetes.ConfigLookups{kubernetes.LookupEnv{"POSTGRESQL_PORT_NUMBER"}}},
		{"cnpg", args{config.Global{DBPod: newCNPGPod()}}, kubernetes.ConfigLookups{
			kubernetes.LookupNamedSecret{
				Name: "postgresql-app",
				Key:  "port",
			},
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := Postgres{}
			assert.Equal(t, tt.want, db.PortEnvs(tt.args.conf))
		})
	}
}

func TestPostgres_DatabaseEnvs(t *testing.T) {
	type args struct {
		conf config.Global
	}
	tests := []struct {
		name string
		args args
		want kubernetes.ConfigLookups
	}{
		{"default", args{}, kubernetes.ConfigLookups{kubernetes.LookupEnv{"POSTGRES_DATABASE", "POSTGRES_DB"}}},
		{"cnpg", args{config.Global{DBPod: newCNPGPod()}}, kubernetes.ConfigLookups{
			kubernetes.LookupNamedSecret{
				Name: "postgresql-app",
				Key:  "dbname",
			},
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := Postgres{}
			assert.Equal(t, tt.want, db.DatabaseEnvs(tt.args.conf))
		})
	}
}

func TestPostgres_UserEnvs(t *testing.T) {
	type args struct {
		conf config.Global
	}
	tests := []struct {
		name string
		args args
		want kubernetes.ConfigLookups
	}{
		{"default", args{}, kubernetes.ConfigLookups{kubernetes.LookupEnv{"POSTGRES_USER", "PGPOOL_POSTGRES_USERNAME", "PGUSER_SUPERUSER"}}},
		{"cnpg", args{config.Global{DBPod: newCNPGPod()}}, kubernetes.ConfigLookups{
			kubernetes.LookupNamedSecret{
				Name: "postgresql-app",
				Key:  "username",
			},
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := Postgres{}
			assert.Equal(t, tt.want, db.UserEnvs(tt.args.conf))
		})
	}
}

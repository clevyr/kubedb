package postgres

import (
	"testing"

	"github.com/clevyr/kubedb/internal/command"
	"github.com/clevyr/kubedb/internal/config/conftypes"
	"github.com/clevyr/kubedb/internal/database/sqlformat"
	"github.com/clevyr/kubedb/internal/kubernetes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func newCNPGPod() corev1.Pod {
	return corev1.Pod{
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
		conf *conftypes.Dump
	}
	tests := []struct {
		name string
		args args
		want *command.Builder
	}{
		{
			"default",
			args{
				&conftypes.Dump{
					Global: &conftypes.Global{Host: "1.1.1.1", Database: "d", Username: "u", Password: "p"},
				},
			},
			command.NewBuilder(
				command.NewEnv("PGPASSWORD", "p"),
				"pg_dump",
				"--host=1.1.1.1",
				"--username=u",
				"--dbname=d",
				"--verbose",
			),
		},
		{
			"clean",
			args{
				&conftypes.Dump{
					Clean:  true,
					Global: &conftypes.Global{Host: "1.1.1.1", Database: "d", Username: "u", Password: "p"},
				},
			},
			command.NewBuilder(
				command.NewEnv("PGPASSWORD", "p"),
				"pg_dump",
				"--host=1.1.1.1",
				"--username=u",
				"--dbname=d",
				"--clean",
				"--verbose",
			),
		},
		{
			"no-owner",
			args{
				&conftypes.Dump{
					NoOwner: true,
					Global:  &conftypes.Global{Host: "1.1.1.1", Database: "d", Username: "u", Password: "p"},
				},
			},
			command.NewBuilder(
				command.NewEnv("PGPASSWORD", "p"),
				"pg_dump",
				"--host=1.1.1.1",
				"--username=u",
				"--dbname=d",
				"--no-owner",
				"--verbose",
			),
		},
		{
			"if-exists",
			args{
				&conftypes.Dump{
					Clean:    true,
					IfExists: true,
					Global:   &conftypes.Global{Host: "1.1.1.1", Database: "d", Username: "u", Password: "p"},
				},
			},
			command.NewBuilder(
				command.NewEnv("PGPASSWORD", "p"),
				"pg_dump",
				"--host=1.1.1.1",
				"--username=u",
				"--dbname=d",
				"--clean",
				"--if-exists",
				"--verbose",
			),
		},
		{
			"if-exists without clean",
			args{
				&conftypes.Dump{
					IfExists: true,
					Global:   &conftypes.Global{Host: "1.1.1.1", Database: "d", Username: "u", Password: "p"},
				},
			},
			command.NewBuilder(
				command.NewEnv("PGPASSWORD", "p"),
				"pg_dump",
				"--host=1.1.1.1",
				"--username=u",
				"--dbname=d",
				"--verbose",
			),
		},
		{
			"tables",
			args{
				&conftypes.Dump{
					Table:  []string{"table1", "table2"},
					Global: &conftypes.Global{Host: "1.1.1.1", Database: "d", Username: "u", Password: "p"},
				},
			},
			command.NewBuilder(
				command.NewEnv("PGPASSWORD", "p"),
				"pg_dump",
				"--host=1.1.1.1",
				"--username=u",
				"--dbname=d",
				`--table="table1"`,
				`--table="table2"`,
				"--verbose",
			),
		},
		{
			"exclude-table",
			args{
				&conftypes.Dump{
					ExcludeTable: []string{"table1", "table2"},
					Global:       &conftypes.Global{Host: "1.1.1.1", Database: "d", Username: "u", Password: "p"},
				},
			},
			command.NewBuilder(
				command.NewEnv("PGPASSWORD", "p"),
				"pg_dump",
				"--host=1.1.1.1",
				"--username=u",
				"--dbname=d",
				`--exclude-table="table1"`,
				`--exclude-table="table2"`,
				"--verbose",
			),
		},
		{
			"exclude-table-data",
			args{
				&conftypes.Dump{
					ExcludeTableData: []string{"table1", "table2"},
					Global:           &conftypes.Global{Host: "1.1.1.1", Database: "d", Username: "u", Password: "p"},
				},
			},
			command.NewBuilder(
				command.NewEnv("PGPASSWORD", "p"),
				"pg_dump",
				"--host=1.1.1.1",
				"--username=u",
				"--dbname=d",
				`--exclude-table-data="table1"`,
				`--exclude-table-data="table2"`,
				"--verbose",
			),
		},
		{
			"custom",
			args{
				&conftypes.Dump{
					Format: sqlformat.Custom,
					Global: &conftypes.Global{Host: "1.1.1.1", Database: "d", Username: "u", Password: "p"},
				},
			},
			command.NewBuilder(
				command.NewEnv("PGPASSWORD", "p"),
				"pg_dump",
				"--host=1.1.1.1",
				"--username=u",
				"--dbname=d",
				"--format=custom",
				"--verbose",
			),
		},
		{
			"port",
			args{&conftypes.Dump{Global: &conftypes.Global{Port: 1234}}},
			command.NewBuilder("pg_dump", "--host=", "--username=", "--port=1234", "--verbose"),
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
		conf *conftypes.Exec
	}
	tests := []struct {
		name string
		args args
		want *command.Builder
	}{
		{
			"default",
			args{
				&conftypes.Exec{
					Global: &conftypes.Global{Host: "1.1.1.1", Database: "d", Username: "u", Password: "p"},
				},
			},
			command.NewBuilder(
				command.NewEnv("PGPASSWORD", "p"),
				"exec",
				"psql",
				"--host=1.1.1.1",
				"--username=u",
				"--dbname=d",
			),
		},
		{
			"disable-headers",
			args{
				&conftypes.Exec{
					DisableHeaders: true,
					Global:         &conftypes.Global{Host: "1.1.1.1", Database: "d", Username: "u", Password: "p"},
				},
			},
			command.NewBuilder(
				command.NewEnv("PGPASSWORD", "p"),
				"exec",
				"psql",
				"--host=1.1.1.1",
				"--username=u",
				"--dbname=d",
				"--tuples-only",
			),
		},
		{
			"command",
			args{
				&conftypes.Exec{
					Command: "select true",
					Global:  &conftypes.Global{Host: "1.1.1.1", Database: "d", Username: "u", Password: "p"},
				},
			},
			command.NewBuilder(
				command.NewEnv("PGPASSWORD", "p"),
				"exec",
				"psql",
				"--host=1.1.1.1",
				"--username=u",
				"--dbname=d",
				"--command=select true",
			),
		},
		{
			"default",
			args{&conftypes.Exec{Global: &conftypes.Global{Port: 1234}}},
			command.NewBuilder("exec", "psql", "--host=", "--username=", "--port=1234"),
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
	postgresPod := corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{
				"app.kubernetes.io/name":      "postgresql",
				"app.kubernetes.io/component": "primary",
			},
		},
	}

	type args struct {
		client kubernetes.KubeClient
		pods   []corev1.Pod
	}
	tests := []struct {
		name    string
		args    args
		want    []corev1.Pod
		wantErr require.ErrorAssertionFunc
	}{
		{
			"postgresql",
			args{
				kubernetes.KubeClient{},
				[]corev1.Pod{postgresPod},
			},
			[]corev1.Pod{postgresPod},
			require.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			po := Postgres{}
			got, err := po.FilterPods(t.Context(), tt.args.client, tt.args.pods)
			tt.wantErr(t, err)

			assert.Equal(t, tt.want, got)
		})
	}
}

func TestPostgres_PasswordEnvs(t *testing.T) {
	type args struct {
		c *conftypes.Global
	}
	tests := []struct {
		name string
		args args
		want kubernetes.ConfigLookups
	}{
		{"default", args{&conftypes.Global{}}, kubernetes.ConfigLookups{
			kubernetes.LookupEnv{"POSTGRES_PASSWORD", "PGPOOL_POSTGRES_PASSWORD", "PGPASSWORD_SUPERUSER"},
			kubernetes.LookupSecretVolume{Name: "password", Key: "password"},
			kubernetes.LookupSecretVolume{Name: "postgresql-password", Key: "password"},
		}},
		{"postgres", args{&conftypes.Global{Username: "postgres"}}, kubernetes.ConfigLookups{
			kubernetes.LookupSecretVolume{Name: "postgresql-password", Key: "postgres-password"},
			kubernetes.LookupEnv{
				"POSTGRES_POSTGRES_PASSWORD",
				"POSTGRES_PASSWORD",
				"PGPOOL_POSTGRES_PASSWORD",
				"PGPASSWORD_SUPERUSER",
			},
			kubernetes.LookupSecretVolume{Name: "password", Key: "password"},
			kubernetes.LookupSecretVolume{Name: "postgresql-password", Key: "password"},
		}},
		{"cnpg", args{&conftypes.Global{DBPod: newCNPGPod()}}, kubernetes.ConfigLookups{
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
	pgpassword := command.NewEnv("PGPASSWORD", "p")

	type args struct {
		conf        *conftypes.Restore
		inputFormat sqlformat.Format
	}
	tests := []struct {
		name string
		args args
		want *command.Builder
	}{
		{
			"gzip",
			args{
				&conftypes.Restore{
					Global: &conftypes.Global{Host: "1.1.1.1", Database: "d", Username: "u", Password: "p"},
				},
				sqlformat.Gzip,
			},
			command.NewBuilder(pgpassword, "psql", "--host=1.1.1.1", "--username=u", "--dbname=d"),
		},
		{
			"plain",
			args{
				&conftypes.Restore{
					Global: &conftypes.Global{Host: "1.1.1.1", Database: "d", Username: "u", Password: "p"},
				},
				sqlformat.Plain,
			},
			command.NewBuilder(pgpassword, "psql", "--host=1.1.1.1", "--username=u", "--dbname=d"),
		},
		{
			"custom",
			args{
				&conftypes.Restore{
					Global: &conftypes.Global{Host: "1.1.1.1", Database: "d", Username: "u", Password: "p"},
				},
				sqlformat.Custom,
			},
			command.NewBuilder(
				pgpassword,
				"pg_restore",
				"--format=custom",
				"--verbose",
				"--host=1.1.1.1",
				"--username=u",
				"--dbname=d",
			),
		},
		{
			"custom-no-owner",
			args{
				&conftypes.Restore{
					NoOwner: true,
					Global:  &conftypes.Global{Host: "1.1.1.1", Database: "d", Username: "u", Password: "p"},
				},
				sqlformat.Custom,
			},
			command.NewBuilder(
				pgpassword,
				"pg_restore",
				"--format=custom",
				"--no-owner",
				"--verbose",
				"--host=1.1.1.1",
				"--username=u",
				"--dbname=d",
			),
		},
		{
			"single-transaction",
			args{
				&conftypes.Restore{
					SingleTransaction: true,
					Global:            &conftypes.Global{Host: "1.1.1.1", Database: "d", Username: "u", Password: "p"},
				},
				sqlformat.Custom,
			},
			command.NewBuilder(
				pgpassword,
				"pg_restore",
				"--format=custom",
				"--verbose",
				"--host=1.1.1.1",
				"--username=u",
				"--dbname=d",
				"--single-transaction",
			),
		},
		{
			"custom clean",
			args{
				&conftypes.Restore{
					Clean:  true,
					Global: &conftypes.Global{Host: "1.1.1.1", Database: "d", Username: "u", Password: "p"},
				},
				sqlformat.Custom,
			},
			command.NewBuilder(
				pgpassword,
				"pg_restore",
				"--format=custom",
				"--clean",
				"--verbose",
				"--host=1.1.1.1",
				"--username=u",
				"--dbname=d",
			),
		},
		{
			"custom halt",
			args{
				&conftypes.Restore{
					HaltOnError: true,
					Global:      &conftypes.Global{Host: "1.1.1.1", Database: "d", Username: "u", Password: "p"},
				},
				sqlformat.Custom,
			},
			command.NewBuilder(
				pgpassword,
				"pg_restore",
				"--format=custom",
				"--exit-on-error",
				"--verbose",
				"--host=1.1.1.1",
				"--username=u",
				"--dbname=d",
			),
		},
		{
			"sql-quiet",
			args{
				&conftypes.Restore{
					Global: &conftypes.Global{
						Host:     "1.1.1.1",
						Database: "d",
						Username: "u",
						Password: "p",
						Quiet:    true,
					},
				},
				sqlformat.Gzip,
			},
			command.NewBuilder(
				pgpassword,
				command.NewEnv("PGOPTIONS", "-c client_min_messages=WARNING"),
				"psql",
				"--quiet",
				"--output=/dev/null",
				"--host=1.1.1.1",
				"--username=u",
				"--dbname=d",
			),
		},
		{
			"port",
			args{&conftypes.Restore{Global: &conftypes.Global{Port: 1234}}, sqlformat.Plain},
			command.NewBuilder("psql", "--host=", "--username=", "--dbname=", "--port=1234"),
		},
		{
			"halt_on_error",
			args{&conftypes.Restore{HaltOnError: true, Global: &conftypes.Global{}}, sqlformat.Plain},
			command.NewBuilder("psql", "--set=ON_ERROR_STOP=1", "--host=", "--username=", "--dbname="),
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
		conf *conftypes.Global
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"default", args{&conftypes.Global{DBPod: newCNPGPod()}}, "postgresql-app"},
		{"postgres", args{&conftypes.Global{Username: "postgres", DBPod: newCNPGPod()}}, "postgresql-superuser"},
		{"other", args{&conftypes.Global{}}, ""},
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
		conf *conftypes.Global
	}
	tests := []struct {
		name string
		args args
		want kubernetes.ConfigLookups
	}{
		{
			"default",
			args{&conftypes.Global{}},
			kubernetes.ConfigLookups{kubernetes.LookupEnv{"POSTGRESQL_PORT_NUMBER"}},
		},
		{"cnpg", args{&conftypes.Global{DBPod: newCNPGPod()}}, kubernetes.ConfigLookups{
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
		conf *conftypes.Global
	}
	tests := []struct {
		name string
		args args
		want kubernetes.ConfigLookups
	}{
		{
			"default",
			args{&conftypes.Global{}},
			kubernetes.ConfigLookups{
				kubernetes.LookupEnv{"POSTGRES_DATABASE", "POSTGRES_DB"},
				kubernetes.LookupDefault("postgres"),
			},
		},
		{"cnpg", args{&conftypes.Global{DBPod: newCNPGPod()}}, kubernetes.ConfigLookups{
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
		conf *conftypes.Global
	}
	tests := []struct {
		name string
		args args
		want kubernetes.ConfigLookups
	}{
		{
			"default",
			args{&conftypes.Global{}},
			kubernetes.ConfigLookups{
				kubernetes.LookupEnv{"POSTGRES_USER", "PGPOOL_POSTGRES_USERNAME", "PGUSER_SUPERUSER"},
			},
		},
		{"cnpg", args{&conftypes.Global{DBPod: newCNPGPod()}}, kubernetes.ConfigLookups{
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

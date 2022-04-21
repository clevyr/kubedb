package dialect

import (
	"github.com/clevyr/kubedb/internal/command"
	"github.com/clevyr/kubedb/internal/config"
	"github.com/clevyr/kubedb/internal/database/sqlformat"
	"github.com/clevyr/kubedb/internal/kubernetes"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"reflect"
	"testing"
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
			if got := po.AnalyzeQuery(); got != tt.want {
				t.Errorf("AnalyzeQuery() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPostgres_DatabaseEnvNames(t *testing.T) {
	tests := []struct {
		name string
		want []string
	}{
		{"default", []string{"POSTGRES_DB"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			po := Postgres{}
			if got := po.DatabaseEnvNames(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DatabaseEnvNames() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPostgres_DefaultDatabase(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{"default", "db"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			po := Postgres{}
			if got := po.DefaultDatabase(); got != tt.want {
				t.Errorf("DefaultDatabase() = %v, want %v", got, tt.want)
			}
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
			if got := po.DefaultPort(); got != tt.want {
				t.Errorf("DefaultPort() = %v, want %v", got, tt.want)
			}
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
			if got := po.DefaultUser(); got != tt.want {
				t.Errorf("DefaultUser() = %v, want %v", got, tt.want)
			}
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
			if got := po.DropDatabaseQuery(tt.args.database); got != tt.want {
				t.Errorf("DropDatabaseQuery() = %v, want %v", got, tt.want)
			}
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
			command.NewBuilder(command.NewEnv("PGPASSWORD", ""), "pg_dump", "--host=127.0.0.1", "--username=u", "--dbname=d"),
		},
		{
			"clean",
			args{config.Dump{Clean: true, Global: config.Global{Database: "d", Username: "u"}}},
			command.NewBuilder(command.NewEnv("PGPASSWORD", ""), "pg_dump", "--host=127.0.0.1", "--username=u", "--dbname=d", "--clean"),
		},
		{
			"no-owner",
			args{config.Dump{NoOwner: true, Global: config.Global{Database: "d", Username: "u"}}},
			command.NewBuilder(command.NewEnv("PGPASSWORD", ""), "pg_dump", "--host=127.0.0.1", "--username=u", "--dbname=d", "--no-owner"),
		},
		{
			"if-exists",
			args{config.Dump{IfExists: true, Global: config.Global{Database: "d", Username: "u"}}},
			command.NewBuilder(command.NewEnv("PGPASSWORD", ""), "pg_dump", "--host=127.0.0.1", "--username=u", "--dbname=d", "--if-exists"),
		},
		{
			"tables",
			args{config.Dump{Tables: []string{"table1", "table2"}, Global: config.Global{Database: "d", Username: "u"}}},
			command.NewBuilder(command.NewEnv("PGPASSWORD", ""), "pg_dump", "--host=127.0.0.1", "--username=u", "--dbname=d", `--table="table1"`, `--table="table2"`),
		},
		{
			"exclude-table",
			args{config.Dump{ExcludeTable: []string{"table1", "table2"}, Global: config.Global{Database: "d", Username: "u"}}},
			command.NewBuilder(command.NewEnv("PGPASSWORD", ""), "pg_dump", "--host=127.0.0.1", "--username=u", "--dbname=d", `--exclude-table="table1"`, `--exclude-table="table2"`),
		},
		{
			"exclude-table-data",
			args{config.Dump{ExcludeTableData: []string{"table1", "table2"}, Global: config.Global{Database: "d", Username: "u"}}},
			command.NewBuilder(command.NewEnv("PGPASSWORD", ""), "pg_dump", "--host=127.0.0.1", "--username=u", "--dbname=d", `--exclude-table-data="table1"`, `--exclude-table-data="table2"`),
		},
		{
			"custom",
			args{config.Dump{Files: config.Files{Format: sqlformat.Custom}, Global: config.Global{Database: "d", Username: "u"}}},
			command.NewBuilder(command.NewEnv("PGPASSWORD", ""), "pg_dump", "--host=127.0.0.1", "--username=u", "--dbname=d", "--format=c"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			po := Postgres{}
			if got := po.DumpCommand(tt.args.conf); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DumpCommand() = %v, want %v", got, tt.want)
			}
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			po := Postgres{}
			if got := po.ExecCommand(tt.args.conf); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ExecCommand() = %v, want %v", got, tt.want)
			}
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
			got, err := po.FilterPods(tt.args.client, tt.args.pods)
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
			if got := po.ListDatabasesQuery(); got != tt.want {
				t.Errorf("ListDatabasesQuery() = %v, want %v", got, tt.want)
			}
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
			if got := po.ListTablesQuery(); got != tt.want {
				t.Errorf("ListTablesQuery() = %v, want %v", got, tt.want)
			}
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
			if got := po.Name(); got != tt.want {
				t.Errorf("Name() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPostgres_PasswordEnvNames(t *testing.T) {
	tests := []struct {
		name string
		want []string
	}{
		{"default", []string{"POSTGRES_PASSWORD", "PGPOOL_POSTGRES_PASSWORD"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			po := Postgres{}
			if got := po.PasswordEnvNames(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("PasswordEnvNames() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPostgres_PodLabels(t *testing.T) {
	tests := []struct {
		name string
		want []kubernetes.LabelQueryable
	}{
		{"default", []kubernetes.LabelQueryable{
			kubernetes.LabelQuery{Name: "app", Value: "postgresql"},
			kubernetes.LabelQueryAnd{
				{Name: "app.kubernetes.io/name", Value: "postgresql"},
				{Name: "app.kubernetes.io/component", Value: "primary"},
			},
			kubernetes.LabelQueryAnd{
				{Name: "app.kubernetes.io/name", Value: "postgresql-ha"},
				{Name: "app.kubernetes.io/component", Value: "postgresql"},
			},
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			po := Postgres{}
			if got := po.PodLabels(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("PodLabels() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPostgres_RestoreCommand(t *testing.T) {
	pgpassword := command.NewEnv("PGPASSWORD", "")
	pgoptions := command.NewEnv("PGOPTIONS", "-c client_min_messages=WARNING")

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
			command.NewBuilder(pgpassword, pgoptions, "psql", "--quiet", "--output=/dev/null", "--host=127.0.0.1", "--username=u", "--dbname=d"),
		},
		{
			"plain",
			args{config.Restore{Global: config.Global{Database: "d", Username: "u"}}, sqlformat.Plain},
			command.NewBuilder(pgpassword, pgoptions, "psql", "--quiet", "--output=/dev/null", "--host=127.0.0.1", "--username=u", "--dbname=d"),
		},
		{
			"custom",
			args{config.Restore{Global: config.Global{Database: "d", Username: "u"}}, sqlformat.Custom},
			command.NewBuilder(pgpassword, pgoptions, "pg_restore", "--format=custom", "--verbose", "--clean", "--exit-on-error", "--host=127.0.0.1", "--username=u", "--dbname=d"),
		},
		{
			"custom-no-owner",
			args{config.Restore{NoOwner: true, Global: config.Global{Database: "d", Username: "u"}}, sqlformat.Custom},
			command.NewBuilder(pgpassword, pgoptions, "pg_restore", "--format=custom", "--verbose", "--clean", "--exit-on-error", "--no-owner", "--host=127.0.0.1", "--username=u", "--dbname=d"),
		},
		{
			"single-transaction",
			args{config.Restore{SingleTransaction: true, Global: config.Global{Database: "d", Username: "u"}}, sqlformat.Custom},
			command.NewBuilder(pgpassword, pgoptions, "pg_restore", "--format=custom", "--verbose", "--clean", "--exit-on-error", "--host=127.0.0.1", "--username=u", "--dbname=d", "--single-transaction"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			po := Postgres{}
			if got := po.RestoreCommand(tt.args.conf, tt.args.inputFormat); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("RestoreCommand() = %v, want %v", got, tt.want)
			}
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
			if got := po.UserEnvNames(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("UserEnvNames() = %v, want %v", got, tt.want)
			}
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
			if got := quoteParam(tt.args.param); got != tt.want {
				t.Errorf("quoteParam() = %v, want %v", got, tt.want)
			}
		})
	}
}

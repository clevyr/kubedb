package dump

import (
	"github.com/clevyr/kubedb/internal/command"
	"github.com/clevyr/kubedb/internal/config"
	"github.com/clevyr/kubedb/internal/database/dialect"
	"github.com/clevyr/kubedb/internal/database/sqlformat"
	"github.com/spf13/cobra"
	"reflect"
	"testing"
)

func Test_buildCommand(t *testing.T) {
	pgpassword := command.NewEnv("PGPASSWORD", "")
	mysql_pwd := command.NewEnv("MYSQL_PWD", "")

	type args struct {
		db   config.Databaser
		conf config.Dump
	}
	tests := []struct {
		name string
		args args
		want *command.Builder
	}{
		{
			"postgres-gzip",
			args{dialect.Postgres{}, config.Dump{Global: config.Global{Database: "d", Username: "u"}}},
			command.NewBuilder(pgpassword, "pg_dump", "--host=127.0.0.1", "--username=u", "--dbname=d", command.Pipe, "gzip", "--force"),
		},
		{
			"postgres-plain",
			args{dialect.Postgres{}, config.Dump{Files: config.Files{Format: sqlformat.Plain}, Global: config.Global{Database: "d", Username: "u"}}},
			command.NewBuilder(pgpassword, "pg_dump", "--host=127.0.0.1", "--username=u", "--dbname=d", command.Pipe, "gzip", "--force"),
		},
		{
			"postgres-custom",
			args{dialect.Postgres{}, config.Dump{Files: config.Files{Format: sqlformat.Custom}, Global: config.Global{Database: "d", Username: "u"}}},
			command.NewBuilder(pgpassword, "pg_dump", "--host=127.0.0.1", "--username=u", "--dbname=d", "--format=c"),
		},
		{
			"mariadb-gzip",
			args{dialect.MariaDB{}, config.Dump{Files: config.Files{Format: sqlformat.Gzip}, Global: config.Global{Database: "d", Username: "u"}}},
			command.NewBuilder(mysql_pwd, "mysqldump", "--host=127.0.0.1", "--user=u", "d", command.Pipe, "gzip", "--force"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := buildCommand(tt.args.db, tt.args.conf); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("buildCommand() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_validArgs(t *testing.T) {
	type args struct {
		cmd        *cobra.Command
		args       []string
		toComplete string
	}
	tests := []struct {
		name  string
		args  args
		want  []string
		want1 cobra.ShellCompDirective
	}{
		{"0 arg", args{}, []string{"sql", "sql.gz", "dmp"}, cobra.ShellCompDirectiveFilterFileExt},
		{"1 arg", args{args: []string{"sql.gz"}}, nil, cobra.ShellCompDirectiveNoFileComp},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := validArgs(tt.args.cmd, tt.args.args, tt.args.toComplete)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("validArgs() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("validArgs() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

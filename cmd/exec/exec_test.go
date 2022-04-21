package exec

import (
	"github.com/clevyr/kubedb/internal/command"
	"github.com/clevyr/kubedb/internal/config"
	"github.com/clevyr/kubedb/internal/database/dialect"
	"reflect"
	"testing"
)

func Test_buildCommand(t *testing.T) {
	pgpassword := command.NewEnv("PGPASSWORD", "")
	mysql_pwd := command.NewEnv("MYSQL_PWD", "")

	type args struct {
		db   config.Databaser
		conf config.Exec
		args []string
	}
	tests := []struct {
		name string
		args args
		want *command.Builder
	}{
		{
			"command",
			args{dialect.Postgres{}, config.Exec{Global: config.Global{Database: "d", Username: "u"}}, []string{"whoami"}},
			command.NewBuilder("exec", "whoami"),
		},
		{
			"postgres",
			args{dialect.Postgres{}, config.Exec{Global: config.Global{Database: "d", Username: "u"}}, []string{}},
			command.NewBuilder(pgpassword, "psql", "--host=127.0.0.1", "--username=u", "--dbname=d"),
		},
		{
			"mariadb-gzip",
			args{dialect.MariaDB{}, config.Exec{Global: config.Global{Database: "d", Username: "u"}}, []string{}},
			command.NewBuilder(mysql_pwd, "mysql", "--host=127.0.0.1", "--user=u", "--database=d"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := buildCommand(tt.args.db, tt.args.conf, tt.args.args); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("buildCommand() = %v, want %v", got, tt.want)
			}
		})
	}
}

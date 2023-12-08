package dump

import (
	"testing"

	"github.com/clevyr/kubedb/internal/command"
	"github.com/clevyr/kubedb/internal/config"
	"github.com/clevyr/kubedb/internal/database/mariadb"
	"github.com/clevyr/kubedb/internal/database/postgres"
	"github.com/clevyr/kubedb/internal/database/sqlformat"
	"github.com/stretchr/testify/assert"
)

func Test_buildCommand(t *testing.T) {
	pgpassword := command.NewEnv("PGPASSWORD", "")
	mysql_pwd := command.NewEnv("MYSQL_PWD", "")

	type args struct {
		conf Dump
	}
	tests := []struct {
		name string
		args args
		want *command.Builder
	}{
		{
			"postgres-gzip",
			args{Dump{Dump: config.Dump{Global: config.Global{Dialect: postgres.Postgres{}, Host: "1.1.1.1", Database: "d", Username: "u", RemoteGzip: true}}}},
			command.NewBuilder(pgpassword, "pg_dump", "--host=1.1.1.1", "--username=u", "--dbname=d", "--verbose", command.Pipe, "gzip", "--force"),
		},
		{
			"postgres-gzip-no-compression",
			args{Dump{Dump: config.Dump{Global: config.Global{Dialect: postgres.Postgres{}, Host: "1.1.1.1", Database: "d", Username: "u"}}}},
			command.NewBuilder(pgpassword, "pg_dump", "--host=1.1.1.1", "--username=u", "--dbname=d", "--verbose"),
		},
		{
			"postgres-plain",
			args{Dump{Dump: config.Dump{Files: config.Files{Format: sqlformat.Plain}, Global: config.Global{Dialect: postgres.Postgres{}, Host: "1.1.1.1", Database: "d", Username: "u", RemoteGzip: true}}}},
			command.NewBuilder(pgpassword, "pg_dump", "--host=1.1.1.1", "--username=u", "--dbname=d", "--verbose", command.Pipe, "gzip", "--force"),
		},
		{
			"postgres-custom",
			args{Dump{Dump: config.Dump{Files: config.Files{Format: sqlformat.Custom}, Global: config.Global{Dialect: postgres.Postgres{}, Host: "1.1.1.1", Database: "d", Username: "u", RemoteGzip: true}}}},
			command.NewBuilder(pgpassword, "pg_dump", "--host=1.1.1.1", "--username=u", "--dbname=d", "--format=c", "--verbose"),
		},
		{
			"mariadb-gzip",
			args{Dump{Dump: config.Dump{Files: config.Files{Format: sqlformat.Gzip}, Global: config.Global{Dialect: mariadb.MariaDB{}, Host: "1.1.1.1", Database: "d", Username: "u", RemoteGzip: true}}}},
			command.NewBuilder(mysql_pwd, "mysqldump", "--host=1.1.1.1", "--user=u", "d", "--verbose", command.Pipe, "gzip", "--force"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.args.conf.buildCommand()
			assert.Equal(t, tt.want, got)
		})
	}
}

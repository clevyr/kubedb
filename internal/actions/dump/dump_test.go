package dump

import (
	"testing"

	"github.com/clevyr/kubedb/internal/command"
	"github.com/clevyr/kubedb/internal/config/conftypes"
	"github.com/clevyr/kubedb/internal/database/mariadb"
	"github.com/clevyr/kubedb/internal/database/postgres"
	"github.com/clevyr/kubedb/internal/database/sqlformat"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_buildCommand(t *testing.T) {
	type args struct {
		conf Dump
	}
	tests := []struct {
		name    string
		args    args
		want    *command.Builder
		wantErr require.ErrorAssertionFunc
	}{
		{
			"postgres-gzip",
			args{Dump{Dump: conftypes.Dump{Global: &conftypes.Global{Dialect: postgres.Postgres{}, Host: "1.1.1.1", Database: "d", Username: "u", RemoteGzip: true}}}},
			command.NewBuilder(command.Raw("{"), "pg_dump", "--host=1.1.1.1", "--username=u", "--dbname=d", "--verbose", command.Raw("|| kill $$; }"), command.Pipe, "gzip", "--force"),
			require.NoError,
		},
		{
			"postgres-gzip-no-compression",
			args{Dump{Dump: conftypes.Dump{Global: &conftypes.Global{Dialect: postgres.Postgres{}, Host: "1.1.1.1", Database: "d", Username: "u"}}}},
			command.NewBuilder(command.Raw("{"), "pg_dump", "--host=1.1.1.1", "--username=u", "--dbname=d", "--verbose", command.Raw("|| kill $$; }")),
			require.NoError,
		},
		{
			"postgres-plain",
			args{Dump{Dump: conftypes.Dump{Format: sqlformat.Plain, Global: &conftypes.Global{Dialect: postgres.Postgres{}, Host: "1.1.1.1", Database: "d", Username: "u", RemoteGzip: true}}}},
			command.NewBuilder(command.Raw("{"), "pg_dump", "--host=1.1.1.1", "--username=u", "--dbname=d", "--verbose", command.Raw("|| kill $$; }"), command.Pipe, "gzip", "--force"),
			require.NoError,
		},
		{
			"postgres-custom",
			args{Dump{Dump: conftypes.Dump{Format: sqlformat.Custom, Global: &conftypes.Global{Dialect: postgres.Postgres{}, Host: "1.1.1.1", Database: "d", Username: "u", RemoteGzip: true}}}},
			command.NewBuilder(command.Raw("{"), "pg_dump", "--host=1.1.1.1", "--username=u", "--dbname=d", "--format=custom", "--verbose", command.Raw("|| kill $$; }")),
			require.NoError,
		},
		{
			"mariadb-gzip",
			args{Dump{Dump: conftypes.Dump{Format: sqlformat.Gzip, Global: &conftypes.Global{Dialect: mariadb.MariaDB{}, Host: "1.1.1.1", Database: "d", Username: "u", RemoteGzip: true}}}},
			command.NewBuilder(command.Raw("{"), command.Raw(`"$(which mariadb-dump || which mysqldump)"`), "--host=1.1.1.1", "--user=u", "d", "--verbose", command.Raw("|| kill $$; }"), command.Pipe, "gzip", "--force"),
			require.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.args.conf.buildCommand()
			tt.wantErr(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

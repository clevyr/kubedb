package restore

import (
	"bytes"
	"compress/gzip"
	"io"
	"strings"
	"testing"

	"github.com/clevyr/kubedb/internal/command"
	"github.com/clevyr/kubedb/internal/config/conftypes"
	"github.com/clevyr/kubedb/internal/database/postgres"
	"github.com/clevyr/kubedb/internal/database/sqlformat"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRestore_buildCommand(t *testing.T) {
	type fields struct {
		Restore conftypes.Restore
		Analyze bool
	}
	type args struct {
		inputFormat sqlformat.Format
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *command.Builder
		wantErr require.ErrorAssertionFunc
	}{
		{
			"postgres-gzip",
			fields{Restore: conftypes.Restore{Global: &conftypes.Global{Dialect: postgres.Postgres{}, Host: "1.1.1.1", Database: "d", Username: "u", RemoteGzip: true}}},
			args{sqlformat.Gzip},
			command.NewBuilder("gunzip", "--force", command.Pipe, command.Raw("{"), "psql", "--host=1.1.1.1", "--username=u", "--dbname=d", command.Raw("|| { cat >/dev/null; kill $$; }; }")),
			require.NoError,
		},
		{
			"postgres-plain",
			fields{Restore: conftypes.Restore{Global: &conftypes.Global{Dialect: postgres.Postgres{}, Host: "1.1.1.1", Database: "d", Username: "u", RemoteGzip: true}}},
			args{sqlformat.Gzip},
			command.NewBuilder("gunzip", "--force", command.Pipe, command.Raw("{"), "psql", "--host=1.1.1.1", "--username=u", "--dbname=d", command.Raw("|| { cat >/dev/null; kill $$; }; }")),
			require.NoError,
		},
		{
			"postgres-custom",
			fields{Restore: conftypes.Restore{Global: &conftypes.Global{Dialect: postgres.Postgres{}, Host: "1.1.1.1", Database: "d", Username: "u", RemoteGzip: true}}},
			args{sqlformat.Gzip},
			command.NewBuilder("gunzip", "--force", command.Pipe, command.Raw("{"), "psql", "--host=1.1.1.1", "--username=u", "--dbname=d", command.Raw("|| { cat >/dev/null; kill $$; }; }")),
			require.NoError,
		},
		{
			"postgres-remote-gzip-disabled",
			fields{Restore: conftypes.Restore{Global: &conftypes.Global{Dialect: postgres.Postgres{}, Host: "1.1.1.1", Database: "d", Username: "u"}}},
			args{sqlformat.Gzip},
			command.NewBuilder(command.Raw("{"), "psql", "--host=1.1.1.1", "--username=u", "--dbname=d", command.Raw("|| { cat >/dev/null; kill $$; }; }")),
			require.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			action := Restore{
				Restore: tt.fields.Restore,
				Analyze: tt.fields.Analyze,
			}
			cmd, err := action.buildCommand(tt.args.inputFormat)
			tt.wantErr(t, err)
			assert.Equal(t, tt.want, cmd)
		})
	}
}

func TestRestore_copy(t *testing.T) {
	input := "hello world"
	var gzipped strings.Builder
	gzw := gzip.NewWriter(&gzipped)
	_, err := gzw.Write([]byte(input))
	require.NoError(t, err)
	require.NoError(t, gzw.Close())

	type fields struct {
		Restore conftypes.Restore
		Analyze bool
	}
	type args struct {
		r io.Reader
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    string
		wantErr require.ErrorAssertionFunc
	}{
		{"gzip", fields{Restore: conftypes.Restore{Global: &conftypes.Global{RemoteGzip: true}}}, args{strings.NewReader(input)}, gzipped.String(), require.NoError},
		{"raw", fields{Restore: conftypes.Restore{Global: &conftypes.Global{}}}, args{strings.NewReader(input)}, "hello world", require.NoError},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			action := Restore{
				Restore: tt.fields.Restore,
				Analyze: tt.fields.Analyze,
			}
			w := &bytes.Buffer{}
			_, err := action.copy(w, tt.args.r)
			tt.wantErr(t, err)
			assert.Equal(t, tt.want, w.String())
		})
	}
}

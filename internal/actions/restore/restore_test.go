package restore

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/clevyr/kubedb/internal/command"
	"github.com/clevyr/kubedb/internal/config"
	"github.com/clevyr/kubedb/internal/database/postgres"
	"github.com/clevyr/kubedb/internal/database/sqlformat"
	gzip "github.com/klauspost/pgzip"
	"github.com/stretchr/testify/assert"
)

func TestRestore_buildCommand(t *testing.T) {
	pgpassword := command.NewEnv("PGPASSWORD", "")

	type fields struct {
		Restore config.Restore
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
		wantErr assert.ErrorAssertionFunc
	}{
		{
			"postgres-gzip",
			fields{Restore: config.Restore{Global: config.Global{Dialect: postgres.Postgres{}, Host: "1.1.1.1", Database: "d", Username: "u", RemoteGzip: true}}},
			args{sqlformat.Gzip},
			command.NewBuilder("gunzip", "--force", command.Pipe, pgpassword, "psql", "--host=1.1.1.1", "--username=u", "--dbname=d"),
			assert.NoError,
		},
		{
			"postgres-plain",
			fields{Restore: config.Restore{Global: config.Global{Dialect: postgres.Postgres{}, Host: "1.1.1.1", Database: "d", Username: "u", RemoteGzip: true}}},
			args{sqlformat.Gzip},
			command.NewBuilder("gunzip", "--force", command.Pipe, pgpassword, "psql", "--host=1.1.1.1", "--username=u", "--dbname=d"),
			assert.NoError,
		},
		{
			"postgres-custom",
			fields{Restore: config.Restore{Global: config.Global{Dialect: postgres.Postgres{}, Host: "1.1.1.1", Database: "d", Username: "u", RemoteGzip: true}}},
			args{sqlformat.Gzip},
			command.NewBuilder("gunzip", "--force", command.Pipe, pgpassword, "psql", "--host=1.1.1.1", "--username=u", "--dbname=d"),
			assert.NoError,
		},
		{
			"postgres-remote-gzip-disabled",
			fields{Restore: config.Restore{Global: config.Global{Dialect: postgres.Postgres{}, Host: "1.1.1.1", Database: "d", Username: "u"}}},
			args{sqlformat.Gzip},
			command.NewBuilder(command.Env{Key: "PGPASSWORD", Value: ""}, "psql", "--host=1.1.1.1", "--username=u", "--dbname=d"),
			assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			action := Restore{
				Restore: tt.fields.Restore,
				Analyze: tt.fields.Analyze,
			}
			cmd, err := action.buildCommand(tt.args.inputFormat)
			if !tt.wantErr(t, err) {
				return
			}
			assert.Equal(t, tt.want, cmd)
		})
	}
}

func Test_gzipCopy(t *testing.T) {
	input := "hello world"
	var output strings.Builder
	gzw := gzip.NewWriter(&output)
	if _, err := gzw.Write([]byte(input)); err != nil {
		panic(err)
	}
	if err := gzw.Close(); err != nil {
		panic(err)
	}

	type args struct {
		r io.Reader
	}
	tests := []struct {
		name    string
		args    args
		wantW   string
		wantErr bool
	}{
		{"", args{strings.NewReader(input)}, output.String(), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &bytes.Buffer{}
			err := gzipCopy(w, tt.args.r)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.wantW, w.String())
		})
	}
}

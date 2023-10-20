package restore

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/clevyr/kubedb/internal/command"
	"github.com/clevyr/kubedb/internal/config"
	"github.com/clevyr/kubedb/internal/database/dialect"
	"github.com/clevyr/kubedb/internal/database/sqlformat"
	gzip "github.com/klauspost/pgzip"
	"github.com/stretchr/testify/assert"
)

func Test_buildCommand(t *testing.T) {
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
			"postgres-gzip",
			args{
				config.Restore{Global: config.Global{Dialect: dialect.Postgres{}, Host: "1.1.1.1", Database: "d", Username: "u", RemoteGzip: true}},
				sqlformat.Gzip,
			},
			command.NewBuilder("gunzip", "--force", command.Pipe, pgpassword, "psql", "--host=1.1.1.1", "--username=u", "--dbname=d"),
		},
		{
			"postgres-plain",
			args{
				config.Restore{Global: config.Global{Dialect: dialect.Postgres{}, Host: "1.1.1.1", Database: "d", Username: "u", RemoteGzip: true}},
				sqlformat.Gzip,
			},
			command.NewBuilder("gunzip", "--force", command.Pipe, pgpassword, "psql", "--host=1.1.1.1", "--username=u", "--dbname=d"),
		},
		{
			"postgres-custom",
			args{
				config.Restore{Global: config.Global{Dialect: dialect.Postgres{}, Host: "1.1.1.1", Database: "d", Username: "u", RemoteGzip: true}},
				sqlformat.Gzip,
			},
			command.NewBuilder("gunzip", "--force", command.Pipe, pgpassword, "psql", "--host=1.1.1.1", "--username=u", "--dbname=d"),
		},
		{
			"postgres-remote-gzip-disabled",
			args{
				config.Restore{Global: config.Global{Dialect: dialect.Postgres{}, Host: "1.1.1.1", Database: "d", Username: "u"}},
				sqlformat.Gzip,
			},
			command.NewBuilder(command.Env{Key: "PGPASSWORD", Value: ""}, "psql", "--host=1.1.1.1", "--username=u", "--dbname=d"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildCommand(tt.args.conf, tt.args.inputFormat)
			assert.Equal(t, tt.want, got)
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

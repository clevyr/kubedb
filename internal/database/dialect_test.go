package database

import (
	"testing"

	"github.com/clevyr/kubedb/internal/config/conftypes"
	"github.com/clevyr/kubedb/internal/database/mariadb"
	"github.com/clevyr/kubedb/internal/database/mongodb"
	"github.com/clevyr/kubedb/internal/database/postgres"
	"github.com/clevyr/kubedb/internal/database/sqlformat"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		name    string
		args    args
		want    conftypes.Database
		wantErr require.ErrorAssertionFunc
	}{
		{"postgresql", args{"postgresql"}, postgres.Postgres{}, require.NoError},
		{"postgres", args{"postgres"}, postgres.Postgres{}, require.NoError},
		{"psql", args{"psql"}, postgres.Postgres{}, require.NoError},
		{"pg", args{"pg"}, postgres.Postgres{}, require.NoError},
		{"mariadb", args{"mariadb"}, mariadb.MariaDB{}, require.NoError},
		{"maria", args{"maria"}, mariadb.MariaDB{}, require.NoError},
		{"mysql", args{"mysql"}, mariadb.MariaDB{}, require.NoError},
		{"mongodb", args{"mongodb"}, mongodb.MongoDB{}, require.NoError},
		{"mongo", args{"mongo"}, mongodb.MongoDB{}, require.NoError},
		{"invalid", args{"invalid"}, nil, require.Error},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := New(tt.args.name)
			tt.wantErr(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestDetectFormat(t *testing.T) {
	type args struct {
		db   conftypes.DBFiler
		path string
	}
	tests := []struct {
		name string
		args args
		want sqlformat.Format
	}{
		{"postgres plain", args{postgres.Postgres{}, "test.sql"}, sqlformat.Plain},
		{"postgres gzipped", args{postgres.Postgres{}, "test.sql.gz"}, sqlformat.Gzip},
		{"postgres custom", args{postgres.Postgres{}, "test.dmp"}, sqlformat.Custom},
		{"mariadb plain", args{mariadb.MariaDB{}, "test.sql"}, sqlformat.Plain},
		{"mariadb gzipped", args{mariadb.MariaDB{}, "test.sql.gz"}, sqlformat.Gzip},
		{"mariadb unknown", args{mariadb.MariaDB{}, "test.sql.gz"}, sqlformat.Gzip},
		{"mongodb plain", args{mongodb.MongoDB{}, "test.archive"}, sqlformat.Plain},
		{"mongodb gzipped", args{mongodb.MongoDB{}, "test.archive.gz"}, sqlformat.Gzip},
		{"unknown", args{postgres.Postgres{}, "test.txt"}, sqlformat.Unknown},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, DetectFormat(tt.args.db, tt.args.path), "DetectFormat(%v, %v)", tt.args.db, tt.args.path)
		})
	}
}

func TestGetExtension(t *testing.T) {
	type args struct {
		db     conftypes.DBFiler
		format sqlformat.Format
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"postgres plain", args{postgres.Postgres{}, sqlformat.Plain}, ".sql"},
		{"postgres gzipped", args{postgres.Postgres{}, sqlformat.Gzip}, ".sql.gz"},
		{"postgres custom", args{postgres.Postgres{}, sqlformat.Custom}, ".dmp"},
		{"mariadb plain", args{mariadb.MariaDB{}, sqlformat.Plain}, ".sql"},
		{"mariadb gzipped", args{mariadb.MariaDB{}, sqlformat.Gzip}, ".sql.gz"},
		{"mongodb plain", args{mongodb.MongoDB{}, sqlformat.Plain}, ".archive"},
		{"mongodb gzipped", args{mongodb.MongoDB{}, sqlformat.Gzip}, ".archive.gz"},
		{"unknown", args{postgres.Postgres{}, sqlformat.Unknown}, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, GetExtension(tt.args.db, tt.args.format), "GetExtension(%v, %v)", tt.args.db, tt.args.format)
		})
	}
}

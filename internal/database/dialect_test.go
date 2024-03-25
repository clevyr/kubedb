package database

import (
	"testing"

	"github.com/clevyr/kubedb/internal/config"
	"github.com/clevyr/kubedb/internal/database/mariadb"
	"github.com/clevyr/kubedb/internal/database/mongodb"
	"github.com/clevyr/kubedb/internal/database/postgres"
	"github.com/clevyr/kubedb/internal/database/sqlformat"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	t.Parallel()
	type args struct {
		name string
	}
	tests := []struct {
		name    string
		args    args
		want    config.Database
		wantErr bool
	}{
		{"postgresql", args{"postgresql"}, postgres.Postgres{}, false},
		{"postgres", args{"postgres"}, postgres.Postgres{}, false},
		{"psql", args{"psql"}, postgres.Postgres{}, false},
		{"pg", args{"pg"}, postgres.Postgres{}, false},
		{"mariadb", args{"mariadb"}, mariadb.MariaDB{}, false},
		{"maria", args{"maria"}, mariadb.MariaDB{}, false},
		{"mysql", args{"mysql"}, mariadb.MariaDB{}, false},
		{"mongodb", args{"mongodb"}, mongodb.MongoDB{}, false},
		{"mongo", args{"mongo"}, mongodb.MongoDB{}, false},
		{"invalid", args{"invalid"}, nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := New(tt.args.name)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestDetectFormat(t *testing.T) {
	t.Parallel()
	type args struct {
		db   config.DatabaseFile
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
			t.Parallel()
			assert.Equalf(t, tt.want, DetectFormat(tt.args.db, tt.args.path), "DetectFormat(%v, %v)", tt.args.db, tt.args.path)
		})
	}
}

func TestGetExtension(t *testing.T) {
	t.Parallel()
	type args struct {
		db     config.DatabaseFile
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
			t.Parallel()
			assert.Equalf(t, tt.want, GetExtension(tt.args.db, tt.args.format), "GetExtension(%v, %v)", tt.args.db, tt.args.format)
		})
	}
}

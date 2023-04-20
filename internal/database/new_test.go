package database

import (
	"testing"

	"github.com/clevyr/kubedb/internal/config"
	"github.com/clevyr/kubedb/internal/database/dialect"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		name    string
		args    args
		want    config.Databaser
		wantErr bool
	}{
		{"postgresql", args{"postgresql"}, dialect.Postgres{}, false},
		{"postgres", args{"postgres"}, dialect.Postgres{}, false},
		{"psql", args{"psql"}, dialect.Postgres{}, false},
		{"pg", args{"pg"}, dialect.Postgres{}, false},
		{"mariadb", args{"mariadb"}, dialect.MariaDB{}, false},
		{"maria", args{"maria"}, dialect.MariaDB{}, false},
		{"mysql", args{"mysql"}, dialect.MariaDB{}, false},
		{"mongodb", args{"mongodb"}, dialect.MongoDB{}, false},
		{"mongo", args{"mongo"}, dialect.MongoDB{}, false},
		{"invalid", args{"invalid"}, nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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

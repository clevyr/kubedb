package database

import (
	"testing"

	"github.com/clevyr/kubedb/internal/config"
	"github.com/clevyr/kubedb/internal/database/mariadb"
	"github.com/clevyr/kubedb/internal/database/mongodb"
	"github.com/clevyr/kubedb/internal/database/postgres"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
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

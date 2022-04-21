package database

import (
	"github.com/clevyr/kubedb/internal/config"
	"github.com/clevyr/kubedb/internal/database/dialect"
	"reflect"
	"testing"
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
		{"invalid", args{"invalid"}, nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := New(tt.args.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("New() got = %v, want %v", got, tt.want)
			}
		})
	}
}

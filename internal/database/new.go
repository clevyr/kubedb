package database

import (
	"errors"
	"fmt"
	"github.com/clevyr/kubedb/internal/database/grammar"
)

var ErrUnsupportedDatabase = errors.New("unsupported database")

func New(name string) (Databaser, error) {
	switch name {
	case "postgresql", "postgres", "psql", "pg":
		return grammar.Postgres{}, nil
	case "mariadb", "maria", "mysql":
		return grammar.MariaDB{}, nil
	}
	return nil, fmt.Errorf("%v: %s", ErrUnsupportedDatabase, name)
}

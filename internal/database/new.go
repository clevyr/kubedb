package database

import (
	"errors"
	"fmt"

	"github.com/clevyr/kubedb/internal/config"
	"github.com/clevyr/kubedb/internal/database/dialect"
)

var ErrUnsupportedDatabase = errors.New("unsupported database")

func New(name string) (config.Database, error) {
	switch name {
	case "postgresql", "postgres", "psql", "pg":
		return dialect.Postgres{}, nil
	case "mariadb", "maria", "mysql":
		return dialect.MariaDB{}, nil
	case "mongodb", "mongo":
		return dialect.MongoDB{}, nil
	}
	return nil, fmt.Errorf("%v: %s", ErrUnsupportedDatabase, name)
}

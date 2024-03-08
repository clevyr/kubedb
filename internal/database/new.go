package database

import (
	"errors"
	"fmt"

	"github.com/clevyr/kubedb/internal/config"
	"github.com/clevyr/kubedb/internal/database/mariadb"
	"github.com/clevyr/kubedb/internal/database/mongodb"
	"github.com/clevyr/kubedb/internal/database/postgres"
	"github.com/clevyr/kubedb/internal/database/redis"
)

var ErrUnsupportedDatabase = errors.New("unsupported database")

func New(name string) (config.Database, error) {
	switch name {
	case "postgresql", "postgres", "psql", "pg":
		return postgres.Postgres{}, nil
	case "mariadb", "maria", "mysql":
		return mariadb.MariaDB{}, nil
	case "mongodb", "mongo":
		return mongodb.MongoDB{}, nil
	case "redis":
		return redis.Redis{}, nil
	}
	return nil, fmt.Errorf("%v: %s", ErrUnsupportedDatabase, name)
}

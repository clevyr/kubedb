package database

import (
	"errors"
	"fmt"
	"slices"

	"github.com/clevyr/kubedb/internal/config"
	"github.com/clevyr/kubedb/internal/database/mariadb"
	"github.com/clevyr/kubedb/internal/database/mongodb"
	"github.com/clevyr/kubedb/internal/database/postgres"
	"github.com/clevyr/kubedb/internal/database/redis"
)

func All() []config.Database {
	return []config.Database{
		postgres.Postgres{},
		mariadb.MariaDB{},
		mongodb.MongoDB{},
		redis.Redis{},
	}
}

func Names() []string {
	all := All()
	names := make([]string, 0, len(all))
	for _, db := range All() {
		names = append(names, db.Name())
	}
	return names
}

var ErrUnsupportedDatabase = errors.New("unsupported database")

func New(name string) (config.Database, error) {
	for _, db := range All() {
		if name == db.Name() {
			return db, nil
		}
		if dbAlias, ok := db.(config.DatabaseAliases); ok {
			if slices.Contains(dbAlias.Aliases(), name) {
				return db, nil
			}
		}
	}
	return nil, fmt.Errorf("%v: %s", ErrUnsupportedDatabase, name)
}

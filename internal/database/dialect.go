package database

import (
	"github.com/clevyr/kubedb/internal/config"
	"github.com/clevyr/kubedb/internal/database/mariadb"
	"github.com/clevyr/kubedb/internal/database/mongodb"
	"github.com/clevyr/kubedb/internal/database/postgres"
)

func All() []config.Database {
	return []config.Database{
		postgres.Postgres{},
		mariadb.MariaDB{},
		mongodb.MongoDB{},
	}
}

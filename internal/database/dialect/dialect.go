package dialect

import "github.com/clevyr/kubedb/internal/config"

func All() []config.Database {
	return []config.Database{
		Postgres{},
		MariaDB{},
		MongoDB{},
	}
}

package dialect

import "github.com/clevyr/kubedb/internal/config"

func All() []config.Databaser {
	return []config.Databaser{
		Postgres{},
		MariaDB{},
		MongoDB{},
	}
}

package conftypes

import "github.com/clevyr/kubedb/internal/database/sqlformat"

type Files struct {
	Filename string           `koanf:"name"`
	Format   sqlformat.Format `koanf:"format"`
}

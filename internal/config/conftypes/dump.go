package conftypes

import "github.com/clevyr/kubedb/internal/database/sqlformat"

type Dump struct {
	*Global          `koanf:"-"`
	Output           string           `koanf:"output"`
	Format           sqlformat.Format `koanf:"format"`
	IfExists         bool             `koanf:"if-exists"`
	Clean            bool             `koanf:"clean"`
	NoOwner          bool             `koanf:"no-owner"`
	Table            []string         `koanf:"table"`
	ExcludeTable     []string         `koanf:"exclude-table"`
	ExcludeTableData []string         `koanf:"exclude-table-data"`
}

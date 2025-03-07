package conftypes

type Dump struct {
	*Global          `koanf:"-"`
	Files            `koanf:",squash"`
	IfExists         bool     `koanf:"if-exists"`
	Clean            bool     `koanf:"clean"`
	NoOwner          bool     `koanf:"no-owner"`
	Table            []string `koanf:"table"`
	ExcludeTable     []string `koanf:"exclude-table"`
	ExcludeTableData []string `koanf:"exclude-table-data"`
}

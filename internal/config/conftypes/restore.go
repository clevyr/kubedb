package conftypes

import "github.com/clevyr/kubedb/internal/database/sqlformat"

type Restore struct {
	*Global           `koanf:"-"`
	Input             string           `koanf:"input"`
	Format            sqlformat.Format `koanf:"format"`
	SingleTransaction bool             `koanf:"single-transaction"`
	Clean             bool             `koanf:"clean"`
	NoOwner           bool             `koanf:"no-owner"`
	Force             bool             `koanf:"force"`
	Spinner           string           `koanf:"spinner"`
	HaltOnError       bool             `koanf:"halt-on-error"`
}

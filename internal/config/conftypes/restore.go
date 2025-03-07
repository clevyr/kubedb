package conftypes

type Restore struct {
	*Global           `koanf:"-"`
	Files             `koanf:",squash"`
	SingleTransaction bool   `koanf:"single-transaction"`
	Clean             bool   `koanf:"clean"`
	NoOwner           bool   `koanf:"no-owner"`
	Force             bool   `koanf:"force"`
	Spinner           string `koanf:"spinner"`
	HaltOnError       bool   `koanf:"halt-on-error"`
}

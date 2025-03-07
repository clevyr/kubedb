package conftypes

type Exec struct {
	*Global        `koanf:"-"`
	DisableHeaders bool   `koanf:"disable-headers"`
	Command        string `koanf:"command"`
	Opts           string `koanf:"opts"`
}

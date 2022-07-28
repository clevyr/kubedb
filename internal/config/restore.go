package config

type Restore struct {
	Global `mapstructure:",squash"`
	Files
	SingleTransaction bool
	Clean             bool
	NoOwner           bool
	Force             bool
}

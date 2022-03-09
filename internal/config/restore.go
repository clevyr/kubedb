package config

type Restore struct {
	Global
	SingleTransaction bool
	Clean             bool
	NoOwner           bool
	Force             bool
}

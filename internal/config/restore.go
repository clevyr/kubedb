package config

type Restore struct {
	Global
	Files
	SingleTransaction bool
	Clean             bool
	NoOwner           bool
	Force             bool
}

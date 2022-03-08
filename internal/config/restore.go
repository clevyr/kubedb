package config

type Restore struct {
	Database          string
	Username          string
	Password          string
	SingleTransaction bool
	Clean             bool
	NoOwner           bool
	Force             bool
}

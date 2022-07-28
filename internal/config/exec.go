package config

type Exec struct {
	Global         `mapstructure:",squash"`
	DisableHeaders bool
	Command        string
}

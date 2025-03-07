package config

type Dump struct {
	Global `mapstructure:",squash"`
	Files
	IfExists         bool
	Clean            bool
	NoOwner          bool
	Tables           []string
	ExcludeTable     []string
	ExcludeTableData []string
	Spinner          string
}

package config

type Dump struct {
	Global
	Files
	Directory        string
	IfExists         bool
	Clean            bool
	NoOwner          bool
	Tables           []string
	ExcludeTable     []string
	ExcludeTableData []string
}

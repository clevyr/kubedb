package config

import "github.com/clevyr/kubedb/internal/database/sqlformat"

type Restore struct {
	Global
	SingleTransaction bool
	Clean             bool
	NoOwner           bool
	Force             bool
	Format            sqlformat.Format
	Filename          string
}

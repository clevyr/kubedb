package config

import "github.com/clevyr/kubedb/internal/database/sqlformat"

type Files struct {
	Filename string `mapstructure:"name"`
	Format   sqlformat.Format
}

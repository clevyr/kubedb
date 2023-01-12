package config

import (
	"github.com/clevyr/kubedb/internal/config"
	"github.com/spf13/cobra"
)

type Config struct {
	config.Global
	RootCmd *cobra.Command
	Cmd     *cobra.Command
	Run     bool
}

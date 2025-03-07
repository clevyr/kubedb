package flags

import (
	"os"
	"strings"

	"gabe565.com/utils/must"
	"github.com/clevyr/kubedb/internal/config/path"
	"github.com/clevyr/kubedb/internal/consts"
	"github.com/spf13/cobra"
)

func Config(cmd *cobra.Command) {
	confPath := os.Getenv("KUBEDB_CONFIG")
	if confPath == "" {
		confPath, _ = path.GetConfigFile()
		if home, err := os.UserHomeDir(); err == nil && home != "/" {
			confPath = strings.Replace(confPath, home, "$HOME", 1)
		}
	}
	cmd.PersistentFlags().String(consts.FlagConfig, confPath, "Path to the config file")
}

func Spinner(cmd *cobra.Command) {
	cmd.Flags().String(consts.FlagSpinner, consts.DefaultSpinner, "Spinner from https://jsfiddle.net/sindresorhus/2eLtsbey/embedded/result/")
	must.Must(cmd.Flags().MarkHidden(consts.FlagSpinner))
}

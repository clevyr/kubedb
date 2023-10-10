package flags

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func Force(cmd *cobra.Command, p *bool) {
	cmd.Flags().BoolVarP(p, "force", "f", false, "Do not prompt before restore")
}

func Redact(cmd *cobra.Command) {
	cmd.PersistentFlags().Bool("redact", true, "Redact password from logs")
	if err := cmd.PersistentFlags().MarkHidden("redact"); err != nil {
		panic(err)
	}
}

func BindRedact(cmd *cobra.Command) {
	if err := viper.BindPFlag("log.redact", cmd.Flags().Lookup("redact")); err != nil {
		panic(err)
	}
}

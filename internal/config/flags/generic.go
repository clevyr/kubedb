package flags

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func Force(cmd *cobra.Command, p *bool) {
	cmd.Flags().BoolVarP(p, "force", "f", false, "do not prompt before restore")
	if err := viper.BindPFlag("force", cmd.Flags().Lookup("force")); err != nil {
		panic(err)
	}
}

func GitHubActions(cmd *cobra.Command) {
	cmd.PersistentFlags().Bool("github-actions", false, "Enables GitHub Actions log output")
	_ = cmd.PersistentFlags().MarkHidden("github-actions")
	if err := viper.BindPFlag("github-actions", cmd.PersistentFlags().Lookup("github-actions")); err != nil {
		panic(err)
	}
}

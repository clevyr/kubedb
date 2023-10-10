package flags

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func Force(cmd *cobra.Command, p *bool) {
	cmd.Flags().BoolVarP(p, "force", "f", false, "Do not prompt before restore")
}

func GitHubActions(cmd *cobra.Command) {
	cmd.PersistentFlags().Bool("github-actions", false, "Enables GitHub Actions log output")
	if err := cmd.PersistentFlags().MarkHidden("github-actions"); err != nil {
		panic(err)
	}
}

func BindGitHubActions(cmd *cobra.Command) {
	if err := viper.BindPFlag("github-actions", cmd.Flags().Lookup("github-actions")); err != nil {
		panic(err)
	}
}

func Redact(cmd *cobra.Command) {
	cmd.PersistentFlags().Bool("redact", true, "Redact password from logs")
	if err := cmd.PersistentFlags().MarkHidden("redact"); err != nil {
		panic(err)
	}
}

func BindRedact(cmd *cobra.Command) {
	if err := viper.BindPFlag("redact", cmd.Flags().Lookup("redact")); err != nil {
		panic(err)
	}
}

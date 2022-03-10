package flags

import "github.com/spf13/cobra"

func Force(cmd *cobra.Command, p *bool) {
	cmd.Flags().BoolVarP(p, "force", "f", false, "do not prompt before restore")
}

func GitHubActions(cmd *cobra.Command) {
	cmd.PersistentFlags().Bool("github-actions", false, "Enables GitHub Actions log output")
	_ = cmd.PersistentFlags().MarkHidden("github-actions")
}

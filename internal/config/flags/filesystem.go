package flags

import (
	"gabe565.com/utils/must"
	"github.com/clevyr/kubedb/internal/consts"
	"github.com/spf13/cobra"
)

func Directory(cmd *cobra.Command, p *string) {
	cmd.Flags().StringVarP(p, consts.FlagDirectory, "C", ".", "Directory to dump to")
	must.Must(cmd.RegisterFlagCompletionFunc(consts.FlagDirectory,
		func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
			return nil, cobra.ShellCompDirectiveFilterDirs
		}),
	)
	must.Must(cmd.Flags().MarkHidden(consts.FlagDirectory))
}

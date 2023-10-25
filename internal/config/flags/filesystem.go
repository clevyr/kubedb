package flags

import (
	"github.com/clevyr/kubedb/internal/consts"
	"github.com/spf13/cobra"
)

func Directory(cmd *cobra.Command, p *string) {
	cmd.Flags().StringVarP(p, consts.DirectoryFlag, "C", ".", "Directory to dump to")
	err := cmd.RegisterFlagCompletionFunc(
		consts.DirectoryFlag,
		func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return nil, cobra.ShellCompDirectiveFilterDirs
		})
	if err != nil {
		panic(err)
	}
}

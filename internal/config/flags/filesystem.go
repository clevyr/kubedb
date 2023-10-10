package flags

import (
	"github.com/spf13/cobra"
)

func Directory(cmd *cobra.Command, p *string) {
	cmd.Flags().StringVarP(p, "directory", "C", ".", "Directory to dump to")
	err := cmd.RegisterFlagCompletionFunc(
		"directory",
		func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return nil, cobra.ShellCompDirectiveFilterDirs
		})
	if err != nil {
		panic(err)
	}
}

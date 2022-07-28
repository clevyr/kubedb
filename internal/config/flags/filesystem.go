package flags

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func Directory(cmd *cobra.Command, p *string) {
	cmd.Flags().StringVarP(p, "directory", "C", ".", "directory to dump to")
	err := cmd.RegisterFlagCompletionFunc(
		"directory",
		func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return nil, cobra.ShellCompDirectiveFilterDirs
		})
	if err != nil {
		panic(err)
	}

	if err := viper.BindPFlag("directory", cmd.Flags().Lookup("directory")); err != nil {
		panic(err)
	}
}

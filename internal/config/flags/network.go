package flags

import (
	"github.com/spf13/cobra"
)

func Address(cmd *cobra.Command, p *[]string) {
	cmd.PersistentFlags().StringSliceVar(p, "address", []string{"127.0.0.1", "::1"}, "addresses to listen on (comma separated)")
	err := cmd.RegisterFlagCompletionFunc(
		"address",
		func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return []string{"127.0.0.1", "::1", "0.0.0.0", "::"}, cobra.ShellCompDirectiveNoFileComp
		})
	if err != nil {
		panic(err)
	}
}

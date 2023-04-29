package flags

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func Address(cmd *cobra.Command, p *[]string) {
	cmd.PersistentFlags().StringSliceVar(p, "address", []string{"127.0.0.1", "::1"}, "Addresses to listen on (comma separated)")
	err := cmd.RegisterFlagCompletionFunc(
		"address",
		func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return []string{"127.0.0.1\tprivate", "::1\tprivate", "0.0.0.0\tpublic", "::\tpublic"}, cobra.ShellCompDirectiveNoFileComp
		})
	if err != nil {
		panic(err)
	}
	if err := viper.BindPFlag("address", cmd.PersistentFlags().Lookup("address")); err != nil {
		panic(err)
	}
}

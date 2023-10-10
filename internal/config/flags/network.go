package flags

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func Address(cmd *cobra.Command) {
	cmd.PersistentFlags().StringSlice("address", []string{"127.0.0.1", "::1"}, "Addresses to listen on (comma separated)")
	err := cmd.RegisterFlagCompletionFunc(
		"address",
		func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return []string{"127.0.0.1\tprivate", "::1\tprivate", "0.0.0.0\tpublic", "::\tpublic"}, cobra.ShellCompDirectiveNoFileComp
		})
	if err != nil {
		panic(err)
	}
}

func BindAddress(cmd *cobra.Command) {
	if err := viper.BindPFlag("address", cmd.Flags().Lookup("address")); err != nil {
		panic(err)
	}
}

func RemoteGzip(cmd *cobra.Command) {
	cmd.Flags().Bool("remote-gzip", true, "Compress data over the wire. Results in lower bandwidth usage, but higher database load. May improve speed on fast connections.")
	err := cmd.RegisterFlagCompletionFunc(
		"remote-gzip",
		func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return []string{"true", "false"}, cobra.ShellCompDirectiveNoFileComp
		})
	if err != nil {
		panic(err)
	}
}

func BindRemoteGzip(cmd *cobra.Command) {
	if err := viper.BindPFlag("remote-gzip", cmd.Flags().Lookup("remote-gzip")); err != nil {
		panic(err)
	}
}

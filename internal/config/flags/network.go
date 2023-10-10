package flags

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func RemoteGzip(cmd *cobra.Command) {
	cmd.Flags().Bool("remote-gzip", true, "Compress data over the wire. Results in lower bandwidth usage, but higher database load. May improve speed on slow connections.")
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

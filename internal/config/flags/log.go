package flags

import "github.com/spf13/cobra"

func LogLevel(cmd *cobra.Command) {
	cmd.PersistentFlags().String("log-level", "info", "log level (trace, debug, info, warning, error, fatal, panic)")
	err := cmd.RegisterFlagCompletionFunc(
		"log-level",
		func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return []string{"trace", "debug", "info", "warning", "error", "fatal", "panic"}, cobra.ShellCompDirectiveNoFileComp
		})
	if err != nil {
		panic(err)
	}
}

func LogFormat(cmd *cobra.Command) {
	cmd.PersistentFlags().String("log-format", "text", "log formatter (text, json)")
	err := cmd.RegisterFlagCompletionFunc(
		"log-format",
		func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return []string{"text", "json"}, cobra.ShellCompDirectiveNoFileComp
		})
	if err != nil {
		panic(err)
	}
}

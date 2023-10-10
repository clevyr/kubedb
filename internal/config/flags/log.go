package flags

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func Quiet(cmd *cobra.Command, p *bool) {
	cmd.PersistentFlags().BoolVarP(p, "quiet", "q", false, "Silence remote log output")
}

func LogLevel(cmd *cobra.Command) {
	cmd.PersistentFlags().String("log-level", "info", "Log level (trace, debug, info, warning, error, fatal, panic)")
	err := cmd.RegisterFlagCompletionFunc(
		"log-level",
		func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return []string{
				log.TraceLevel.String(),
				log.DebugLevel.String(),
				log.InfoLevel.String(),
				log.WarnLevel.String(),
				log.ErrorLevel.String(),
				log.FatalLevel.String(),
				log.PanicLevel.String(),
			}, cobra.ShellCompDirectiveNoFileComp
		})
	if err != nil {
		panic(err)
	}
}

func BindLogLevel(cmd *cobra.Command) {
	if err := viper.BindPFlag("log.level", cmd.Flags().Lookup("log-level")); err != nil {
		panic(err)
	}
}

func LogFormat(cmd *cobra.Command) {
	cmd.PersistentFlags().String("log-format", "text", "Log formatter (text, json)")
	err := cmd.RegisterFlagCompletionFunc(
		"log-format",
		func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return []string{"text", "json"}, cobra.ShellCompDirectiveNoFileComp
		})
	if err != nil {
		panic(err)
	}
}

func BindLogFormat(cmd *cobra.Command) {
	if err := viper.BindPFlag("log.format", cmd.Flags().Lookup("log-format")); err != nil {
		panic(err)
	}
}

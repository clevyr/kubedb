package flags

import (
	"github.com/clevyr/kubedb/internal/consts"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func Quiet(cmd *cobra.Command, p *bool) {
	cmd.PersistentFlags().BoolVarP(p, consts.QuietFlag, "q", false, "Silence remote log output")
}

func LogLevel(cmd *cobra.Command) {
	cmd.PersistentFlags().String(consts.LogLevelFlag, "info", "Log level. One of (trace|debug|info|warning|error|fatal|panic)")
	err := cmd.RegisterFlagCompletionFunc(
		consts.LogLevelFlag,
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
	if err := viper.BindPFlag(consts.LogLevelKey, cmd.Flags().Lookup(consts.LogLevelFlag)); err != nil {
		panic(err)
	}
}

func LogFormat(cmd *cobra.Command) {
	cmd.PersistentFlags().String(consts.LogFormatFlag, "text", "Log formatter. One of (text|json)")
	err := cmd.RegisterFlagCompletionFunc(
		consts.LogFormatFlag,
		func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return []string{"text", "json"}, cobra.ShellCompDirectiveNoFileComp
		})
	if err != nil {
		panic(err)
	}
}

func BindLogFormat(cmd *cobra.Command) {
	if err := viper.BindPFlag(consts.LogFormatKey, cmd.Flags().Lookup(consts.LogFormatFlag)); err != nil {
		panic(err)
	}
}

func Redact(cmd *cobra.Command) {
	cmd.PersistentFlags().Bool(consts.RedactFlag, true, "Redact password from logs")
	if err := cmd.PersistentFlags().MarkHidden(consts.RedactFlag); err != nil {
		panic(err)
	}
}

func BindRedact(cmd *cobra.Command) {
	if err := viper.BindPFlag(consts.LogRedactKey, cmd.Flags().Lookup(consts.RedactFlag)); err != nil {
		panic(err)
	}
}

func Healthchecks(cmd *cobra.Command) {
	cmd.PersistentFlags().String(consts.HealthchecksPingUrlFlag, "", "Notification handler URL")
}

func BindHealthchecks(cmd *cobra.Command) {
	if err := viper.BindPFlag(consts.HealthchecksPingUrlKey, cmd.Flags().Lookup(consts.HealthchecksPingUrlFlag)); err != nil {
		panic(err)
	}
}

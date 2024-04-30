package flags

import (
	"github.com/clevyr/kubedb/internal/consts"
	"github.com/clevyr/kubedb/internal/util"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func Quiet(cmd *cobra.Command, p *bool) {
	cmd.PersistentFlags().BoolVarP(p, consts.QuietFlag, "q", false, "Silence remote log output")
	if err := cmd.RegisterFlagCompletionFunc(consts.QuietFlag, util.BoolCompletion); err != nil {
		panic(err)
	}
}

func LogLevel(cmd *cobra.Command) {
	cmd.PersistentFlags().String(consts.LogLevelFlag, "info", "Log level. One of (trace|debug|info|warn|error|fatal|panic)")
	err := cmd.RegisterFlagCompletionFunc(
		consts.LogLevelFlag,
		func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
			return []string{
				zerolog.TraceLevel.String(),
				zerolog.DebugLevel.String(),
				zerolog.InfoLevel.String(),
				zerolog.WarnLevel.String(),
				zerolog.ErrorLevel.String(),
				zerolog.FatalLevel.String(),
				zerolog.PanicLevel.String(),
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
	cmd.PersistentFlags().String(consts.LogFormatFlag, "auto", "Log formatter. One of (auto|color|plain|json)")
	err := cmd.RegisterFlagCompletionFunc(
		consts.LogFormatFlag,
		func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
			return []string{"auto", "color", "plain", "json"}, cobra.ShellCompDirectiveNoFileComp
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
	cmd.PersistentFlags().String(consts.HealthchecksPingURLFlag, "", "Notification handler URL")
	if err := cmd.RegisterFlagCompletionFunc(consts.HealthchecksPingURLFlag, cobra.NoFileCompletions); err != nil {
		panic(err)
	}
}

func BindHealthchecks(cmd *cobra.Command) {
	if err := viper.BindPFlag(consts.HealthchecksPingURLKey, cmd.Flags().Lookup(consts.HealthchecksPingURLFlag)); err != nil {
		panic(err)
	}
}

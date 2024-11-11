package flags

import (
	"os"
	"strings"

	"gabe565.com/utils/must"
	"github.com/clevyr/kubedb/internal/consts"
	"github.com/clevyr/kubedb/internal/log"
	"github.com/clevyr/kubedb/internal/util"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"k8s.io/kubectl/pkg/util/term"
)

func Quiet(cmd *cobra.Command, p *bool) {
	cmd.PersistentFlags().BoolVarP(p, consts.QuietFlag, "q", false, "Silence remote log output")
	must.Must(cmd.RegisterFlagCompletionFunc(consts.QuietFlag, util.BoolCompletion))
}

func Progress(cmd *cobra.Command, p *bool) {
	cmd.PersistentFlags().BoolVar(p, consts.ProgressFlag, (term.TTY{Out: os.Stderr}).IsTerminalOut(), "Enables the progress bar")
	must.Must(cmd.RegisterFlagCompletionFunc(consts.ProgressFlag, util.BoolCompletion))
}

func BindProgress(cmd *cobra.Command) {
	must.Must(viper.BindPFlag(consts.ProgressKey, cmd.Flags().Lookup(consts.ProgressFlag)))
}

func LogLevel(cmd *cobra.Command) {
	cmd.PersistentFlags().String(consts.LogLevelFlag, "info", "Log level (one of "+strings.Join(log.LevelStrings(), ", ")+")")
	must.Must(cmd.RegisterFlagCompletionFunc(consts.LogLevelFlag,
		func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
			return log.LevelStrings(), cobra.ShellCompDirectiveNoFileComp
		}),
	)
}

func BindLogLevel(cmd *cobra.Command) {
	must.Must(viper.BindPFlag(consts.LogLevelKey, cmd.Flags().Lookup(consts.LogLevelFlag)))
}

func LogFormat(cmd *cobra.Command) {
	cmd.PersistentFlags().String(consts.LogFormatFlag, "auto", "Log format (one of "+strings.Join(log.FormatStrings(), ", ")+")")
	must.Must(cmd.RegisterFlagCompletionFunc(consts.LogFormatFlag,
		func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
			return []string{"auto", "color", "plain", "json"}, cobra.ShellCompDirectiveNoFileComp
		}),
	)
}

func BindLogFormat(cmd *cobra.Command) {
	must.Must(viper.BindPFlag(consts.LogFormatKey, cmd.Flags().Lookup(consts.LogFormatFlag)))
}

func Mask(cmd *cobra.Command) {
	cmd.PersistentFlags().Bool(consts.MaskFlag, true, "Mask password in logs")
	must.Must(cmd.PersistentFlags().MarkHidden(consts.MaskFlag))
}

func BindMask(cmd *cobra.Command) {
	must.Must(viper.BindPFlag(consts.LogMaskKey, cmd.Flags().Lookup(consts.MaskFlag)))
}

func Healthchecks(cmd *cobra.Command) {
	cmd.PersistentFlags().String(consts.HealthchecksPingURLFlag, "", "Notification handler URL")
	must.Must(cmd.RegisterFlagCompletionFunc(consts.HealthchecksPingURLFlag, cobra.NoFileCompletions))
}

func BindHealthchecks(cmd *cobra.Command) {
	must.Must(viper.BindPFlag(consts.HealthchecksPingURLKey, cmd.Flags().Lookup(consts.HealthchecksPingURLFlag)))
}

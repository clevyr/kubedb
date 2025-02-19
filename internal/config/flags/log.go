package flags

import (
	"os"
	"strings"

	"gabe565.com/utils/must"
	"gabe565.com/utils/slogx"
	"github.com/clevyr/kubedb/internal/consts"
	"github.com/clevyr/kubedb/internal/util"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"k8s.io/kubectl/pkg/util/term"
)

func Quiet(cmd *cobra.Command, p *bool) {
	cmd.PersistentFlags().BoolVarP(p, consts.FlagQuiet, "q", false, "Silence remote log output")
	must.Must(cmd.RegisterFlagCompletionFunc(consts.FlagQuiet, util.BoolCompletion))
}

func Progress(cmd *cobra.Command, p *bool) {
	cmd.PersistentFlags().BoolVar(p, consts.FlagProgress, (term.TTY{Out: os.Stderr}).IsTerminalOut(), "Enables the progress bar")
	must.Must(cmd.RegisterFlagCompletionFunc(consts.FlagProgress, util.BoolCompletion))
}

func BindProgress(cmd *cobra.Command) {
	must.Must(viper.BindPFlag(consts.KeyProgress, cmd.Flags().Lookup(consts.FlagProgress)))
}

func LogLevel(cmd *cobra.Command) {
	cmd.PersistentFlags().String(consts.FlagLogLevel, "info", "Log level (one of "+strings.Join(slogx.LevelStrings(), ", ")+")")
	must.Must(cmd.RegisterFlagCompletionFunc(consts.FlagLogLevel,
		func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
			return slogx.LevelStrings(), cobra.ShellCompDirectiveNoFileComp
		}),
	)
}

func BindLogLevel(cmd *cobra.Command) {
	must.Must(viper.BindPFlag(consts.KeyLogLevel, cmd.Flags().Lookup(consts.FlagLogLevel)))
}

func LogFormat(cmd *cobra.Command) {
	cmd.PersistentFlags().String(consts.FlagLogFormat, "auto", "Log format (one of "+strings.Join(slogx.FormatStrings(), ", ")+")")
	must.Must(cmd.RegisterFlagCompletionFunc(consts.FlagLogFormat,
		func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
			return slogx.FormatStrings(), cobra.ShellCompDirectiveNoFileComp
		}),
	)
}

func BindLogFormat(cmd *cobra.Command) {
	must.Must(viper.BindPFlag(consts.KeyLogFormat, cmd.Flags().Lookup(consts.FlagLogFormat)))
}

func Mask(cmd *cobra.Command) {
	cmd.PersistentFlags().Bool(consts.FlagMask, true, "Mask password in logs")
	must.Must(cmd.PersistentFlags().MarkHidden(consts.FlagMask))
}

func BindMask(cmd *cobra.Command) {
	must.Must(viper.BindPFlag(consts.KeyLogMask, cmd.Flags().Lookup(consts.FlagMask)))
}

func Healthchecks(cmd *cobra.Command) {
	cmd.PersistentFlags().String(consts.FlagHealchecksPingURL, "", "Notification handler URL")
	must.Must(cmd.RegisterFlagCompletionFunc(consts.FlagHealchecksPingURL, cobra.NoFileCompletions))
}

func BindHealthchecks(cmd *cobra.Command) {
	must.Must(viper.BindPFlag(consts.KeyHealthchecksPingURL, cmd.Flags().Lookup(consts.FlagHealchecksPingURL)))
}

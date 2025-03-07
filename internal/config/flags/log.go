package flags

import (
	"os"
	"strings"

	"gabe565.com/utils/must"
	"gabe565.com/utils/slogx"
	"gabe565.com/utils/termx"
	"github.com/clevyr/kubedb/internal/completion"
	"github.com/clevyr/kubedb/internal/consts"
	"github.com/spf13/cobra"
)

func Quiet(cmd *cobra.Command) {
	cmd.PersistentFlags().BoolP(consts.FlagQuiet, "q", false, "Silence remote log output")
	must.Must(cmd.RegisterFlagCompletionFunc(consts.FlagQuiet, completion.BoolCompletion))
}

func Progress(cmd *cobra.Command) {
	cmd.PersistentFlags().Bool(consts.FlagProgress, termx.IsTerminal(os.Stderr), "Enables the progress bar")
	must.Must(cmd.RegisterFlagCompletionFunc(consts.FlagProgress, completion.BoolCompletion))
}

func Log(cmd *cobra.Command) {
	cmd.PersistentFlags().String(consts.FlagLogLevel, "info", "Log level (one of "+strings.Join(slogx.LevelStrings(), ", ")+")")
	must.Must(cmd.RegisterFlagCompletionFunc(consts.FlagLogLevel,
		func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
			return slogx.LevelStrings(), cobra.ShellCompDirectiveNoFileComp
		}),
	)

	cmd.PersistentFlags().String(consts.FlagLogFormat, "auto", "Log format (one of "+strings.Join(slogx.FormatStrings(), ", ")+")")
	must.Must(cmd.RegisterFlagCompletionFunc(consts.FlagLogFormat,
		func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
			return slogx.FormatStrings(), cobra.ShellCompDirectiveNoFileComp
		}),
	)

	cmd.PersistentFlags().Bool(consts.FlagLogMask, true, "Mask password in logs")
	must.Must(cmd.PersistentFlags().MarkHidden(consts.FlagLogMask))
}

func Healthchecks(cmd *cobra.Command) {
	cmd.PersistentFlags().String(consts.FlagHealthchecksPingURL, "", "Notification handler URL")
	must.Must(cmd.RegisterFlagCompletionFunc(consts.FlagHealthchecksPingURL, cobra.NoFileCompletions))
}

package restore

import (
	"errors"
	"fmt"
	"os"

	"github.com/AlecAivazis/survey/v2"
	"github.com/clevyr/kubedb/internal/actions/restore"
	"github.com/clevyr/kubedb/internal/config/flags"
	"github.com/clevyr/kubedb/internal/util"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/kubectl/pkg/util/term"
)

var action restore.Restore

func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "restore filename",
		Aliases: []string{"r", "import"},
		Short:   "Restore a sql file to a database",
		Long: `Restore a sql file to a database.

Supported Input Filetypes:
  - Raw sql file. Typically with the ` + "`" + `.sql` + "`" + ` extension
  - Gzipped sql file. Typically with the ".sql.gz" extension
  - For Postgres: custom dump file. Typically with the ".dmp" extension`,

		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: validArgs,
		GroupID:           "rw",

		PreRunE: preRun,
		RunE:    run,
	}

	flags.JobPodLabels(cmd)
	flags.NoJob(cmd)
	flags.Format(cmd, &action.Format)
	flags.Port(cmd)
	flags.Database(cmd)
	flags.Username(cmd)
	flags.Password(cmd)
	flags.SingleTransaction(cmd, &action.SingleTransaction)
	flags.Clean(cmd, &action.Clean)
	flags.NoOwner(cmd, &action.NoOwner)
	flags.Force(cmd, &action.Force)
	flags.Quiet(cmd, &action.Quiet)
	flags.RemoteGzip(cmd)
	flags.Analyze(cmd)

	return cmd
}

var setupOptions = util.SetupOptions{Name: "restore"}

func validArgs(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) != 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	setupOptions.DisableJob = true

	err := preRun(cmd, args)
	if err != nil {
		return []string{"sql", "sql.gz", "dmp", "archive", "archive.gz"}, cobra.ShellCompDirectiveFilterFileExt
	}

	formats := action.Dialect.Formats()
	exts := make([]string, 0, len(formats))
	for _, ext := range formats {
		exts = append(exts, ext[1:])
	}
	return exts, cobra.ShellCompDirectiveFilterFileExt
}

func preRun(cmd *cobra.Command, args []string) (err error) {
	flags.BindRemoteGzip(cmd)
	action.RemoteGzip = viper.GetBool("remote-gzip")
	flags.BindAnalyze(cmd)
	action.Analyze = viper.GetBool("restore.analyze")
	flags.BindJobPodLabels(cmd)
	flags.BindNoJob(cmd)

	if len(args) > 0 {
		action.Filename = args[0]
	}

	if err := util.DefaultSetup(cmd, &action.Global, setupOptions); err != nil {
		return err
	}

	if !cmd.Flags().Lookup("format").Changed {
		action.Format = action.Dialect.FormatFromFilename(action.Filename)
	}

	return nil
}

func run(cmd *cobra.Command, args []string) (err error) {
	defer func() {
		util.Teardown(cmd, &action.Global)
	}()

	if !action.Force {
		tty := term.TTY{In: os.Stdin}.IsTerminalIn()
		if tty {
			var response bool
			err := survey.AskOne(&survey.Confirm{
				Message: "Restore to " + action.DbPod.Name + " in " + action.Client.Namespace + "?",
			}, &response)
			if err != nil {
				return err
			}
			if !response {
				fmt.Println("restore canceled")
				return nil
			}
		} else {
			return errors.New("refusing to restore a database non-interactively without the --force flag")
		}
	}

	return action.Run(cmd.Context())
}

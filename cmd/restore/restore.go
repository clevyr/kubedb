package restore

import (
	"errors"
	"fmt"
	"os"

	"github.com/AlecAivazis/survey/v2"
	"github.com/clevyr/kubedb/internal/actions/restore"
	"github.com/clevyr/kubedb/internal/config"
	"github.com/clevyr/kubedb/internal/config/flags"
	"github.com/clevyr/kubedb/internal/consts"
	"github.com/clevyr/kubedb/internal/database"
	"github.com/clevyr/kubedb/internal/util"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"k8s.io/kubectl/pkg/util/term"
)

var action restore.Restore

func New() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "restore filename",
		Aliases: []string{"r", "import"},
		Short:   "Restore a sql file to a database",
		Long:    newDescription(),

		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: validArgs,
		GroupID:           "rw",

		PreRunE: preRun,
		RunE:    run,
	}

	flags.JobPodLabels(cmd)
	flags.CreateJob(cmd)
	flags.CreateNetworkPolicy(cmd)
	flags.Format(cmd, &action.Format)
	flags.Port(cmd)
	flags.Database(cmd)
	flags.Username(cmd)
	flags.Password(cmd)
	flags.SingleTransaction(cmd, &action.SingleTransaction)
	flags.Clean(cmd, &action.Clean)
	flags.NoOwner(cmd, &action.NoOwner)
	flags.Quiet(cmd, &action.Quiet)
	flags.RemoteGzip(cmd)
	flags.Analyze(cmd)
	flags.HaltOnError(cmd)
	flags.Spinner(cmd, &action.Spinner)
	flags.Opts(cmd)
	cmd.Flags().BoolVarP(&action.Force, consts.ForceFlag, "f", false, "Do not prompt before restore")

	return cmd
}

func validArgs(cmd *cobra.Command, args []string, _ string) ([]string, cobra.ShellCompDirective) {
	if len(args) != 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	viper.Set(consts.CreateJobKey, false)
	action.Force = true
	action.Filename = "-"

	err := preRun(cmd, args)
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	db, ok := action.Dialect.(config.DatabaseRestore)
	if !ok {
		return nil, cobra.ShellCompDirectiveError
	}

	formats := db.Formats()
	exts := make([]string, 0, len(formats))
	for _, ext := range formats {
		exts = append(exts, ext[1:])
	}
	return exts, cobra.ShellCompDirectiveFilterFileExt
}

func preRun(cmd *cobra.Command, args []string) error {
	flags.BindRemoteGzip(cmd)
	action.RemoteGzip = viper.GetBool(consts.RemoteGzipKey)
	flags.BindAnalyze(cmd)
	action.Analyze = viper.GetBool(consts.AnalyzeKey)
	flags.BindJobPodLabels(cmd)
	flags.BindCreateJob(cmd)
	flags.BindCreateNetworkPolicy(cmd)
	flags.BindSpinner(cmd)
	flags.BindHaltOnError(cmd)
	flags.BindOpts(cmd)
	action.HaltOnError = viper.GetBool(consts.HaltOnErrorKey)
	action.Spinner = viper.GetString(consts.SpinnerKey)

	if len(args) > 0 {
		action.Filename = args[0]
	}

	setupOptions := util.SetupOptions{Name: "restore"}
	if err := util.DefaultSetup(cmd, &action.Global, setupOptions); err != nil {
		return err
	}

	if action.Filename != "-" {
		if _, err := os.Stat(action.Filename); errors.Is(err, os.ErrNotExist) {
			return err
		}
	}

	if !cmd.Flags().Lookup(consts.FormatFlag).Changed {
		db, ok := action.Dialect.(config.DatabaseRestore)
		if !ok {
			return fmt.Errorf("%w: %s", util.ErrNoRestore, action.Dialect.Name())
		}

		action.Format = database.DetectFormat(db, action.Filename)
	}

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
				return errors.New("restore canceled")
			}
		} else {
			return errors.New("refusing to restore a database non-interactively without the --force flag")
		}
	}

	if err := util.CreateJob(cmd.Context(), &action.Global, setupOptions); err != nil {
		return err
	}

	return nil
}

func run(cmd *cobra.Command, _ []string) error {
	return action.Run(cmd.Context())
}

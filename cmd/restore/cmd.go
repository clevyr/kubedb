package restore

import (
	"errors"
	"fmt"
	"log/slog"
	"maps"
	"net/url"
	"os"
	"slices"

	"gabe565.com/utils/must"
	"github.com/charmbracelet/huh"
	"github.com/clevyr/kubedb/internal/actions/restore"
	"github.com/clevyr/kubedb/internal/config"
	"github.com/clevyr/kubedb/internal/config/flags"
	"github.com/clevyr/kubedb/internal/consts"
	"github.com/clevyr/kubedb/internal/database"
	"github.com/clevyr/kubedb/internal/storage"
	"github.com/clevyr/kubedb/internal/tui"
	"github.com/clevyr/kubedb/internal/util"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"k8s.io/kubectl/pkg/util/term"
)

//nolint:gochecknoglobals
var (
	action       restore.Restore
	setupOptions = util.SetupOptions{Name: "restore"}
)

func New() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "restore filename",
		Aliases: []string{"r", "import"},
		Short:   "Restore a sql file to a database",
		Long:    newDescription(),

		Args:              cobra.MaximumNArgs(1),
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
	flags.Progress(cmd, &action.Progress)
	cmd.Flags().BoolVarP(&action.Force, consts.ForceFlag, "f", false, "Do not prompt before restore")
	must.Must(cmd.RegisterFlagCompletionFunc(consts.ForceFlag, util.BoolCompletion))

	return cmd
}

func validArgs(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) != 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	viper.Set(consts.CreateJobKey, false)
	action.Force = true
	action.Filename = "-"
	setupOptions.NoSurvey = true

	err := preRun(cmd, args)
	if err != nil {
		slog.Error("Pre-run failed", "error", err)
		return nil, cobra.ShellCompDirectiveError
	}

	db, ok := action.Dialect.(config.DBRestorer)
	if !ok {
		slog.Error("Dialect does not support restore", "name", action.Dialect.Name())
		return nil, cobra.ShellCompDirectiveError
	}

	formats := db.Formats()

	if storage.IsCloud(toComplete) {
		u, err := url.Parse(toComplete)
		if err != nil {
			slog.Error("Failed to parse URL", "error", err)
			return nil, cobra.ShellCompDirectiveError
		}

		switch {
		case storage.IsS3(toComplete):
			if u.Host == "" || u.Path == "" {
				return storage.CompleteBucketsS3(u)
			} else {
				return storage.CompleteObjectsS3(u, slices.Collect(maps.Values(formats)), false)
			}
		case storage.IsGCS(toComplete):
			if u.Host == "" || u.Path == "" {
				return storage.CompleteBucketsGCS(u, "")
			} else {
				return storage.CompleteObjectsGCS(u, slices.Collect(maps.Values(formats)), false)
			}
		}
	}

	exts := make([]string, 0, len(formats))
	for _, ext := range formats {
		exts = append(exts, ext[1:])
	}
	return exts, cobra.ShellCompDirectiveFilterFileExt
}

var (
	ErrRestoreCanceled = errors.New("restore canceled")
	ErrRestoreRefused  = errors.New("refusing to restore a database non-interactively without the --force flag")
)

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
	flags.BindProgress(cmd)
	action.Progress = viper.GetBool(consts.ProgressKey)
	action.HaltOnError = viper.GetBool(consts.HaltOnErrorKey)
	action.Spinner = viper.GetString(consts.SpinnerKey)

	if err := util.DefaultSetup(cmd, &action.Global, setupOptions); err != nil {
		return err
	}

	if len(args) > 0 {
		action.Filename = args[0]
	}

	switch {
	case action.Filename == "-", storage.IsCloud(action.Filename):
	case action.Filename == "":
		if (term.TTY{In: os.Stdin}).IsTerminalIn() {
			db, ok := action.Dialect.(config.DBRestorer)
			if !ok {
				return fmt.Errorf("%w: %s", util.ErrNoRestore, action.Dialect.Name())
			}

			wd, err := os.Getwd()
			if err != nil {
				return err
			}

			form := tui.NewForm(huh.NewGroup(
				huh.NewFilePicker().
					Title("Choose a file to restore").
					Picking(true).
					CurrentDirectory(wd).
					ShowSize(true).
					ShowPermissions(false).
					Height(15).
					AllowedTypes(slices.Collect(maps.Values(db.Formats()))).
					Value(&action.Filename),
			))

			if err := form.Run(); err != nil {
				return err
			}
		} else {
			return cobra.ExactArgs(1)(cmd, args)
		}
		fallthrough
	default:
		if _, err := os.Stat(action.Filename); errors.Is(err, os.ErrNotExist) {
			return err
		}
	}

	if !cmd.Flags().Lookup(consts.FormatFlag).Changed {
		db, ok := action.Dialect.(config.DBRestorer)
		if !ok {
			return fmt.Errorf("%w: %s", util.ErrNoRestore, action.Dialect.Name())
		}

		action.Format = database.DetectFormat(db, action.Filename)
	}

	switch {
	case action.Force:
	case (term.TTY{In: os.Stdin}).IsTerminalIn():
		if response, err := action.Confirm(); err != nil {
			return err
		} else if !response {
			return ErrRestoreCanceled
		}
	default:
		return ErrRestoreRefused
	}

	if err := util.CreateJob(cmd.Context(), &action.Global, setupOptions); err != nil {
		return err
	}

	return nil
}

func run(cmd *cobra.Command, _ []string) error {
	return action.Run(cmd.Context())
}

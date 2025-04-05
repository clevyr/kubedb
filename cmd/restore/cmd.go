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
	"gabe565.com/utils/termx"
	"github.com/charmbracelet/huh"
	"github.com/clevyr/kubedb/internal/actions"
	"github.com/clevyr/kubedb/internal/actions/restore"
	"github.com/clevyr/kubedb/internal/completion"
	"github.com/clevyr/kubedb/internal/config"
	"github.com/clevyr/kubedb/internal/config/conftypes"
	"github.com/clevyr/kubedb/internal/config/flags"
	"github.com/clevyr/kubedb/internal/consts"
	"github.com/clevyr/kubedb/internal/database"
	"github.com/clevyr/kubedb/internal/storage"
	"github.com/clevyr/kubedb/internal/tui"
	"github.com/clevyr/kubedb/internal/util"
	"github.com/spf13/cobra"
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
	flags.Format(cmd)
	flags.Port(cmd)
	flags.Database(cmd)
	flags.Username(cmd)
	flags.Password(cmd)
	flags.SingleTransaction(cmd)
	flags.Clean(cmd)
	flags.NoOwner(cmd)
	flags.Quiet(cmd)
	flags.RemoteGzip(cmd)
	flags.Analyze(cmd)
	flags.HaltOnError(cmd)
	flags.Spinner(cmd)
	flags.Opts(cmd)
	flags.Progress(cmd)
	cmd.Flags().StringP(consts.FlagInput, "i", "", "Input file path (can also be set using a positional arg)")
	must.Must(cmd.RegisterFlagCompletionFunc(consts.FlagInput, validArgs))
	cmd.Flags().BoolP(consts.FlagForce, "f", false, "Do not prompt before restore")
	must.Must(cmd.RegisterFlagCompletionFunc(consts.FlagForce, completion.BoolCompletion))

	return cmd
}

func preRun(cmd *cobra.Command, args []string) error {
	action := &restore.Restore{Restore: conftypes.Restore{Global: config.Global}}

	if err := config.Unmarshal(cmd, "restore", &action); err != nil {
		return err
	}
	if err := util.DefaultSetup(cmd, action.Global); err != nil {
		return err
	}
	if len(args) > 0 {
		action.Input = args[0]
	}
	if !cmd.Flags().Lookup(consts.FlagFormat).Changed {
		db, ok := action.Dialect.(conftypes.DBRestorer)
		if !ok {
			return fmt.Errorf("%w: %s", util.ErrNoRestore, action.Dialect.Name())
		}

		action.Format = database.DetectFormat(db, action.Input)
	}

	cmd.SetContext(actions.NewContext(cmd.Context(), action))
	return nil
}

func validArgs(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) != 0 || must.Must2(cmd.Flags().GetString(consts.FlagInput)) != "" {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	must.Must(config.K.Set(consts.FlagCreateJob, false))

	config.Global.SkipSurvey = true
	err := preRun(cmd, args)
	if err != nil {
		slog.Error("Pre-run failed", "error", err)
		return nil, cobra.ShellCompDirectiveError
	}

	action := actions.FromContext[*restore.Restore](cmd.Context())

	db, ok := action.Dialect.(conftypes.DBRestorer)
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
			}
			return storage.CompleteObjectsS3(u, slices.Collect(maps.Values(formats)), false)
		case storage.IsGCS(toComplete):
			if u.Host == "" || u.Path == "" {
				return storage.CompleteBucketsGCS(u, "")
			}
			return storage.CompleteObjectsGCS(u, slices.Collect(maps.Values(formats)), false)
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

func run(cmd *cobra.Command, args []string) error {
	action := actions.FromContext[*restore.Restore](cmd.Context())

	switch {
	case action.Input == "-", storage.IsCloud(action.Input), config.IsCompletion:
	case action.Input == "":
		if termx.IsTerminal(cmd.InOrStdin()) {
			db, ok := action.Dialect.(conftypes.DBRestorer)
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
					Value(&action.Input),
			))

			if err := form.Run(); err != nil {
				return err
			}
		} else {
			return cobra.ExactArgs(1)(cmd, args)
		}
		fallthrough
	default:
		if _, err := os.Stat(action.Input); err != nil {
			return err
		}
	}

	if !cmd.Flags().Lookup(consts.FlagFormat).Changed {
		db, ok := action.Dialect.(conftypes.DBRestorer)
		if !ok {
			return fmt.Errorf("%w: %s", util.ErrNoRestore, action.Dialect.Name())
		}

		action.Format = database.DetectFormat(db, action.Input)
	}

	switch {
	case action.Force:
	case termx.IsTerminal(cmd.InOrStdin()):
		if response, err := action.Confirm(); err != nil {
			return err
		} else if !response {
			return ErrRestoreCanceled
		}
	default:
		return ErrRestoreRefused
	}

	if err := util.CreateJob(cmd.Context(), cmd, action.Global); err != nil {
		return err
	}

	return action.Run(cmd.Context())
}

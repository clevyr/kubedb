package dump

import (
	"fmt"
	"log/slog"
	"maps"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"gabe565.com/utils/must"
	"github.com/clevyr/kubedb/internal/actions/dump"
	"github.com/clevyr/kubedb/internal/config"
	"github.com/clevyr/kubedb/internal/config/conftypes"
	"github.com/clevyr/kubedb/internal/config/flags"
	"github.com/clevyr/kubedb/internal/consts"
	"github.com/clevyr/kubedb/internal/database"
	"github.com/clevyr/kubedb/internal/storage"
	"github.com/clevyr/kubedb/internal/util"
	"github.com/spf13/cobra"
)

//nolint:gochecknoglobals
var action = &dump.Dump{Dump: conftypes.Dump{Global: config.Global}}

func New() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "dump [filename | bucket URI]",
		Aliases: []string{"d", "export"},
		Short:   "Dump a database to a sql file",
		Long:    newDescription(),

		Args:              cobra.MaximumNArgs(1),
		ValidArgsFunction: validArgs,
		GroupID:           "ro",

		PreRunE: preRun,
		RunE:    run,
	}

	flags.JobPodLabels(cmd)
	flags.CreateJob(cmd)
	flags.CreateNetworkPolicy(cmd)
	flags.Port(cmd)
	flags.Database(cmd)
	flags.Username(cmd)
	flags.Password(cmd)
	flags.Format(cmd)
	flags.IfExists(cmd)
	flags.Clean(cmd)
	flags.NoOwner(cmd)
	flags.Tables(cmd)
	flags.ExcludeTable(cmd)
	flags.ExcludeTableData(cmd)
	flags.Quiet(cmd)
	flags.RemoteGzip(cmd)
	flags.Spinner(cmd)
	flags.Opts(cmd)
	flags.Progress(cmd)
	cmd.Flags().StringP(consts.FlagOutput, "o", "", "Output file path (can also be set using a positional arg)")
	must.Must(cmd.RegisterFlagCompletionFunc(consts.FlagOutput, validArgs))

	return cmd
}

func validArgs(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) != 0 || must.Must2(cmd.Flags().GetString(consts.FlagOutput)) != "" {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	must.Must(config.K.Set(consts.FlagCreateJob, false))
	config.Global.SkipSurvey = true
	err := preRun(cmd, args)
	if err != nil {
		slog.Error("Pre-run failed", "error", err)
		return nil, cobra.ShellCompDirectiveError
	}

	db, ok := action.Dialect.(conftypes.DBDumper)
	if !ok {
		slog.Error("Dialect does not support dump", "name", action.Dialect.Name())
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
			return storage.CompleteObjectsS3(u, slices.Collect(maps.Values(formats)), true)
		case storage.IsGCS(toComplete):
			if u.Host == "" || u.Path == "" {
				return storage.CompleteBucketsGCS(u, "")
			}
			return storage.CompleteObjectsGCS(u, slices.Collect(maps.Values(formats)), true)
		}
	}

	exts := make([]string, 0, len(formats))
	for _, ext := range formats {
		exts = append(exts, ext[1:])
	}
	return exts, cobra.ShellCompDirectiveFilterFileExt
}

func preRun(cmd *cobra.Command, args []string) error {
	if err := config.Unmarshal(cmd, "dump", &action); err != nil {
		return err
	}
	if len(args) > 0 {
		action.Output = args[0]
	}
	return util.DefaultSetup(cmd, action.Global)
}

func run(cmd *cobra.Command, _ []string) error {
	db, ok := action.Dialect.(conftypes.DBDumper)
	if !ok {
		return fmt.Errorf("%w: %s", util.ErrNoDump, action.Dialect.Name())
	}

	isDir := action.Output == "" || strings.HasSuffix(action.Output, string(os.PathSeparator)) ||
		storage.IsCloudDir(action.Output)
	if !isDir && !storage.IsCloud(action.Output) {
		if stat, err := os.Stat(action.Output); err == nil {
			isDir = stat.IsDir()
		}
	}

	if isDir {
		generated := dump.Filename{
			Database:  action.Database,
			Namespace: action.Client.Namespace,
			Ext:       database.GetExtension(db, action.Format),
			Date:      time.Now(),
		}.Generate()
		if storage.IsCloud(action.Output) {
			u, err := url.Parse(action.Output)
			if err != nil {
				return err
			}
			u.Path = path.Join(u.Path, generated)
			action.Output = u.String()
		} else {
			action.Output = filepath.Join(action.Output, generated)
		}
	} else if !cmd.Flags().Lookup(consts.FlagFormat).Changed {
		action.Format = database.DetectFormat(db, action.Output)
	}

	if err := util.CreateJob(cmd.Context(), cmd, action.Global); err != nil {
		return err
	}

	return action.Run(cmd.Context())
}

package dump

import (
	"fmt"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/clevyr/kubedb/internal/actions/dump"
	"github.com/clevyr/kubedb/internal/config"
	"github.com/clevyr/kubedb/internal/config/flags"
	"github.com/clevyr/kubedb/internal/consts"
	"github.com/clevyr/kubedb/internal/database"
	"github.com/clevyr/kubedb/internal/storage"
	"github.com/clevyr/kubedb/internal/util"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

//nolint:gochecknoglobals
var (
	action       dump.Dump
	setupOptions = util.SetupOptions{Name: "dump"}
)

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
	flags.Directory(cmd, &action.Directory)
	flags.Format(cmd, &action.Format)
	flags.IfExists(cmd, &action.IfExists)
	flags.Clean(cmd, &action.Clean)
	flags.NoOwner(cmd, &action.NoOwner)
	flags.Tables(cmd, &action.Tables)
	flags.ExcludeTable(cmd, &action.ExcludeTable)
	flags.ExcludeTableData(cmd, &action.ExcludeTableData)
	flags.Quiet(cmd, &action.Quiet)
	flags.RemoteGzip(cmd)
	flags.Spinner(cmd, &action.Spinner)
	flags.Opts(cmd)

	return cmd
}

func validArgs(cmd *cobra.Command, args []string, _ string) ([]string, cobra.ShellCompDirective) {
	if len(args) != 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	viper.Set(consts.CreateJobKey, false)
	setupOptions.NoSurvey = true
	err := preRun(cmd, args)
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	db, ok := action.Dialect.(config.DBDumper)
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
	flags.BindJobPodLabels(cmd)
	flags.BindCreateJob(cmd)
	flags.BindCreateNetworkPolicy(cmd)
	flags.BindRemoteGzip(cmd)
	action.RemoteGzip = viper.GetBool(consts.RemoteGzipKey)
	flags.BindSpinner(cmd)
	action.Spinner = viper.GetString(consts.SpinnerKey)
	flags.BindOpts(cmd)

	if len(args) > 0 {
		action.Filename = args[0]
	}
	if action.Directory != "" && action.Directory != "." {
		log.Warn().Msg("Flag --directory has been deprecated, please pass the directory as a positional arg instead.")
		action.Filename = filepath.Join(action.Directory, action.Filename)
	}

	if err := util.DefaultSetup(cmd, &action.Global, setupOptions); err != nil {
		return err
	}

	db, ok := action.Dialect.(config.DBDumper)
	if !ok {
		return fmt.Errorf("%w: %s", util.ErrNoDump, action.Dialect.Name())
	}

	var isDir bool
	if action.Filename != "" {
		if stat, err := os.Stat(action.Filename); err == nil && stat.IsDir() {
			isDir = true
		}
	}

	if action.Filename == "" || isDir || storage.IsCloudDir(action.Filename) {
		generated := dump.Filename{
			Database:  action.Database,
			Namespace: action.Client.Namespace,
			Ext:       database.GetExtension(db, action.Format),
			Date:      time.Now(),
		}.Generate()
		if storage.IsCloud(action.Filename) {
			u, err := url.Parse(action.Filename)
			if err != nil {
				return err
			}
			u.Path = path.Join(u.Path, generated)
			action.Filename = u.String()
		} else {
			action.Filename = filepath.Join(action.Filename, generated)
		}
	} else if !cmd.Flags().Lookup(consts.FormatFlag).Changed {
		action.Format = database.DetectFormat(db, action.Filename)
	}

	if err := util.CreateJob(cmd.Context(), &action.Global, setupOptions); err != nil {
		return err
	}

	return nil
}

func run(cmd *cobra.Command, _ []string) error {
	return action.Run(cmd.Context())
}

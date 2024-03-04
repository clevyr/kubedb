package dump

import (
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/clevyr/kubedb/internal/actions/dump"
	"github.com/clevyr/kubedb/internal/config/flags"
	"github.com/clevyr/kubedb/internal/consts"
	"github.com/clevyr/kubedb/internal/s3"
	"github.com/clevyr/kubedb/internal/util"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

var action dump.Dump

func New() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "dump [filename | S3 URI]",
		Aliases: []string{"d", "export"},
		Short:   "Dump a database to a sql file",
		Long: `Dump a database to a sql file.

Filenames:  
  If a filename is provided, and it does not end with a "/", then it will be used verbatim.
  Otherwise, the filename will be generated and appended to the given path.
  For example, if a dump is performed in the namespace "clevyr" with no extra flags,
  the generated filename might look like "` + dump.HelpFilename() + `".
  Similarly, if the filename is passed as "backups/", then the generated path might look like
  "backups/` + dump.HelpFilename() + `".

S3:  
  If the filename begins with "s3://", then the dump will be directly uploaded to an S3 bucket.
  S3 configuration will be loaded from the environment or from the current aws cli profile.
  Note the above section on filenames. For example, if the filename is set to "s3://clevyr-backups/dev/",
  then the resulting filename might look like "s3://clevyr-backups/dev/` + dump.HelpFilename() + `".
  The only exception is if a bucket name is provided without any sub-path (like "s3://backups"), then
  the generated filename will be appended without requiring an ending "/".
`,

		Args:              cobra.MaximumNArgs(1),
		ValidArgsFunction: validArgs,
		GroupID:           "ro",

		PreRunE: preRun,
		RunE:    run,
	}

	flags.JobPodLabels(cmd)
	flags.NoJob(cmd)
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

	return cmd
}

func validArgs(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) != 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	viper.Set(consts.NoJobKey, true)

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
	flags.BindJobPodLabels(cmd)
	flags.BindNoJob(cmd)
	flags.BindRemoteGzip(cmd)
	action.RemoteGzip = viper.GetBool(consts.RemoteGzipKey)
	flags.BindSpinner(cmd)
	action.Spinner = viper.GetString(consts.SpinnerKey)

	if len(args) > 0 {
		action.Filename = args[0]
	}

	if action.Directory != "" {
		cmd.SilenceUsage = true
		if err = os.Chdir(action.Directory); err != nil {
			return err
		}
		cmd.SilenceUsage = false
	}

	setupOptions := util.SetupOptions{Name: "dump"}
	if err := util.DefaultSetup(cmd, &action.Global, setupOptions); err != nil {
		return err
	}
	if err := util.CreateJob(cmd.Context(), &action.Global, setupOptions); err != nil {
		return err
	}

	if action.Filename == "" || strings.HasSuffix(action.Filename, string(os.PathSeparator)) || s3.IsS3Dir(action.Filename) {
		generated := dump.Filename{
			Database:  action.Database,
			Namespace: action.Client.Namespace,
			Ext:       action.Dialect.DumpExtension(action.Format),
			Date:      time.Now(),
		}.Generate()
		if s3.IsS3(action.Filename) {
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
		action.Format = action.Dialect.FormatFromFilename(action.Filename)
	}

	return nil
}

func run(cmd *cobra.Command, args []string) (err error) {
	return action.Run(cmd.Context())
}

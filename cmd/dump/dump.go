package dump

import (
	"os"
	"strconv"

	"github.com/clevyr/kubedb/internal/actions/dump"
	"github.com/clevyr/kubedb/internal/config"
	"github.com/clevyr/kubedb/internal/config/flags"
	"github.com/clevyr/kubedb/internal/util"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

var action dump.Dump

func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "dump [filename]",
		Aliases: []string{"d", "export"},
		Short:   "Dump a database to a sql file",
		Long: `Dump a database to a sql file.

If no filename is provided, the filename will be generated.
For example, if a dump is performed in the namespace "clevyr" with no extra flags,
the generated filename might look like "` + dump.HelpFilename() + `"`,

		Args:              cobra.MaximumNArgs(1),
		ValidArgsFunction: validArgs,
		GroupID:           "ro",

		PreRunE: preRun,
		RunE:    run,

		Annotations: map[string]string{
			"access": strconv.Itoa(config.ReadOnly),
		},
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

var setupOptions = util.SetupOptions{Name: "dump"}

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
	flags.BindJobPodLabels(cmd)
	flags.BindNoJob(cmd)
	flags.BindRemoteGzip(cmd)
	action.RemoteGzip = viper.GetBool("remote-gzip")
	flags.BindSpinner(cmd)
	action.Spinner = viper.GetString("spinner.name")

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

	if err := util.DefaultSetup(cmd, &action.Global, setupOptions); err != nil {
		return err
	}

	if action.Filename != "" && !cmd.Flags().Lookup("format").Changed {
		action.Format = action.Dialect.FormatFromFilename(action.Filename)
	}

	return nil
}

func run(cmd *cobra.Command, args []string) (err error) {
	defer func() {
		util.Teardown(cmd, &action.Global)
	}()
	return action.Run(cmd.Context())
}

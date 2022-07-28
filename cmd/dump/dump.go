package dump

import (
	"github.com/clevyr/kubedb/internal/actions/dump"
	"github.com/clevyr/kubedb/internal/config"
	"github.com/clevyr/kubedb/internal/config/flags"
	"github.com/clevyr/kubedb/internal/util"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"os"
	"strconv"
)

var Command = &cobra.Command{
	Use:     "dump [filename]",
	Aliases: []string{"d", "export"},
	Short:   "dump a database to a sql file",
	Long: `The "dump" command dumps a running database to a sql file.

If no filename is provided, the filename will be generated.
For example, if a dump is performed in the namespace "clevyr" with no extra flags,
the generated filename might look like "` + dump.HelpFilename() + `"`,

	Args:              cobra.MaximumNArgs(1),
	ValidArgsFunction: validArgs,

	PreRunE: preRun,
	RunE:    run,

	Annotations: map[string]string{
		"access": strconv.Itoa(config.ReadOnly),
	},
}

var action dump.Dump

func init() {
	flags.Directory(Command, &action.Directory)
	flags.Format(Command, &action.Format)
	flags.IfExists(Command, &action.IfExists)
	flags.Clean(Command, &action.Clean)
	flags.NoOwner(Command, &action.NoOwner)
	flags.Tables(Command, &action.Tables)
	flags.ExcludeTable(Command, &action.ExcludeTable)
	flags.ExcludeTableData(Command, &action.ExcludeTableData)
	flags.Quiet(Command, &action.Quiet)
}

func validArgs(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) != 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	err := preRun(cmd, args)
	if err != nil {
		return []string{"sql", "sql.gz", "dmp", "archive", "archive.gz"}, cobra.ShellCompDirectiveFilterFileExt
	}

	var exts []string
	for _, ext := range action.Dialect.Formats() {
		exts = append(exts, ext[1:])
	}
	return exts, cobra.ShellCompDirectiveFilterFileExt
}

func preRun(cmd *cobra.Command, args []string) (err error) {
	if err := viper.Unmarshal(&action); err != nil {
		return err
	}

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

	if err := util.DefaultSetup(cmd, &action.Global); err != nil {
		return err
	}

	if action.Filename != "" && !cmd.Flags().Lookup("format").Changed {
		action.Format = action.Dialect.FormatFromFilename(action.Filename)
	}

	return nil
}

func run(cmd *cobra.Command, args []string) (err error) {
	return action.Run()
}

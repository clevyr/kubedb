package flags

import (
	"strings"

	"gabe565.com/utils/must"
	"github.com/clevyr/kubedb/internal/completion"
	"github.com/clevyr/kubedb/internal/consts"
	"github.com/clevyr/kubedb/internal/database"
	"github.com/clevyr/kubedb/internal/database/sqlformat"
	"github.com/spf13/cobra"
)

func Dialect(cmd *cobra.Command) {
	cmd.PersistentFlags().String(consts.FlagDialect, "", "Database dialect. (one of "+strings.Join(database.Names(), ", ")+") (default discovered)")
	must.Must(cmd.RegisterFlagCompletionFunc(consts.FlagDialect,
		func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
			return database.Names(), cobra.ShellCompDirectiveNoFileComp
		}),
	)
}

func Format(cmd *cobra.Command) {
	format := sqlformat.Gzip
	cmd.Flags().VarP(&format, consts.FlagFormat, "F", `Output file format (one of gzip, custom, plain)`)
	must.Must(cmd.RegisterFlagCompletionFunc(consts.FlagFormat,
		func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
			return []string{
				sqlformat.Gzip.String(),
				sqlformat.Plain.String(),
				sqlformat.Custom.String(),
			}, cobra.ShellCompDirectiveNoFileComp
		}),
	)
}

func Port(cmd *cobra.Command) {
	cmd.PersistentFlags().Uint16(consts.FlagPort, 0, "Database port (default discovered)")
	must.Must(cmd.RegisterFlagCompletionFunc(consts.FlagPort, cobra.NoFileCompletions))
}

func Database(cmd *cobra.Command) {
	cmd.Flags().StringP(consts.FlagDBName, "d", "", "Database name to use (default discovered)")
	must.Must(cmd.RegisterFlagCompletionFunc(consts.FlagDBName, completion.DatabasesList))
}

func Username(cmd *cobra.Command) {
	cmd.Flags().StringP(consts.FlagUsername, "U", "", "Database username (default discovered)")
	must.Must(cmd.RegisterFlagCompletionFunc(consts.FlagUsername, cobra.NoFileCompletions))
}

func Password(cmd *cobra.Command) {
	cmd.Flags().StringP(consts.FlagPassword, "p", "", "Database password (default discovered)")
	must.Must(cmd.RegisterFlagCompletionFunc(consts.FlagPassword, cobra.NoFileCompletions))
}

func SingleTransaction(cmd *cobra.Command) {
	cmd.Flags().BoolP(consts.FlagSingleTransaction, "1", true, "Restore as a single transaction")
}

func Clean(cmd *cobra.Command) {
	cmd.Flags().BoolP(consts.FlagClean, "c", true, "Clean (drop) database objects before recreating")
	must.Must(cmd.RegisterFlagCompletionFunc(consts.FlagClean, completion.BoolCompletion))
}

func IfExists(cmd *cobra.Command) {
	cmd.Flags().Bool(consts.FlagIfExists, true, "Use IF EXISTS when dropping objects")
	must.Must(cmd.RegisterFlagCompletionFunc(consts.FlagIfExists, completion.BoolCompletion))
}

func NoOwner(cmd *cobra.Command) {
	cmd.Flags().BoolP(consts.FlagNoOwner, "O", true, "Skip restoration of object ownership in plain-text format")
	must.Must(cmd.RegisterFlagCompletionFunc(consts.FlagNoOwner, completion.BoolCompletion))
}

func Tables(cmd *cobra.Command) {
	cmd.Flags().StringSliceP(consts.FlagTable, "t", nil, "Dump the specified table(s) only")
	must.Must(cmd.RegisterFlagCompletionFunc(consts.FlagTable, completion.TablesList))
}

func ExcludeTable(cmd *cobra.Command) {
	cmd.Flags().StringSliceP(consts.FlagExcludeTable, "T", nil, "Do NOT dump the specified table(s)")
	must.Must(cmd.RegisterFlagCompletionFunc(consts.FlagExcludeTable, completion.TablesList))
}

func ExcludeTableData(cmd *cobra.Command) {
	cmd.Flags().StringSliceP(consts.FlagExcludeTableData, "D", nil, "Do NOT dump data for the specified table(s)")
	must.Must(cmd.RegisterFlagCompletionFunc(consts.FlagExcludeTableData, completion.TablesList))
}

func Analyze(cmd *cobra.Command) {
	cmd.Flags().Bool(consts.FlagAnalyze, true, "Run an analyze query after restore")
	must.Must(cmd.RegisterFlagCompletionFunc(consts.FlagAnalyze, completion.BoolCompletion))
}

func HaltOnError(cmd *cobra.Command) {
	cmd.Flags().Bool(consts.FlagHaltOnError, true, "Halt on error (Postgres only)")
	must.Must(cmd.RegisterFlagCompletionFunc(consts.FlagHaltOnError, completion.BoolCompletion))
}

func Opts(cmd *cobra.Command) {
	cmd.Flags().String(consts.FlagOpts, "", "Additional options to pass to the database client command")
	must.Must(cmd.RegisterFlagCompletionFunc(consts.FlagOpts, cobra.NoFileCompletions))
}

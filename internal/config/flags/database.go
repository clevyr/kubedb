package flags

import (
	"log/slog"
	"os"
	"strings"

	"gabe565.com/utils/must"
	"github.com/clevyr/kubedb/internal/config"
	"github.com/clevyr/kubedb/internal/consts"
	"github.com/clevyr/kubedb/internal/database"
	"github.com/clevyr/kubedb/internal/database/sqlformat"
	"github.com/clevyr/kubedb/internal/kubernetes"
	"github.com/clevyr/kubedb/internal/util"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func Dialect(cmd *cobra.Command) {
	cmd.PersistentFlags().String(consts.DialectFlag, "", "Database dialect. (one of "+strings.Join(database.Names(), ", ")+") (default discovered)")
	must.Must(cmd.RegisterFlagCompletionFunc(consts.DialectFlag,
		func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
			return database.Names(), cobra.ShellCompDirectiveNoFileComp
		}),
	)
}

func Format(cmd *cobra.Command, p *sqlformat.Format) {
	*p = sqlformat.Gzip
	cmd.Flags().VarP(p, consts.FormatFlag, "F", `Output file format (one of gzip, custom, plain)`)
	must.Must(cmd.RegisterFlagCompletionFunc(consts.FormatFlag,
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
	cmd.PersistentFlags().Uint16(consts.PortFlag, 0, "Database port (default discovered)")
	must.Must(cmd.RegisterFlagCompletionFunc(consts.PortFlag, cobra.NoFileCompletions))
}

func Database(cmd *cobra.Command) {
	cmd.Flags().StringP(consts.DbnameFlag, "d", "", "Database name to use (default discovered)")
	must.Must(cmd.RegisterFlagCompletionFunc(consts.DbnameFlag, listDatabases))
}

func Username(cmd *cobra.Command) {
	cmd.Flags().StringP(consts.UsernameFlag, "U", "", "Database username (default discovered)")
	must.Must(cmd.RegisterFlagCompletionFunc(consts.UsernameFlag, cobra.NoFileCompletions))
}

func Password(cmd *cobra.Command) {
	cmd.Flags().StringP(consts.PasswordFlag, "p", "", "Database password (default discovered)")
	must.Must(cmd.RegisterFlagCompletionFunc(consts.PasswordFlag, cobra.NoFileCompletions))
}

func SingleTransaction(cmd *cobra.Command, p *bool) {
	cmd.Flags().BoolVarP(p, consts.SingleTransactionFlag, "1", true, "Restore as a single transaction")
}

func Clean(cmd *cobra.Command, p *bool) {
	cmd.Flags().BoolVarP(p, consts.CleanFlag, "c", true, "Clean (drop) database objects before recreating")
	must.Must(cmd.RegisterFlagCompletionFunc(consts.CleanFlag, util.BoolCompletion))
}

func IfExists(cmd *cobra.Command, p *bool) {
	cmd.Flags().BoolVar(p, consts.IfExistsFlag, true, "Use IF EXISTS when dropping objects")
	must.Must(cmd.RegisterFlagCompletionFunc(consts.IfExistsFlag, util.BoolCompletion))
}

func NoOwner(cmd *cobra.Command, p *bool) {
	cmd.Flags().BoolVarP(p, consts.NoOwnerFlag, "O", true, "Skip restoration of object ownership in plain-text format")
	must.Must(cmd.RegisterFlagCompletionFunc(consts.NoOwnerFlag, util.BoolCompletion))
}

func Tables(cmd *cobra.Command, p *[]string) {
	cmd.Flags().StringSliceVarP(p, consts.TableFlag, "t", nil, "Dump the specified table(s) only")
	must.Must(cmd.RegisterFlagCompletionFunc(consts.TableFlag, listTables))
}

func ExcludeTable(cmd *cobra.Command, p *[]string) {
	cmd.Flags().StringSliceVarP(p, consts.ExcludeTableFlag, "T", nil, "Do NOT dump the specified table(s)")
	must.Must(cmd.RegisterFlagCompletionFunc(consts.ExcludeTableFlag, listTables))
}

func ExcludeTableData(cmd *cobra.Command, p *[]string) {
	cmd.Flags().StringSliceVarP(p, consts.ExcludeTableDataFlag, "D", nil, "Do NOT dump data for the specified table(s)")
	must.Must(cmd.RegisterFlagCompletionFunc(consts.ExcludeTableDataFlag, listTables))
}

func Analyze(cmd *cobra.Command) {
	cmd.Flags().Bool(consts.AnalyzeFlag, true, "Run an analyze query after restore")
	must.Must(cmd.RegisterFlagCompletionFunc(consts.AnalyzeFlag, util.BoolCompletion))
}

func BindAnalyze(cmd *cobra.Command) {
	must.Must(viper.BindPFlag(consts.AnalyzeKey, cmd.Flags().Lookup(consts.AnalyzeFlag)))
}

func HaltOnError(cmd *cobra.Command) {
	cmd.Flags().Bool(consts.HaltOnErrorFlag, true, "Halt on error (Postgres only)")
	must.Must(cmd.RegisterFlagCompletionFunc(consts.HaltOnErrorFlag, util.BoolCompletion))
}

func BindHaltOnError(cmd *cobra.Command) {
	must.Must(viper.BindPFlag(consts.HaltOnErrorKey, cmd.Flags().Lookup(consts.HaltOnErrorFlag)))
}

func Opts(cmd *cobra.Command) {
	cmd.Flags().String(consts.OptsFlag, "", "Additional options to pass to the database client command")
	must.Must(cmd.RegisterFlagCompletionFunc(consts.OptsFlag, cobra.NoFileCompletions))
}

func BindOpts(cmd *cobra.Command) {
	must.Must(viper.BindPFlag(consts.OptsKey, cmd.Flags().Lookup(consts.OptsFlag)))
}

func listTables(cmd *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
	conf := config.Exec{DisableHeaders: true}

	viper.Set(consts.CreateJobKey, false)
	err := util.DefaultSetup(cmd, &conf.Global, util.SetupOptions{NoSurvey: true})
	if err != nil {
		slog.Error("Setup failed", "error", err)
		return nil, cobra.ShellCompDirectiveError
	}

	db, ok := conf.Dialect.(config.DBTableLister)
	if !ok {
		slog.Error("Dialect does not support listing tables", "error", err)
		return nil, cobra.ShellCompDirectiveError
	}

	conf.Command = db.TableListQuery()
	return queryInDatabase(cmd, conf)
}

func listDatabases(cmd *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
	conf := config.Exec{DisableHeaders: true}

	viper.Set(consts.CreateJobKey, false)
	err := util.DefaultSetup(cmd, &conf.Global, util.SetupOptions{NoSurvey: true})
	if err != nil {
		slog.Error("Setup failed", "error", err)
		return nil, cobra.ShellCompDirectiveError
	}

	db, ok := conf.Dialect.(config.DBDatabaseLister)
	if !ok {
		slog.Error("Dialect does not support listing databases", "error", err)
		return nil, cobra.ShellCompDirectiveError
	}

	conf.Command = db.DatabaseListQuery()
	return queryInDatabase(cmd, conf)
}

func queryInDatabase(cmd *cobra.Command, conf config.Exec) ([]string, cobra.ShellCompDirective) {
	db, ok := conf.Dialect.(config.DBExecer)
	if !ok {
		slog.Error("Dialect does not support exec", "name", conf.Dialect.Name())
		return nil, cobra.ShellCompDirectiveError
	}

	var buf strings.Builder
	if err := conf.Client.Exec(cmd.Context(), kubernetes.ExecOptions{
		Pod:    conf.DBPod,
		Cmd:    db.ExecCommand(conf).String(),
		Stdout: &buf,
		Stderr: os.Stderr,
	}); err != nil {
		slog.Error("Exec failed", "error", err)
		return nil, cobra.ShellCompDirectiveError
	}

	names := strings.Split(buf.String(), "\n")
	return names, cobra.ShellCompDirectiveNoFileComp
}

package flags

import (
	"os"
	"strings"

	"github.com/clevyr/kubedb/internal/config"
	"github.com/clevyr/kubedb/internal/consts"
	"github.com/clevyr/kubedb/internal/database/mariadb"
	"github.com/clevyr/kubedb/internal/database/mongodb"
	"github.com/clevyr/kubedb/internal/database/postgres"
	"github.com/clevyr/kubedb/internal/database/sqlformat"
	"github.com/clevyr/kubedb/internal/kubernetes"
	"github.com/clevyr/kubedb/internal/util"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func Dialect(cmd *cobra.Command) {
	cmd.PersistentFlags().String(consts.DialectFlag, "", "Database dialect. One of (postgres|mariadb|mongodb) (default discovered)")
	err := cmd.RegisterFlagCompletionFunc(
		consts.DialectFlag,
		func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return []string{
				postgres.Postgres{}.Name(),
				mariadb.MariaDB{}.Name(),
				mongodb.MongoDB{}.Name(),
			}, cobra.ShellCompDirectiveNoFileComp
		})
	if err != nil {
		panic(err)
	}
}

func Format(cmd *cobra.Command, p *sqlformat.Format) {
	*p = sqlformat.Gzip
	cmd.Flags().VarP(p, consts.FormatFlag, "F", `Output file format One of (gzip|custom|plain)`)
	err := cmd.RegisterFlagCompletionFunc(
		consts.FormatFlag,
		func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return []string{
				sqlformat.Gzip.String(),
				sqlformat.Plain.String(),
				sqlformat.Custom.String(),
			}, cobra.ShellCompDirectiveNoFileComp
		})
	if err != nil {
		panic(err)
	}
}

func Port(cmd *cobra.Command) {
	cmd.PersistentFlags().Uint16(consts.PortFlag, 0, "Database port (default discovered)")
}

func Database(cmd *cobra.Command) {
	cmd.Flags().StringP(consts.DbnameFlag, "d", "", "Database name to use (default discovered)")
	err := cmd.RegisterFlagCompletionFunc(consts.DbnameFlag, listDatabases)
	if err != nil {
		panic(err)
	}
}

func Username(cmd *cobra.Command) {
	cmd.Flags().StringP(consts.UsernameFlag, "U", "", "Database username (default discovered)")
}

func Password(cmd *cobra.Command) {
	cmd.Flags().StringP(consts.PasswordFlag, "p", "", "Database password (default discovered)")
}

func SingleTransaction(cmd *cobra.Command, p *bool) {
	cmd.Flags().BoolVarP(p, consts.SingleTransactionFlag, "1", true, "Restore as a single transaction")
}

func Clean(cmd *cobra.Command, p *bool) {
	cmd.Flags().BoolVarP(p, consts.CleanFlag, "c", true, "Clean (drop) database objects before recreating")
}

func IfExists(cmd *cobra.Command, p *bool) {
	cmd.Flags().BoolVar(p, consts.IfExistsFlag, true, "Use IF EXISTS when dropping objects")
}

func NoOwner(cmd *cobra.Command, p *bool) {
	cmd.Flags().BoolVarP(p, consts.NoOwnerFlag, "O", true, "Skip restoration of object ownership in plain-text format")
}

func Tables(cmd *cobra.Command, p *[]string) {
	cmd.Flags().StringSliceVarP(p, consts.TableFlag, "t", []string{}, "Dump the specified table(s) only")
	err := cmd.RegisterFlagCompletionFunc(consts.TableFlag, listTables)
	if err != nil {
		panic(err)
	}
}

func ExcludeTable(cmd *cobra.Command, p *[]string) {
	cmd.Flags().StringSliceVarP(p, consts.ExcludeTableFlag, "T", []string{}, "Do NOT dump the specified table(s)")
	err := cmd.RegisterFlagCompletionFunc(consts.ExcludeTableFlag, listTables)
	if err != nil {
		panic(err)
	}
}

func ExcludeTableData(cmd *cobra.Command, p *[]string) {
	cmd.Flags().StringSliceVarP(p, consts.ExcludeTableDataFlag, "D", []string{}, "Do NOT dump data for the specified table(s)")
	err := cmd.RegisterFlagCompletionFunc(consts.ExcludeTableDataFlag, listTables)
	if err != nil {
		panic(err)
	}
}

func Analyze(cmd *cobra.Command) {
	cmd.Flags().Bool(consts.AnalyzeFlag, true, "Run an analyze query after restore")
}

func BindAnalyze(cmd *cobra.Command) {
	if err := viper.BindPFlag(consts.AnalyzeKey, cmd.Flags().Lookup(consts.AnalyzeFlag)); err != nil {
		panic(err)
	}
}

func HaltOnError(cmd *cobra.Command) {
	cmd.Flags().Bool(consts.HaltOnErrorFlag, true, "Halt on error (Postgres only)")
}

func BindHaltOnError(cmd *cobra.Command) {
	if err := viper.BindPFlag(consts.HaltOnErrorKey, cmd.Flags().Lookup(consts.HaltOnErrorFlag)); err != nil {
		panic(err)
	}
}

func listTables(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	conf := config.Exec{DisableHeaders: true}

	viper.Set(consts.NoJobKey, true)
	err := util.DefaultSetup(cmd, &conf.Global, util.SetupOptions{NoSurvey: true})
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}
	conf.Command = conf.Dialect.ListTablesQuery()
	return queryInDatabase(cmd, args, conf)
}

func listDatabases(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	conf := config.Exec{DisableHeaders: true}

	viper.Set(consts.NoJobKey, true)
	err := util.DefaultSetup(cmd, &conf.Global, util.SetupOptions{NoSurvey: true})
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}
	conf.Command = conf.Dialect.ListDatabasesQuery()
	return queryInDatabase(cmd, args, conf)
}

func queryInDatabase(cmd *cobra.Command, args []string, conf config.Exec) ([]string, cobra.ShellCompDirective) {
	var buf strings.Builder
	if err := conf.Client.Exec(cmd.Context(), kubernetes.ExecOptions{
		Pod:    conf.DbPod,
		Cmd:    conf.Dialect.ExecCommand(conf).String(),
		Stdout: &buf,
		Stderr: os.Stderr,
	}); err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	names := strings.Split(buf.String(), "\n")
	return names, cobra.ShellCompDirectiveNoFileComp
}

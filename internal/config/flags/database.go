package flags

import (
	"os"
	"strings"
	"time"

	"github.com/clevyr/kubedb/internal/config"
	"github.com/clevyr/kubedb/internal/database/dialect"
	"github.com/clevyr/kubedb/internal/database/sqlformat"
	"github.com/clevyr/kubedb/internal/util"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func Dialect(cmd *cobra.Command) {
	cmd.PersistentFlags().String("grammar", "", "Database dialect. Detected if not set. (postgres, mariadb, mongodb)")
	err := cmd.PersistentFlags().MarkDeprecated("grammar", "please use --dialect instead")
	if err != nil {
		panic(err)
	}

	cmd.PersistentFlags().String("dialect", "", "Database dialect. Detected if not set. (postgres, mariadb, mongodb)")
	err = cmd.RegisterFlagCompletionFunc(
		"dialect",
		func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return []string{
				dialect.Postgres{}.Name(),
				dialect.MariaDB{}.Name(),
				dialect.MongoDB{}.Name(),
			}, cobra.ShellCompDirectiveNoFileComp
		})
	if err != nil {
		panic(err)
	}
}

func Format(cmd *cobra.Command, p *sqlformat.Format) {
	*p = sqlformat.Gzip
	cmd.Flags().VarP(p, "format", "F", "Output file format ([g]zip, [c]ustom, [p]lain)")
	err := cmd.RegisterFlagCompletionFunc(
		"format",
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

func Database(cmd *cobra.Command) {
	cmd.Flags().StringP("dbname", "d", "", "Database name to connect to")
	err := cmd.RegisterFlagCompletionFunc("dbname", listDatabases)
	if err != nil {
		panic(err)
	}
}

func Username(cmd *cobra.Command) {
	cmd.Flags().StringP("username", "U", "", "Database username")
}

func Password(cmd *cobra.Command) {
	cmd.Flags().StringP("password", "p", "", "Database password")
}

func SingleTransaction(cmd *cobra.Command, p *bool) {
	cmd.Flags().BoolVarP(p, "single-transaction", "1", true, "Restore as a single transaction")
}

func Clean(cmd *cobra.Command, p *bool) {
	cmd.Flags().BoolVarP(p, "clean", "c", true, "Clean (drop) database objects before recreating")
}

func IfExists(cmd *cobra.Command, p *bool) {
	cmd.Flags().BoolVar(p, "if-exists", true, "Use IF EXISTS when dropping objects")
}

func NoOwner(cmd *cobra.Command, p *bool) {
	cmd.Flags().BoolVarP(p, "no-owner", "O", true, "Skip restoration of object ownership in plain-text format")
}

func Tables(cmd *cobra.Command, p *[]string) {
	cmd.Flags().StringSliceVarP(p, "table", "t", []string{}, "Dump the specified table(s) only")
	err := cmd.RegisterFlagCompletionFunc("table", listTables)
	if err != nil {
		panic(err)
	}
}

func ExcludeTable(cmd *cobra.Command, p *[]string) {
	cmd.Flags().StringSliceVarP(p, "exclude-table", "T", []string{}, "Do NOT dump the specified table(s)")
	err := cmd.RegisterFlagCompletionFunc("exclude-table", listTables)
	if err != nil {
		panic(err)
	}
}

func ExcludeTableData(cmd *cobra.Command, p *[]string) {
	cmd.Flags().StringSliceVarP(p, "exclude-table-data", "D", []string{}, "Do NOT dump data for the specified table(s)")
	err := cmd.RegisterFlagCompletionFunc("exclude-table-data", listTables)
	if err != nil {
		panic(err)
	}
}

func Analyze(cmd *cobra.Command) {
	cmd.Flags().Bool("analyze", true, "Run an analyze query after restore")
}

func BindAnalyze(cmd *cobra.Command) {
	if err := viper.BindPFlag("restore.analyze", cmd.Flags().Lookup("analyze")); err != nil {
		panic(err)
	}
}

func listTables(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	conf := config.Exec{DisableHeaders: true}

	err := util.DefaultSetup(cmd, &conf.Global, util.SetupOptions{DisableJob: true})
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}
	conf.Command = conf.Dialect.ListTablesQuery()
	return queryInDatabase(cmd, args, conf)
}

func listDatabases(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	conf := config.Exec{DisableHeaders: true}

	err := util.DefaultSetup(cmd, &conf.Global, util.SetupOptions{DisableJob: true})
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}
	conf.Command = conf.Dialect.ListDatabasesQuery()
	return queryInDatabase(cmd, args, conf)
}

func queryInDatabase(cmd *cobra.Command, args []string, conf config.Exec) ([]string, cobra.ShellCompDirective) {
	var buf strings.Builder
	sqlCmd := conf.Dialect.ExecCommand(conf)
	err := conf.Client.Exec(cmd.Context(), conf.DbPod, sqlCmd.String(), nil, &buf, os.Stderr, false, nil, 5*time.Second)
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	names := strings.Split(buf.String(), "\n")
	return names, cobra.ShellCompDirectiveNoFileComp
}

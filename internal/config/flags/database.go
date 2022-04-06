package flags

import (
	"github.com/clevyr/kubedb/internal/config"
	"github.com/clevyr/kubedb/internal/database/grammar"
	"github.com/clevyr/kubedb/internal/database/sqlformat"
	"github.com/clevyr/kubedb/internal/util"
	"github.com/spf13/cobra"
	"os"
	"strings"
)

func Grammar(cmd *cobra.Command) {
	cmd.PersistentFlags().String("grammar", "", "database grammar. detected if not set. (postgres, mariadb)")
	err := cmd.RegisterFlagCompletionFunc(
		"grammar",
		func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return []string{
				grammar.Postgres{}.Name(),
				grammar.MariaDB{}.Name(),
			}, cobra.ShellCompDirectiveNoFileComp
		})
	if err != nil {
		panic(err)
	}
}

func Format(cmd *cobra.Command, p *sqlformat.Format) {
	*p = sqlformat.Gzip
	cmd.Flags().VarP(p, "format", "F", "output file format ([g]zip, [c]ustom, [p]lain)")
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
	cmd.PersistentFlags().StringP("dbname", "d", "", "database name to connect to")
	err := cmd.RegisterFlagCompletionFunc("dbname", listDatabases)
	if err != nil {
		panic(err)
	}
}

func Username(cmd *cobra.Command) {
	cmd.PersistentFlags().StringP("username", "U", "", "database username")
}

func Password(cmd *cobra.Command) {
	cmd.PersistentFlags().StringP("password", "p", "", "database password")
}

func SingleTransaction(cmd *cobra.Command, p *bool) {
	cmd.Flags().BoolVarP(p, "single-transaction", "1", true, "restore as a single transaction")
}

func Clean(cmd *cobra.Command, p *bool) {
	cmd.Flags().BoolVarP(p, "clean", "c", true, "clean (drop) database objects before recreating")
}

func IfExists(cmd *cobra.Command, p *bool) {
	cmd.Flags().BoolVar(p, "if-exists", true, "use IF EXISTS when dropping objects")
}

func NoOwner(cmd *cobra.Command, p *bool) {
	cmd.Flags().BoolVarP(p, "no-owner", "O", true, "skip restoration of object ownership in plain-text format")
}

func Tables(cmd *cobra.Command, p *[]string) {
	cmd.Flags().StringSliceVarP(p, "table", "t", []string{}, "dump the specified table(s) only")
	err := cmd.RegisterFlagCompletionFunc("table", listTables)
	if err != nil {
		panic(err)
	}
}

func ExcludeTable(cmd *cobra.Command, p *[]string) {
	cmd.Flags().StringSliceVarP(p, "exclude-table", "T", []string{}, "do NOT dump the specified table(s)")
	err := cmd.RegisterFlagCompletionFunc("exclude-table", listTables)
	if err != nil {
		panic(err)
	}
}

func ExcludeTableData(cmd *cobra.Command, p *[]string) {
	cmd.Flags().StringSliceVarP(p, "exclude-table-data", "D", []string{}, "do NOT dump data for the specified table(s)")
	err := cmd.RegisterFlagCompletionFunc("exclude-table-data", listTables)
	if err != nil {
		panic(err)
	}
}

func listTables(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	conf := config.Exec{DisableHeaders: true}
	err := util.DefaultSetup(cmd, &conf.Global)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	return queryInDatabase(cmd, args, conf, conf.Grammar.ListTablesQuery())
}

func listDatabases(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	conf := config.Exec{DisableHeaders: true}
	err := util.DefaultSetup(cmd, &conf.Global)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	return queryInDatabase(cmd, args, conf, conf.Grammar.ListDatabasesQuery())
}

func queryInDatabase(cmd *cobra.Command, args []string, conf config.Exec, query string) ([]string, cobra.ShellCompDirective) {
	r := strings.NewReader(query)
	var buf strings.Builder
	sqlCmd := conf.Grammar.ExecCommand(conf)
	err := conf.Client.Exec(conf.Pod, sqlCmd.String(), r, &buf, os.Stderr, false, nil)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	names := strings.Split(buf.String(), "\n")
	return names, cobra.ShellCompDirectiveNoFileComp
}

package flags

import (
	"os"
	"strings"

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

	if err := viper.BindPFlag("dialect", cmd.PersistentFlags().Lookup("dialect")); err != nil {
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

	if err := viper.BindPFlag("format", cmd.Flags().Lookup("format")); err != nil {
		panic(err)
	}
}

func Database(cmd *cobra.Command) {
	cmd.PersistentFlags().StringP("dbname", "d", "", "Database name to connect to")
	err := cmd.RegisterFlagCompletionFunc("dbname", listDatabases)
	if err != nil {
		panic(err)
	}

	if err := viper.BindPFlag("dbname", cmd.PersistentFlags().Lookup("dbname")); err != nil {
		panic(err)
	}
}

func Username(cmd *cobra.Command) {
	cmd.PersistentFlags().StringP("username", "U", "", "Database username")
	if err := viper.BindPFlag("username", cmd.PersistentFlags().Lookup("username")); err != nil {
		panic(err)
	}
}

func Password(cmd *cobra.Command) {
	cmd.PersistentFlags().StringP("password", "p", "", "Database password")
	if err := viper.BindPFlag("password", cmd.PersistentFlags().Lookup("password")); err != nil {
		panic(err)
	}
}

func SingleTransaction(cmd *cobra.Command, p *bool) {
	cmd.Flags().BoolVarP(p, "single-transaction", "1", true, "Restore as a single transaction")
	if err := viper.BindPFlag("single-transaction", cmd.Flags().Lookup("single-transaction")); err != nil {
		panic(err)
	}
}

func Clean(cmd *cobra.Command, p *bool) {
	cmd.Flags().BoolVarP(p, "clean", "c", true, "Clean (drop) database objects before recreating")
	if err := viper.BindPFlag("clean", cmd.Flags().Lookup("clean")); err != nil {
		panic(err)
	}
}

func IfExists(cmd *cobra.Command, p *bool) {
	cmd.Flags().BoolVar(p, "if-exists", true, "Use IF EXISTS when dropping objects")
	if err := viper.BindPFlag("if-exists", cmd.Flags().Lookup("if-exists")); err != nil {
		panic(err)
	}
}

func NoOwner(cmd *cobra.Command, p *bool) {
	cmd.Flags().BoolVarP(p, "no-owner", "O", true, "Skip restoration of object ownership in plain-text format")
	if err := viper.BindPFlag("no-owner", cmd.Flags().Lookup("no-owner")); err != nil {
		panic(err)
	}
}

func Tables(cmd *cobra.Command, p *[]string) {
	cmd.Flags().StringSliceVarP(p, "table", "t", []string{}, "Dump the specified table(s) only")
	err := cmd.RegisterFlagCompletionFunc("table", listTables)
	if err != nil {
		panic(err)
	}
	if err := viper.BindPFlag("table", cmd.Flags().Lookup("table")); err != nil {
		panic(err)
	}
}

func ExcludeTable(cmd *cobra.Command, p *[]string) {
	cmd.Flags().StringSliceVarP(p, "exclude-table", "T", []string{}, "Do NOT dump the specified table(s)")
	err := cmd.RegisterFlagCompletionFunc("exclude-table", listTables)
	if err != nil {
		panic(err)
	}
	if err := viper.BindPFlag("exclude-table", cmd.Flags().Lookup("exclude-table")); err != nil {
		panic(err)
	}
}

func ExcludeTableData(cmd *cobra.Command, p *[]string) {
	cmd.Flags().StringSliceVarP(p, "exclude-table-data", "D", []string{}, "Do NOT dump data for the specified table(s)")
	err := cmd.RegisterFlagCompletionFunc("exclude-table-data", listTables)
	if err != nil {
		panic(err)
	}
	if err := viper.BindPFlag("exclude-table-data", cmd.Flags().Lookup("exclude-table-data")); err != nil {
		panic(err)
	}
}

func listTables(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	conf := config.Exec{DisableHeaders: true}
	err := util.DefaultSetup(cmd, &conf.Global)
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}
	conf.Command = conf.Dialect.ListTablesQuery()
	return queryInDatabase(cmd, args, conf)
}

func listDatabases(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	conf := config.Exec{DisableHeaders: true}
	err := util.DefaultSetup(cmd, &conf.Global)
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}
	conf.Command = conf.Dialect.ListDatabasesQuery()
	return queryInDatabase(cmd, args, conf)
}

func queryInDatabase(cmd *cobra.Command, args []string, conf config.Exec) ([]string, cobra.ShellCompDirective) {
	var buf strings.Builder
	sqlCmd := conf.Dialect.ExecCommand(conf)
	err := conf.Client.Exec(cmd.Context(), conf.Pod, sqlCmd.String(), nil, &buf, os.Stderr, false, nil)
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	names := strings.Split(buf.String(), "\n")
	return names, cobra.ShellCompDirectiveNoFileComp
}

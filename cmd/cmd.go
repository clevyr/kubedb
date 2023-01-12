package cmd

import (
	"fmt"
	"github.com/clevyr/kubedb/cmd/dump"
	"github.com/clevyr/kubedb/cmd/exec"
	"github.com/clevyr/kubedb/cmd/port_forward"
	"github.com/clevyr/kubedb/cmd/restore"
	"github.com/clevyr/kubedb/cmd/ui"
	"github.com/clevyr/kubedb/internal/config/flags"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
	"strings"
)

var (
	Version = "next"
	Commit  = ""
)

var Command = &cobra.Command{
	Use:               "kubedb",
	Short:             "interact with a database inside of Kubernetes",
	Version:           buildVersion(),
	DisableAutoGenTag: true,
	Long: `kubedb is a command to interact with a database running in a Kubernetes cluster.

Multiple database types (referred to as the "dialect") are supported.
If the dialect is not configured via flag, it will be detected dynamically.

Supported Database Dialects:
  - PostgreSQL
  - MariaDB

If not configured via flag, some configuration variables will be loaded from the target pod's env vars.

Dynamic Env Var Variables:
  - Database
  - Username (fallback value: "postgres" if PostgreSQL, "mariadb" if MariaDB, "root" if MongoDB)
  - Password
`,

	PersistentPreRunE: preRun,
}

func preRun(cmd *cobra.Command, args []string) error {
	kubeconfig, err := cmd.Flags().GetString("kubeconfig")
	if err != nil {
		panic(err)
	}
	if kubeconfig == "$HOME" || strings.HasPrefix(kubeconfig, "$HOME"+string(os.PathSeparator)) {
		home, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		kubeconfig = home + kubeconfig[5:]
		err = cmd.Flags().Set("kubeconfig", kubeconfig)
		if err != nil {
			panic(err)
		}
	}

	return nil
}

func Execute() error {
	return Command.Execute()
}

func init() {
	cobra.OnInitialize(initLog, initConfig)

	flags.Kubeconfig(Command)
	flags.Context(Command)
	flags.Namespace(Command)
	flags.Dialect(Command)
	flags.Pod(Command)
	flags.LogLevel(Command)
	flags.LogFormat(Command)
	flags.GitHubActions(Command)
	flags.Database(Command)
	flags.Username(Command)
	flags.Password(Command)
	flags.Redact(Command)

	Command.AddGroup(
		&cobra.Group{
			ID:    "ro",
			Title: "Read Commands",
		},
		&cobra.Group{
			ID:    "rw",
			Title: "Write Commands",
		},
	)

	Command.AddCommand(
		exec.Command,
		dump.Command,
		restore.Command,
		port_forward.Command,
		ui.Command,
	)
}

func initLog() {
	logLevel, err := Command.PersistentFlags().GetString("log-level")
	if err != nil {
		panic(err)
	}
	parsedLevel, err := log.ParseLevel(logLevel)
	if err != nil {
		log.WithField("log-level", logLevel).Warn("invalid log level. defaulting to info.")
		err = Command.PersistentFlags().Set("log-level", "info")
		if err != nil {
			panic(err)
		}
		parsedLevel = log.InfoLevel
	}
	log.SetLevel(parsedLevel)

	logFormat, err := Command.PersistentFlags().GetString("log-format")
	if err != nil {
		panic(err)
	}
	switch logFormat {
	case "text", "txt", "t":
		log.SetFormatter(&log.TextFormatter{})
	case "json", "j":
		log.SetFormatter(&log.JSONFormatter{})
	default:
		log.WithField("log-format", logFormat).Warn("invalid log formatter. defaulting to text.")
		err = Command.PersistentFlags().Set("log-format", "text")
		if err != nil {
			panic(err)
		}
	}
}

func initConfig() {
	viper.SetConfigName("kubedb")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("$HOME/.config/")
	viper.AddConfigPath("$HOME/")
	viper.AddConfigPath("/etc/kubedb/")

	viper.AutomaticEnv()
	viper.SetEnvPrefix("kubedb")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found; ignore error
		} else {
			// Config file was found but another error was produced
			panic(fmt.Errorf("Fatal error reading config file: %w \n", err))
		}
	}
}

func buildVersion() string {
	result := Version
	if Commit != "" {
		result += " (" + Commit + ")"
	}
	return result
}

package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/clevyr/kubedb/cmd/dump"
	"github.com/clevyr/kubedb/cmd/exec"
	"github.com/clevyr/kubedb/cmd/port_forward"
	"github.com/clevyr/kubedb/cmd/restore"
	"github.com/clevyr/kubedb/internal/config/flags"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func NewCommand(version, commit string) *cobra.Command {
	cmd := &cobra.Command{
		Use:               "kubedb",
		Short:             "interact with a database inside of Kubernetes",
		Version:           buildVersion(version, commit),
		DisableAutoGenTag: true,
		Long: `kubedb is a command to interact with a database running in a Kubernetes cluster.

Multiple database types (referred to as the "dialect") are supported.
If the dialect is not configured via flag, it will be detected dynamically.

Supported Database Dialects:
  - PostgreSQL
  - MariaDB
  - MongoDB

If not configured via flag, some configuration variables will be loaded from the target pod's env vars.

Dynamic Env Var Variables:
  - Database
  - Username (fallback value: "postgres" if PostgreSQL, "mariadb" if MariaDB, "root" if MongoDB)
  - Password
`,

		PersistentPreRunE: preRun,
	}

	flags.Kubeconfig(cmd)
	flags.Context(cmd)
	flags.Namespace(cmd)
	flags.Dialect(cmd)
	flags.Pod(cmd)
	flags.LogLevel(cmd)
	flags.LogFormat(cmd)
	flags.GitHubActions(cmd)
	flags.Database(cmd)
	flags.Username(cmd)
	flags.Password(cmd)
	flags.Redact(cmd)
	cmd.InitDefaultVersionFlag()

	cmd.AddGroup(
		&cobra.Group{
			ID:    "ro",
			Title: "Read Commands",
		},
		&cobra.Group{
			ID:    "rw",
			Title: "Write Commands",
		},
	)

	cmd.AddCommand(
		exec.NewCommand(),
		dump.NewCommand(),
		restore.NewCommand(),
		port_forward.NewCommand(),
	)

	return cmd
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

	initLog(cmd)
	if err := initConfig(); err != nil {
		return err
	}
	initLog(cmd)
	return nil
}

func initLog(cmd *cobra.Command) {
	logLevel, err := cmd.Flags().GetString("log-level")
	if err != nil {
		panic(err)
	}
	parsedLevel, err := log.ParseLevel(logLevel)
	if err != nil {
		log.WithField("log-level", logLevel).Warn("invalid log level. defaulting to info.")
		err = cmd.Flags().Set("log-level", "info")
		if err != nil {
			panic(err)
		}
		parsedLevel = log.InfoLevel
	}
	log.SetLevel(parsedLevel)

	logFormat, err := cmd.Flags().GetString("log-format")
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
		err = cmd.Flags().Set("log-format", "text")
		if err != nil {
			panic(err)
		}
	}
}

func initConfig() error {
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
			log.Debug("Could not find config file")
		} else {
			// Config file was found but another error was produced
			return fmt.Errorf("Fatal error reading config file: %w", err)
		}
	}

	log.WithField("path", viper.ConfigFileUsed()).Debug("Loaded config file")
	return nil
}

func buildVersion(version, commit string) string {
	if commit != "" {
		version += " (" + commit + ")"
	}
	return version
}

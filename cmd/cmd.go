package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/clevyr/kubedb/cmd/dump"
	"github.com/clevyr/kubedb/cmd/exec"
	"github.com/clevyr/kubedb/cmd/port_forward"
	"github.com/clevyr/kubedb/cmd/restore"
	"github.com/clevyr/kubedb/internal/config/flags"
	"github.com/clevyr/kubedb/internal/consts"
	"github.com/clevyr/kubedb/internal/util"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "kubedb",
		Short:             "Painlessly work with databases in Kubernetes.",
		Version:           buildVersion(),
		DisableAutoGenTag: true,

		PersistentPreRunE: preRun,
	}

	flags.Kubeconfig(cmd)
	flags.Context(cmd)
	flags.Namespace(cmd)
	flags.Dialect(cmd)
	flags.Pod(cmd)
	flags.LogLevel(cmd)
	flags.LogFormat(cmd)
	flags.Redact(cmd)
	cmd.InitDefaultVersionFlag()

	cmd.AddGroup(
		&cobra.Group{
			ID:    "ro",
			Title: "Read Only Commands:",
		},
		&cobra.Group{
			ID:    "rw",
			Title: "Read/Write Commands:",
		},
	)

	cmd.AddCommand(
		exec.New(),
		dump.New(),
		restore.New(),
		port_forward.New(),
	)

	return cmd
}

func preRun(cmd *cobra.Command, args []string) error {
	flags.BindKubeconfig(cmd)
	flags.BindLogLevel(cmd)
	flags.BindLogFormat(cmd)
	flags.BindRedact(cmd)

	ctx, cancel := signal.NotifyContext(cmd.Context(), os.Interrupt, os.Kill, syscall.SIGTERM)
	cmd.PersistentPostRun = func(cmd *cobra.Command, args []string) { cancel() }
	cmd.SetContext(ctx)

	kubeconfig := viper.GetString(consts.KubeconfigKey)
	if kubeconfig == "$HOME" || strings.HasPrefix(kubeconfig, "$HOME"+string(os.PathSeparator)) {
		home, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		kubeconfig = home + kubeconfig[5:]
		viper.Set(consts.KubeconfigKey, kubeconfig)
	}

	initLog(cmd)
	if err := initConfig(); err != nil {
		return err
	}
	initLog(cmd)
	return nil
}

func initLog(cmd *cobra.Command) {
	logLevel := viper.GetString(consts.LogLevelKey)
	parsedLevel, err := log.ParseLevel(logLevel)
	if err != nil {
		log.WithField("log-level", logLevel).Warn("invalid log level. defaulting to info.")
		viper.Set(consts.LogLevelKey, "info")
		parsedLevel = log.InfoLevel
	}
	log.SetLevel(parsedLevel)

	logFormat := viper.GetString(consts.LogFormatKey)
	switch logFormat {
	case "text", "txt", "t":
		log.SetFormatter(&log.TextFormatter{})
	case "json", "j":
		log.SetFormatter(&log.JSONFormatter{})
	default:
		log.WithField("log-format", logFormat).Warn("invalid log formatter. defaulting to text.")
		viper.Set(consts.LogFormatKey, "text")
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

func buildVersion() string {
	result := util.GetVersion()
	if commit := util.GetCommit(); commit != "" {
		if len(commit) > 8 {
			commit = commit[:8]
		}
		result += " (" + commit + ")"
	}
	return result
}

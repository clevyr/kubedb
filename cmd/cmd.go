package cmd

import (
	"errors"
	"github.com/clevyr/kubedb/cmd/dump"
	"github.com/clevyr/kubedb/cmd/exec"
	"github.com/clevyr/kubedb/cmd/restore"
	"github.com/clevyr/kubedb/internal/config/flags"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	Version = "next"
	Commit  = ""
)

var Command = &cobra.Command{
	Use:               "kubedb",
	Short:             "Interact with a database inside of Kubernetes",
	Version:           buildVersion(),
	PersistentPreRunE: preRun,
}

func preRun(cmd *cobra.Command, args []string) error {
	grammarFlag, err := cmd.Flags().GetString("grammar")
	if err != nil {
		panic(err)
	}

	podFlag, err := cmd.Flags().GetString("pod")
	if err != nil {
		panic(err)
	}

	if podFlag != "" && grammarFlag == "" {
		return errors.New("pod flag is set, but grammar is missing. please add --grammar")
	}

	return nil
}

func Execute() error {
	return Command.Execute()
}

func init() {
	cobra.OnInitialize(initLog)

	flags.Kubeconfig(Command)
	flags.Namespace(Command)
	flags.Grammar(Command)
	flags.Pod(Command)
	flags.LogLevel(Command)
	flags.LogFormat(Command)
	flags.GitHubActions(Command)

	Command.AddCommand(
		exec.Command,
		dump.Command,
		restore.Command,
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

func buildVersion() string {
	result := Version
	if Commit != "" {
		result += " (" + Commit + ")"
	}
	return result
}

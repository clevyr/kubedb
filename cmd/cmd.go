package cmd

import (
	"errors"
	"github.com/clevyr/kubedb/cmd/dump"
	"github.com/clevyr/kubedb/cmd/exec"
	"github.com/clevyr/kubedb/cmd/restore"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"k8s.io/client-go/util/homedir"
	"path/filepath"
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

const DefaultLogLevel = "info"
const DefaultLogFormat = "text"

func init() {
	cobra.OnInitialize(initLog)

	var kubeconfigDefault string
	if home := homedir.HomeDir(); home != "" {
		kubeconfigDefault = filepath.Join(home, ".kube", "config")
	}
	Command.PersistentFlags().String("kubeconfig", kubeconfigDefault, "absolute path to the kubeconfig file")
	Command.PersistentFlags().StringP("namespace", "n", "", "the namespace scope for this CLI request")
	Command.PersistentFlags().String("grammar", "", "database grammar. detected if not set. one of: (postgres|mariadb)")
	Command.PersistentFlags().String("pod", "", "force a specific pod. if this flag is set, grammar is required.")

	Command.PersistentFlags().String("log-level", DefaultLogLevel, "log level. one of: trace|debug|info|warning|error|fatal|panic")
	Command.PersistentFlags().String("log-format", DefaultLogFormat, "log formatter. one of: text|json")

	Command.PersistentFlags().Bool("github-actions", false, "Enables GitHub Actions log output")
	_ = Command.PersistentFlags().MarkHidden("github-actions")

	Command.AddCommand(
		exec.Command,
		dump.Command,
		restore.Command,
	)
}

func initLog() {
	logLevel, _ := Command.PersistentFlags().GetString("log-level")
	parsedLevel, err := log.ParseLevel(logLevel)
	if err != nil {
		log.WithField("log-level", logLevel).Warn("invalid log level. defaulting to info.")
		_ = Command.PersistentFlags().Set("log-level", DefaultLogLevel)
		parsedLevel = log.InfoLevel
	}
	log.SetLevel(parsedLevel)

	logFormat, _ := Command.PersistentFlags().GetString("log-format")
	switch logFormat {
	case "text", "txt", "t":
		log.SetFormatter(&log.TextFormatter{})
	case "json", "j":
		log.SetFormatter(&log.JSONFormatter{})
	default:
		log.WithField("log-format", logFormat).Warn("invalid log formatter. defaulting to text.")
		_ = Command.PersistentFlags().Set("log-format", DefaultLogFormat)
	}
}

func buildVersion() string {
	result := Version
	if Commit != "" {
		result += " (" + Commit + ")"
	}
	return result
}

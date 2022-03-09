package cmd

import (
	"github.com/clevyr/kubedb/cmd/dump"
	"github.com/clevyr/kubedb/cmd/exec"
	"github.com/clevyr/kubedb/cmd/restore"
	"github.com/spf13/cobra"
	"k8s.io/client-go/util/homedir"
	"path/filepath"
)

var (
	Version = "next"
	Commit  = ""
)

var Command = &cobra.Command{
	Use:     "kubedb",
	Short:   "Interact with a database inside of Kubernetes",
	Version: buildVersion(),
}

func Execute() error {
	return Command.Execute()
}

func init() {
	var kubeconfigDefault string
	if home := homedir.HomeDir(); home != "" {
		kubeconfigDefault = filepath.Join(home, ".kube", "config")
	}
	Command.PersistentFlags().String("kubeconfig", kubeconfigDefault, "absolute path to the kubeconfig file")
	Command.PersistentFlags().StringP("namespace", "n", "", "the namespace scope for this CLI request")
	Command.PersistentFlags().String("type", "", "database type. one of: (postgres|mariadb)")

	Command.PersistentFlags().Bool("github-actions", false, "Enables GitHub Actions log output")
	_ = Command.PersistentFlags().MarkHidden("github-actions")

	Command.AddCommand(
		exec.Command,
		dump.Command,
		restore.Command,
	)
}

func buildVersion() string {
	result := Version
	if Commit != "" {
		result += " (" + Commit + ")"
	}
	return result
}

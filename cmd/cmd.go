package cmd

import (
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/clevyr/kubedb/cmd/dump"
	"github.com/clevyr/kubedb/cmd/exec"
	"github.com/clevyr/kubedb/cmd/portforward"
	"github.com/clevyr/kubedb/cmd/restore"
	"github.com/clevyr/kubedb/cmd/status"
	"github.com/clevyr/kubedb/internal/config"
	"github.com/clevyr/kubedb/internal/config/flags"
	"github.com/clevyr/kubedb/internal/consts"
	"github.com/clevyr/kubedb/internal/log"
	"github.com/clevyr/kubedb/internal/notifier"
	"github.com/clevyr/kubedb/internal/util"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func NewCommand() *cobra.Command {
	name := "kubedb"
	var annotations map[string]string
	if filepath.Base(os.Args[0]) == "kubectl-db" {
		// Installed as a kubectl plugin
		name = "kubectl-db"
		annotations = map[string]string{
			cobra.CommandDisplayNameAnnotation: "kubectl db",
		}
	}

	cmd := &cobra.Command{
		Use:               name,
		Short:             "Painlessly work with databases in Kubernetes.",
		Long:              newDescription(),
		Version:           buildVersion(),
		DisableAutoGenTag: true,
		Annotations:       annotations,

		PersistentPreRunE: preRun,
	}

	flags.Kubeconfig(cmd)
	flags.Context(cmd)
	flags.Namespace(cmd)
	flags.Dialect(cmd)
	flags.Pod(cmd)
	flags.LogLevel(cmd)
	flags.LogFormat(cmd)
	flags.Healthchecks(cmd)
	flags.Mask(cmd)
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
		portforward.New(),
		status.New(),
	)

	return cmd
}

func preRun(cmd *cobra.Command, _ []string) error {
	flags.BindKubeconfig(cmd)
	flags.BindLogLevel(cmd)
	flags.BindLogFormat(cmd)
	flags.BindMask(cmd)
	flags.BindHealthchecks(cmd)

	ctx, cancel := signal.NotifyContext(cmd.Context(), os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)
	defer func() {
		util.OnFinalize(func(_ error) {
			cancel()
		})
	}()
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

	if err := config.LoadViper(); err != nil {
		return err
	}
	log.InitFromCmd(cmd)
	cmd.Root().SilenceErrors = true

	if url := viper.GetString(consts.HealthchecksPingURLKey); url != "" {
		if handler, err := notifier.NewHealthchecks(url); err != nil {
			slog.Error("Notifications creation failed", "error", err)
		} else {
			if err := handler.Started(ctx); err != nil {
				slog.Error("Notifications ping start failed", "error", err)
			}

			util.OnFinalize(func(err error) {
				if err := handler.Finished(ctx, err); err != nil {
					slog.Error("Notifications ping finished failed", "error", err)
				}
			})

			cmd.SetContext(notifier.NewContext(cmd.Context(), handler))
		}
	}

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

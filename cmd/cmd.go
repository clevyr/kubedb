package cmd

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"runtime/debug"
	"syscall"

	"github.com/clevyr/kubedb/cmd/dump"
	"github.com/clevyr/kubedb/cmd/exec"
	"github.com/clevyr/kubedb/cmd/portforward"
	"github.com/clevyr/kubedb/cmd/restore"
	"github.com/clevyr/kubedb/cmd/status"
	"github.com/clevyr/kubedb/internal/config"
	"github.com/clevyr/kubedb/internal/config/flags"
	"github.com/clevyr/kubedb/internal/finalizer"
	"github.com/clevyr/kubedb/internal/log"
	"github.com/clevyr/kubedb/internal/notifier"
	"github.com/clevyr/kubedb/internal/util"
	"github.com/spf13/cobra"
)

func New() *cobra.Command {
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
		SilenceErrors:     true,

		PersistentPreRunE: preRun,
	}

	flags.Config(cmd)
	flags.Kubeconfig(cmd)
	flags.Context(cmd)
	flags.Namespace(cmd)
	flags.Dialect(cmd)
	flags.Pod(cmd)
	flags.Log(cmd)
	flags.Healthchecks(cmd)
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
	http.DefaultTransport = util.NewUserAgentTransport()

	if cmd.Name() == "__complete" {
		config.IsCompletion = true
		return nil
	}

	if err := config.Load(cmd); err != nil {
		return err
	}

	ctx, cancel := signal.NotifyContext(cmd.Context(), os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)
	defer finalizer.Add(func(_ error) { cancel() })
	cmd.SetContext(ctx)

	log.InitGlobal(cmd)

	if config.Global.HealthchecksPingURL != "" {
		if handler, err := notifier.NewHealthchecks(config.Global.HealthchecksPingURL); err != nil {
			slog.Error("Notifications creation failed", "error", err)
		} else {
			if err := handler.Started(ctx); err != nil {
				slog.Error("Notifications ping start failed", "error", err)
			}

			finalizer.Add(func(err error) {
				if err := handler.Finished(ctx, err); err != nil {
					slog.Error("Notifications ping finished failed", "error", err)
				}
			})

			cmd.SetContext(notifier.NewContext(cmd.Context(), handler))
		}
	}

	return nil
}

var errPanic = errors.New("panic")

func buildVersion() string {
	result := util.GetVersion()
	if commit := util.GetCommit(); commit != "" {
		result += " (" + commit + ")"
	}
	return result
}

func Execute(cmd *cobra.Command) (err error) {
	if cmd == nil {
		cmd = New()
	}

	defer finalizer.PostRun(err)
	defer func() {
		if msg := recover(); msg != nil {
			slog.Error("Recovered from panic", "error", msg)
			err = fmt.Errorf("%w: %v\n\n%s", errPanic, msg, string(debug.Stack()))
		}
	}()

	err = cmd.Execute()
	return err
}

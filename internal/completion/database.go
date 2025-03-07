package completion

import (
	"log/slog"
	"os"
	"strings"

	"github.com/clevyr/kubedb/internal/config"
	"github.com/clevyr/kubedb/internal/config/conftypes"
	"github.com/clevyr/kubedb/internal/kubernetes"
	"github.com/clevyr/kubedb/internal/util"
	"github.com/spf13/cobra"
)

func LoadConfig(cmd *cobra.Command) error {
	if err := config.Load(cmd); err != nil {
		slog.Error("Failed to load configuration", "error", err)
		return err
	}

	config.Global.CreateJob = false
	config.Global.CreateNetworkPolicy = false
	return nil
}

func TablesList(cmd *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
	if err := LoadConfig(cmd); err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	if err := util.DefaultSetup(cmd, config.Global, util.SetupOptions{NoSurvey: true}); err != nil {
		slog.Error("Setup failed", "error", err)
		return nil, cobra.ShellCompDirectiveError
	}

	conf := &conftypes.Exec{Global: config.Global, DisableHeaders: true}

	db, ok := conf.Dialect.(conftypes.DBTableLister)
	if !ok {
		slog.Error("Dialect does not support listing tables", "name", conf.Dialect.Name())
		return nil, cobra.ShellCompDirectiveError
	}

	conf.Command = db.TableListQuery()
	return DatabaseQuery(cmd, conf)
}

func DatabasesList(cmd *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
	if err := LoadConfig(cmd); err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	if err := util.DefaultSetup(cmd, config.Global, util.SetupOptions{NoSurvey: true}); err != nil {
		slog.Error("Setup failed", "error", err)
		return nil, cobra.ShellCompDirectiveError
	}

	conf := &conftypes.Exec{Global: config.Global, DisableHeaders: true}

	db, ok := conf.Dialect.(conftypes.DBDatabaseLister)
	if !ok {
		slog.Error("Dialect does not support listing databases", "name", conf.Dialect.Name())
		return nil, cobra.ShellCompDirectiveError
	}

	conf.Command = db.DatabaseListQuery()
	return DatabaseQuery(cmd, conf)
}

func DatabaseQuery(cmd *cobra.Command, conf *conftypes.Exec) ([]string, cobra.ShellCompDirective) {
	db, ok := conf.Dialect.(conftypes.DBExecer)
	if !ok {
		slog.Error("Dialect does not support exec", "name", conf.Dialect.Name())
		return nil, cobra.ShellCompDirectiveError
	}

	var buf strings.Builder
	if err := conf.Client.Exec(cmd.Context(), kubernetes.ExecOptions{
		Pod:    conf.DBPod,
		Cmd:    db.ExecCommand(conf).String(),
		Stdout: &buf,
		Stderr: os.Stderr,
	}); err != nil {
		slog.Error("Exec failed", "error", err)
		return nil, cobra.ShellCompDirectiveError
	}

	names := strings.Split(buf.String(), "\n")
	return names, cobra.ShellCompDirectiveNoFileComp
}

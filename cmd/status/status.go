//nolint:forbidigo
package status

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/clevyr/kubedb/internal/config"
	"github.com/clevyr/kubedb/internal/config/flags"
	"github.com/clevyr/kubedb/internal/consts"
	"github.com/clevyr/kubedb/internal/database"
	"github.com/clevyr/kubedb/internal/kubernetes"
	"github.com/clevyr/kubedb/internal/log"
	"github.com/clevyr/kubedb/internal/tui"
	"github.com/clevyr/kubedb/internal/util"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//nolint:gochecknoglobals
var conf config.Global

func New() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "status",
		Short:   "View connection status",
		GroupID: "ro",
		PreRunE: preRun,
		RunE:    run,

		Args:              cobra.NoArgs,
		ValidArgsFunction: cobra.NoFileCompletions,
	}
	flags.JobPodLabels(cmd)
	flags.Port(cmd)
	flags.Database(cmd)
	flags.Username(cmd)
	flags.Password(cmd)

	return cmd
}

func preRun(cmd *cobra.Command, _ []string) error {
	cmd.SilenceUsage = true
	viper.Set(consts.LogLevelFlag, slog.LevelWarn.String())
	log.InitFromCmd(cmd)
	flags.BindJobPodLabels(cmd)
	return nil
}

func run(cmd *cobra.Command, _ []string) error {
	statusStyle := lipgloss.NewStyle().PaddingLeft(1)
	prefixOk := statusStyle.Foreground(tui.ColorGreen).Render("✓")
	prefixNeutral := statusStyle.Foreground(tui.ColorHiBlack).Render("-")
	prefixErr := statusStyle.Foreground(tui.ColorRed).Render("✗")
	bold := lipgloss.NewStyle().Bold(true).Render

	defaultSetupErr := util.DefaultSetup(cmd, &conf, util.SetupOptions{Name: "status"})

	fmt.Println(bold("Cluster Info"))
	if conf.Client.ClientSet != nil {
		fmt.Println(
			prefixOk,
			"Using cluster",
			bold(conf.Context),
			"at",
			bold(conf.Client.ClientConfig.Host),
		)
	} else {
		fmt.Println(prefixErr, "Failed to load cluster config:", defaultSetupErr.Error())
		os.Exit(1)
	}

	if serverVersion, err := conf.Client.Discovery.ServerVersion(); err == nil {
		fmt.Println(prefixOk, "Cluster version is", bold(serverVersion.String()))
	} else {
		fmt.Println(prefixErr, "Cluster version check failed:", err.Error())
	}

	if namespaces, err := conf.Client.Namespaces().List(cmd.Context(), metav1.ListOptions{}); err == nil {
		fmt.Println(prefixOk, "Cluster has", bold(strconv.Itoa(len(namespaces.Items))), "namespaces")
	} else {
		fmt.Println(prefixErr, "Failed to list namespaces:", err.Error())
	}

	fmt.Println(prefixOk, "Using namespace", bold(conf.Client.Namespace))

	fmt.Println(bold("Database Info"))
	if defaultSetupErr == nil {
		fmt.Println(
			prefixOk,
			"Found",
			bold(conf.Dialect.Name()),
			"database",
			bold(conf.DBPod.ObjectMeta.Name),
		)
	} else {
		if errors.Is(defaultSetupErr, database.ErrDatabaseNotFound) {
			fmt.Println(prefixErr, "Could not detect a database")
		} else {
			fmt.Println(prefixErr, "Failed to search for database:", defaultSetupErr.Error())
		}
		os.Exit(1)
	}

	if err := util.CreateJob(cmd.Context(), &conf, util.SetupOptions{Name: "status"}); err == nil {
		fmt.Println(prefixOk, "Jobs can be created")
	} else {
		fmt.Println(prefixErr, "Job creation failed:", err.Error())
		os.Exit(1)
	}

	var buf strings.Builder
	if db, ok := conf.Dialect.(dbExecList); ok {
		listTablesCmd := db.ExecCommand(config.Exec{
			Global:         conf,
			DisableHeaders: true,
			Command:        db.TableListQuery(),
		})
		execOpts := kubernetes.ExecOptions{
			Pod:    conf.JobPod,
			Cmd:    listTablesCmd.String(),
			Stdout: &buf,
			Stderr: os.Stderr,
		}
		if err := conf.Client.Exec(cmd.Context(), execOpts); err == nil {
			names := strings.Split(buf.String(), "\n")
			fmt.Println(prefixOk, "Database has", bold(strconv.Itoa(len(names))), "tables")
		} else {
			fmt.Println(prefixErr, "Failed to connect to database", err.Error())
			os.Exit(1)
		}
	} else {
		fmt.Println(prefixNeutral, "Database does not support listing tables")
	}

	return nil
}

type dbExecList interface {
	config.DBExecer
	config.DBTableLister
}

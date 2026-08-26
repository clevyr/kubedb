package status

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"charm.land/lipgloss/v2"
	"gabe565.com/utils/must"
	"gabe565.com/utils/slogx"
	"github.com/clevyr/kubedb/internal/config"
	"github.com/clevyr/kubedb/internal/config/conftypes"
	"github.com/clevyr/kubedb/internal/config/flags"
	"github.com/clevyr/kubedb/internal/consts"
	"github.com/clevyr/kubedb/internal/database"
	"github.com/clevyr/kubedb/internal/kubernetes"
	"github.com/clevyr/kubedb/internal/log"
	"github.com/clevyr/kubedb/internal/tui"
	"github.com/clevyr/kubedb/internal/util"
	"github.com/spf13/cobra"
	authorizationv1 "k8s.io/api/authorization/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

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
	must.Must(config.K.Set(consts.FlagLogLevel, slogx.LevelWarn.String()))
	log.InitGlobal(cmd)
	return nil
}

func run(cmd *cobra.Command, _ []string) error {
	conf := config.Global

	out := tui.NewWriter(cmd.OutOrStdout())

	statusStyle := lipgloss.NewStyle().PaddingLeft(1)
	prefixOk := statusStyle.Foreground(tui.ColorGreen).Render("✓")
	prefixNeutral := statusStyle.Foreground(tui.ColorHiBlack).Render("-")
	prefixErr := statusStyle.Foreground(tui.ColorRed).Render("✗")
	bold := lipgloss.NewStyle().Bold(true).Render

	defaultSetupErr := util.DefaultSetup(cmd, conf)

	_, _ = fmt.Fprintln(out, bold("Cluster Info"))
	if conf.Client.ClientSet != nil {
		_, _ = fmt.Fprintln(out,
			prefixOk,
			"Using cluster",
			bold(conf.Context),
			"at",
			bold(conf.Client.ClientConfig.Host),
		)
	} else {
		if defaultSetupErr != nil {
			_, _ = fmt.Fprintln(out, prefixErr, "Failed to load cluster config:", defaultSetupErr.Error())
		}
		os.Exit(1)
	}

	if serverVersion, err := conf.Client.Discovery.ServerVersion(); err == nil {
		_, _ = fmt.Fprintln(out, prefixOk, "Cluster version is", bold(serverVersion.String()))
	} else {
		_, _ = fmt.Fprintln(out, prefixErr, "Cluster version check failed:", err.Error())
	}

	if namespaces, err := conf.Client.Namespaces().List(cmd.Context(), metav1.ListOptions{}); err == nil {
		_, _ = fmt.Fprintln(out,
			prefixOk, "Cluster has", bold(strconv.Itoa(len(namespaces.Items))), "namespaces",
		)
	} else {
		_, _ = fmt.Fprintln(out, prefixErr, "Failed to list namespaces:", err.Error())
	}

	_, _ = fmt.Fprintln(out, prefixOk, "Using namespace", bold(conf.Client.Namespace))

	_, _ = fmt.Fprintln(out, bold("Database Info"))
	if defaultSetupErr == nil {
		_, _ = fmt.Fprintln(out,
			prefixOk, "Found", bold(conf.Dialect.Name()), "database", bold(conf.DBPod.Name),
		)
	} else {
		if errors.Is(defaultSetupErr, database.ErrDatabaseNotFound) {
			_, _ = fmt.Fprintln(out, prefixErr, "Could not detect a database")
		} else {
			_, _ = fmt.Fprintln(out, prefixErr, "Failed to search for database:", defaultSetupErr.Error())
		}
		os.Exit(1)
	}

	if res, err := conf.Client.ClientSet.AuthorizationV1().
		SelfSubjectAccessReviews().
		Create(cmd.Context(), &authorizationv1.SelfSubjectAccessReview{
			Spec: authorizationv1.SelfSubjectAccessReviewSpec{
				ResourceAttributes: &authorizationv1.ResourceAttributes{
					Namespace: conf.Client.Namespace,
					Verb:      "create",
					Group:     "batch",
					Resource:  "jobs",
				},
			},
		}, metav1.CreateOptions{}); err != nil {
		_, _ = fmt.Fprintln(out, prefixErr, "Job permission check failed:", err.Error())
	} else if res.Status.Allowed {
		_, _ = fmt.Fprintln(out, prefixOk, "Jobs can be created")
	} else {
		_, _ = fmt.Fprintln(out, prefixErr, "Missing permission to create jobs")
	}

	var buf strings.Builder
	if db, ok := conf.Dialect.(dbExecList); ok {
		listTablesCmd := db.ExecCommand(&conftypes.Exec{
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
			var count int
			output := strings.TrimSpace(buf.String())
			if len(output) != 0 {
				count = strings.Count(output, "\n") + 1
			}
			_, _ = fmt.Fprintln(out, prefixOk, "Database has", bold(strconv.Itoa(count)), "tables")
		} else {
			_, _ = fmt.Fprintln(out, prefixErr, "Failed to connect to database", err.Error())
			os.Exit(1)
		}
	} else {
		_, _ = fmt.Fprintln(out, prefixNeutral, "Database does not support listing tables")
	}

	return nil
}

type dbExecList interface {
	conftypes.DBExecer
	conftypes.DBTableLister
}

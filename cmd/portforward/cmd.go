package portforward

import (
	"fmt"
	"log/slog"
	"strconv"

	"gabe565.com/utils/must"
	"github.com/clevyr/kubedb/internal/actions/portforward"
	"github.com/clevyr/kubedb/internal/config"
	"github.com/clevyr/kubedb/internal/config/conftypes"
	"github.com/clevyr/kubedb/internal/config/flags"
	"github.com/clevyr/kubedb/internal/consts"
	"github.com/clevyr/kubedb/internal/util"
	"github.com/spf13/cobra"
)

//nolint:gochecknoglobals
var action = &portforward.PortForward{PortForward: conftypes.PortForward{Global: config.Global}}

func New() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "port-forward [local_port]",
		Short:   "Set up a local port forward",
		Long:    newDescription(),
		Args:    cobra.MaximumNArgs(1),
		GroupID: "rw",
		RunE:    run,
		PreRunE: preRun,

		ValidArgsFunction: localPortCompletion,
	}

	flags.Port(cmd)

	cmd.Flags().StringSlice(consts.FlagAddress, []string{"127.0.0.1", "::1"}, "Local listen address")
	must.Must(cmd.RegisterFlagCompletionFunc(consts.FlagAddress,
		func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
			return []string{
				"127.0.0.1\tprivate",
				"::1\tprivate",
				"0.0.0.0\tpublic",
				"::\tpublic",
			}, cobra.ShellCompDirectiveNoFileComp
		}),
	)

	cmd.Flags().Uint16(consts.FlagListenPort, 0, "Local listen port (default discovered)")
	must.Must(cmd.RegisterFlagCompletionFunc(consts.FlagListenPort, localPortCompletion))

	return cmd
}

func localPortCompletion(cmd *cobra.Command, args []string, _ string) ([]string, cobra.ShellCompDirective) {
	if len(args) != 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	config.Global.SkipSurvey = true
	if err := preRun(cmd, args); err != nil {
		slog.Error("Pre-run failed", "error", err)
		return nil, cobra.ShellCompDirectiveError
	}

	db, ok := action.Dialect.(conftypes.DBHasPort)
	if !ok {
		slog.Error("Dialect does not support port-forwarding", "name", action.Dialect.Name())
		return nil, cobra.ShellCompDirectiveError
	}

	defaultPort := db.PortDefault()
	return []string{
		strconv.Itoa(int(action.ListenPort)),
		strconv.Itoa(int(defaultPort)),
		strconv.Itoa(int(defaultPort + 1)),
	}, cobra.ShellCompDirectiveNoFileComp
}

func preRun(cmd *cobra.Command, args []string) error {
	if err := config.Unmarshal(cmd, "port-forward", &action); err != nil {
		return err
	}

	if len(args) != 0 {
		port, err := strconv.ParseUint(args[0], 10, 16)
		if err != nil {
			return err
		}

		action.ListenPort = uint16(port)
	}

	err := util.DefaultSetup(cmd, action.Global)
	if err != nil {
		return err
	}

	if action.ListenPort == 0 {
		db, ok := action.Dialect.(conftypes.DBHasPort)
		if !ok {
			return fmt.Errorf("%w: %s", util.ErrNoPortForward, action.Dialect.Name())
		}

		action.ListenPort = 30000 + db.PortDefault()
	}

	return nil
}

func run(cmd *cobra.Command, _ []string) error {
	return action.Run(cmd.Context())
}

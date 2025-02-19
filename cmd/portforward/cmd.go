package portforward

import (
	"fmt"
	"log/slog"
	"strconv"

	"gabe565.com/utils/must"
	"github.com/clevyr/kubedb/internal/actions/portforward"
	"github.com/clevyr/kubedb/internal/config"
	"github.com/clevyr/kubedb/internal/config/flags"
	"github.com/clevyr/kubedb/internal/consts"
	"github.com/clevyr/kubedb/internal/util"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

//nolint:gochecknoglobals
var (
	action       portforward.PortForward
	setupOptions = util.SetupOptions{DisableAuthFlags: true}
)

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
			return []string{"127.0.0.1\tprivate", "::1\tprivate", "0.0.0.0\tpublic", "::\tpublic"}, cobra.ShellCompDirectiveNoFileComp
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

	setupOptions.NoSurvey = true
	err := preRun(cmd, args)
	if err != nil {
		slog.Error("Pre-run failed", "error", err)
		return nil, cobra.ShellCompDirectiveError
	}

	db, ok := action.Dialect.(config.DBHasPort)
	if !ok {
		slog.Error("Dialect does not support port-forwarding", "name", action.Dialect.Name())
		return nil, cobra.ShellCompDirectiveError
	}

	defaultPort := db.PortDefault()
	return []string{
		strconv.Itoa(int(action.LocalPort)),
		strconv.Itoa(int(defaultPort)),
		strconv.Itoa(int(defaultPort + 1)),
	}, cobra.ShellCompDirectiveNoFileComp
}

func preRun(cmd *cobra.Command, args []string) error {
	must.Must(viper.BindPFlag(consts.KeyPortForwardAddress, cmd.Flags().Lookup(consts.FlagAddress)))
	action.Addresses = viper.GetStringSlice(consts.KeyPortForwardAddress)

	err := util.DefaultSetup(cmd, &action.Global, setupOptions)
	if err != nil {
		return err
	}

	if len(args) != 0 {
		if err := cmd.Flags().Set(consts.FlagListenPort, args[0]); err != nil {
			return err
		}
	}

	action.LocalPort = must.Must2(cmd.Flags().GetUint16(consts.FlagListenPort))
	if action.LocalPort == 0 {
		db, ok := action.Dialect.(config.DBHasPort)
		if !ok {
			return fmt.Errorf("%w: %s", util.ErrNoPortForward, action.Dialect.Name())
		}

		action.LocalPort = 30000 + db.PortDefault()
	}

	return nil
}

func run(cmd *cobra.Command, _ []string) error {
	return action.Run(cmd.Context())
}

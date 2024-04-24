package portforward

import (
	"fmt"
	"strconv"

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

	cmd.Flags().StringSlice(consts.AddrFlag, []string{"127.0.0.1", "::1"}, "Local listen address")
	if err := cmd.RegisterFlagCompletionFunc(
		consts.AddrFlag,
		func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
			return []string{"127.0.0.1\tprivate", "::1\tprivate", "0.0.0.0\tpublic", "::\tpublic"}, cobra.ShellCompDirectiveNoFileComp
		}); err != nil {
		panic(err)
	}

	cmd.Flags().Uint16(consts.ListenPortFlag, 0, "Local listen port (default discovered)")
	if err := cmd.RegisterFlagCompletionFunc(consts.ListenPortFlag, localPortCompletion); err != nil {
		panic(err)
	}

	return cmd
}

func localPortCompletion(cmd *cobra.Command, args []string, _ string) ([]string, cobra.ShellCompDirective) {
	setupOptions.NoSurvey = true
	err := preRun(cmd, args)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	db, ok := action.Dialect.(config.DatabasePort)
	if !ok {
		return nil, cobra.ShellCompDirectiveError
	}

	defaultPort := db.DefaultPort()
	return []string{
		strconv.Itoa(int(action.LocalPort)),
		strconv.Itoa(int(defaultPort)),
		strconv.Itoa(int(defaultPort + 1)),
	}, cobra.ShellCompDirectiveNoFileComp
}

func preRun(cmd *cobra.Command, args []string) error {
	if err := viper.BindPFlag(consts.PortForwardAddrKey, cmd.Flags().Lookup(consts.AddrFlag)); err != nil {
		panic(err)
	}
	action.Addresses = viper.GetStringSlice(consts.PortForwardAddrKey)

	err := util.DefaultSetup(cmd, &action.Global, setupOptions)
	if err != nil {
		return err
	}

	if len(args) != 0 {
		if err := cmd.Flags().Set(consts.ListenPortFlag, args[0]); err != nil {
			return err
		}
	}

	action.LocalPort, err = cmd.Flags().GetUint16(consts.ListenPortFlag)
	if err != nil {
		panic(err)
	}

	if action.LocalPort == 0 {
		db, ok := action.Dialect.(config.DatabasePort)
		if !ok {
			return fmt.Errorf("%w: %s", util.ErrNoPortForward, action.Dialect.Name())
		}

		action.LocalPort = 30000 + db.DefaultPort()
	}

	return nil
}

func run(cmd *cobra.Command, _ []string) error {
	return action.Run(cmd.Context())
}

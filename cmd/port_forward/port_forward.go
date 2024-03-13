package port_forward

import (
	"fmt"
	"strconv"

	"github.com/clevyr/kubedb/internal/actions/port_forward"
	"github.com/clevyr/kubedb/internal/config"
	"github.com/clevyr/kubedb/internal/config/flags"
	"github.com/clevyr/kubedb/internal/consts"
	"github.com/clevyr/kubedb/internal/util"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

var action port_forward.PortForward

func New() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "port-forward [local_port]",
		Short:   "Set up a local port forward",
		Long:    newDescription(),
		Args:    cobra.MaximumNArgs(1),
		GroupID: "rw",
		RunE:    run,
		PreRunE: preRun,
	}

	flags.Port(cmd)

	cmd.Flags().StringSlice(consts.AddrFlag, []string{"127.0.0.1", "::1"}, "Local listen address")
	if err := cmd.RegisterFlagCompletionFunc(
		consts.AddrFlag,
		func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return []string{"127.0.0.1\tprivate", "::1\tprivate", "0.0.0.0\tpublic", "::\tpublic"}, cobra.ShellCompDirectiveNoFileComp
		}); err != nil {
		panic(err)
	}

	cmd.Flags().Uint16(consts.ListenPortFlag, 0, "Local listen port (default discovered)")
	if err := cmd.RegisterFlagCompletionFunc(
		consts.ListenPortFlag,
		func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
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
		}); err != nil {
		panic(err)
	}

	return cmd
}

func preRun(cmd *cobra.Command, args []string) error {
	if err := viper.BindPFlag(consts.PortForwardAddrKey, cmd.Flags().Lookup(consts.AddrFlag)); err != nil {
		panic(err)
	}
	action.Addresses = viper.GetStringSlice(consts.PortForwardAddrKey)

	err := util.DefaultSetup(cmd, &action.Global, util.SetupOptions{DisableAuthFlags: true})
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

func run(cmd *cobra.Command, args []string) (err error) {
	return action.Run(cmd.Context())
}

package port_forward

import (
	"strconv"

	"github.com/clevyr/kubedb/internal/actions/port_forward"
	"github.com/clevyr/kubedb/internal/config/flags"
	"github.com/clevyr/kubedb/internal/util"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

var action port_forward.PortForward

func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "port-forward [local_port]",
		Short:             "Set up a local port forward",
		Args:              cobra.MaximumNArgs(1),
		ValidArgsFunction: validArgs,
		GroupID:           "rw",
		RunE:              run,
		PreRunE:           preRun,
	}

	flags.Address(cmd)

	return cmd
}

func validArgs(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	err := preRun(cmd, args)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	defaultPort := action.Dialect.DefaultPort()

	return []string{
		strconv.Itoa(int(action.LocalPort)),
		strconv.Itoa(int(defaultPort)),
		strconv.Itoa(int(defaultPort + 1)),
	}, cobra.ShellCompDirectiveNoFileComp
}

func preRun(cmd *cobra.Command, args []string) error {
	flags.BindAddress(cmd)
	action.Addresses = viper.GetStringSlice("port-forward.address")

	err := util.DefaultSetup(cmd, &action.Global, util.SetupOptions{DisableJob: true})
	if err != nil {
		return err
	}

	if len(args) == 0 {
		action.LocalPort = 30000 + action.Dialect.DefaultPort()
	} else {
		port, err := strconv.ParseUint(args[0], 10, 16)
		if err != nil {
			util.Teardown(cmd, &action.Global)
			return err
		}
		action.LocalPort = uint16(port)
	}
	return nil
}

func run(cmd *cobra.Command, args []string) (err error) {
	defer func() {
		util.Teardown(cmd, &action.Global)
	}()
	return action.Run(cmd.Context())
}

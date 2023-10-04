package exec

import (
	"github.com/clevyr/kubedb/internal/actions/exec"
	"github.com/clevyr/kubedb/internal/util"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

var action exec.Exec

func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "exec",
		Aliases: []string{"e", "shell"},
		Short:   "Connect to an interactive shell",
		GroupID: "rw",
		RunE:    run,
		PreRunE: preRun,
	}

	return cmd
}

func preRun(cmd *cobra.Command, args []string) error {
	if err := viper.Unmarshal(&action); err != nil {
		return err
	}

	if err := util.DefaultSetup(cmd, &action.Global, util.SetupOptions{Name: "exec"}); err != nil {
		return err
	}

	return nil
}

func run(cmd *cobra.Command, args []string) (err error) {
	defer func() {
		util.Teardown(cmd, &action.Global)
	}()
	return action.Run(cmd.Context())
}

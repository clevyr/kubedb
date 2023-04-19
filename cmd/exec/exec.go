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
		Short:   "connect to an interactive shell",
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

	return util.DefaultSetup(cmd, &action.Global)
}

func run(cmd *cobra.Command, args []string) (err error) {
	return action.Run(cmd.Context())
}

package exec

import (
	"github.com/clevyr/kubedb/internal/actions/exec"
	"github.com/clevyr/kubedb/internal/util"
	"github.com/spf13/cobra"
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

	cmd.Flags().StringVarP(&action.Command, "command", "c", "", "Run a single command and exit")
	return cmd
}

func preRun(cmd *cobra.Command, args []string) error {
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

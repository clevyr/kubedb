package exec

import (
	"github.com/clevyr/kubedb/internal/actions/exec"
	"github.com/clevyr/kubedb/internal/config/flags"
	"github.com/clevyr/kubedb/internal/consts"
	"github.com/clevyr/kubedb/internal/util"
	"github.com/spf13/cobra"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

var action exec.Exec

func New() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "exec",
		Aliases: []string{"e", "shell"},
		Short:   "Connect to an interactive shell",
		GroupID: "rw",
		RunE:    run,
		PreRunE: preRun,
	}

	flags.JobPodLabels(cmd)
	flags.NoJob(cmd)
	flags.Port(cmd)
	flags.Database(cmd)
	flags.Username(cmd)
	flags.Password(cmd)
	cmd.Flags().StringVarP(&action.Command, consts.CommandFlag, "c", "", "Run a single command and exit")

	return cmd
}

func preRun(cmd *cobra.Command, args []string) error {
	flags.BindJobPodLabels(cmd)
	flags.BindNoJob(cmd)

	setupOptions := util.SetupOptions{Name: "exec"}
	if err := util.DefaultSetup(cmd, &action.Global, setupOptions); err != nil {
		return err
	}
	if err := util.CreateJob(cmd, &action.Global, setupOptions); err != nil {
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

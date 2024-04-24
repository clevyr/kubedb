package exec

import (
	"github.com/clevyr/kubedb/internal/actions/exec"
	"github.com/clevyr/kubedb/internal/config/flags"
	"github.com/clevyr/kubedb/internal/consts"
	"github.com/clevyr/kubedb/internal/util"
	"github.com/spf13/cobra"
)

//nolint:gochecknoglobals
var (
	action       exec.Exec
	setupOptions = util.SetupOptions{Name: "exec"}
)

func New() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "exec",
		Aliases: []string{"e", "shell"},
		Short:   "Connect to an interactive shell",
		Long:    newDescription(),
		GroupID: "rw",
		Args:    cobra.NoArgs,
		RunE:    run,
		PreRunE: preRun,
	}

	flags.JobPodLabels(cmd)
	flags.CreateJob(cmd)
	flags.CreateNetworkPolicy(cmd)
	flags.Port(cmd)
	flags.Database(cmd)
	flags.Username(cmd)
	flags.Password(cmd)
	flags.Opts(cmd)
	cmd.Flags().StringVarP(&action.Command, consts.CommandFlag, "c", "", "Run a single command and exit")
	if err := cmd.RegisterFlagCompletionFunc(consts.CommandFlag, cobra.NoFileCompletions); err != nil {
		panic(err)
	}

	return cmd
}

func preRun(cmd *cobra.Command, _ []string) error {
	flags.BindJobPodLabels(cmd)
	flags.BindCreateJob(cmd)
	flags.BindCreateNetworkPolicy(cmd)
	flags.BindOpts(cmd)

	if err := util.DefaultSetup(cmd, &action.Global, setupOptions); err != nil {
		return err
	}
	if err := util.CreateJob(cmd.Context(), &action.Global, setupOptions); err != nil {
		return err
	}

	return nil
}

func run(cmd *cobra.Command, _ []string) error {
	return action.Run(cmd.Context())
}

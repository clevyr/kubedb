package exec

import (
	"gabe565.com/utils/must"
	"github.com/clevyr/kubedb/internal/actions/exec"
	"github.com/clevyr/kubedb/internal/config"
	"github.com/clevyr/kubedb/internal/config/conftypes"
	"github.com/clevyr/kubedb/internal/config/flags"
	"github.com/clevyr/kubedb/internal/consts"
	"github.com/clevyr/kubedb/internal/util"
	"github.com/spf13/cobra"
)

//nolint:gochecknoglobals
var action = &exec.Exec{Exec: conftypes.Exec{Global: config.Global}}

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

		ValidArgsFunction: cobra.NoFileCompletions,
	}

	flags.JobPodLabels(cmd)
	flags.CreateJob(cmd)
	flags.CreateNetworkPolicy(cmd)
	flags.Port(cmd)
	flags.Database(cmd)
	flags.Username(cmd)
	flags.Password(cmd)
	flags.Opts(cmd)
	cmd.Flags().StringP(consts.FlagCommand, "c", "", "Run a single command and exit")
	must.Must(cmd.RegisterFlagCompletionFunc(consts.FlagCommand, cobra.NoFileCompletions))

	return cmd
}

func preRun(cmd *cobra.Command, _ []string) error {
	if err := config.Unmarshal("exec", &action); err != nil {
		return err
	}

	if err := util.DefaultSetup(cmd, action.Global, util.SetupOptions{}); err != nil {
		return err
	}
	if err := util.CreateJob(cmd.Context(), action.Global, util.SetupOptions{}); err != nil {
		return err
	}

	return nil
}

func run(cmd *cobra.Command, _ []string) error {
	return action.Run(cmd.Context())
}

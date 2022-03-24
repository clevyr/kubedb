package exec

import (
	"github.com/clevyr/kubedb/internal/command"
	"github.com/clevyr/kubedb/internal/config"
	"github.com/clevyr/kubedb/internal/util"
	"github.com/docker/cli/cli/streams"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"os"
)

var Command = &cobra.Command{
	Use:     "exec",
	Aliases: []string{"e", "shell"},
	Short:   "connect to an interactive shell",
	RunE:    run,
	PreRunE: preRun,
}

var conf config.Exec

func preRun(cmd *cobra.Command, args []string) error {
	return util.DefaultSetup(cmd, &conf.Global)
}

func run(cmd *cobra.Command, args []string) (err error) {
	log.WithField("pod", conf.Pod.Name).Info("exec into pod")

	stdin := streams.NewIn(os.Stdin)
	if err := stdin.SetRawTerminal(); err != nil {
		return err
	}
	defer stdin.RestoreTerminal()

	return conf.Client.Exec(conf.Pod, buildCommand(conf.Grammar, conf, args), stdin, os.Stdout, stdin.IsTerminal())
}

func buildCommand(db config.Databaser, conf config.Exec, args []string) []string {
	var cmd *command.Builder
	if len(args) == 0 {
		cmd = db.ExecCommand(conf)
	} else {
		cmd = command.NewBuilder("exec")
		for _, arg := range args {
			cmd.Push(arg)
		}
	}
	return []string{"sh", "-c", cmd.String()}
}

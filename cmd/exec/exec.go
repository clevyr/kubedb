package exec

import (
	"github.com/clevyr/kubedb/internal/command"
	"github.com/clevyr/kubedb/internal/config"
	"github.com/clevyr/kubedb/internal/util"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/tools/remotecommand"
	"k8s.io/kubectl/pkg/util/term"
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

	t := term.TTY{
		In:  os.Stdin,
		Out: os.Stdout,
	}
	t.Raw = t.IsTerminalIn()
	var sizeQueue remotecommand.TerminalSizeQueue
	if t.Raw {
		sizeQueue = t.MonitorSize(t.GetSize())
	}

	return t.Safe(func() error {
		return conf.Client.Exec(
			conf.Pod,
			buildCommand(conf.Grammar, conf, args).String(),
			t.In,
			t.Out,
			os.Stderr,
			t.IsTerminalIn(),
			sizeQueue,
		)
	})
}

func buildCommand(db config.Databaser, conf config.Exec, args []string) *command.Builder {
	if len(args) == 0 {
		return db.ExecCommand(conf)
	}

	cmd := command.NewBuilder("exec")
	for _, arg := range args {
		cmd.Push(arg)
	}
	return cmd
}

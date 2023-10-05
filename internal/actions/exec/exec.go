package exec

import (
	"context"
	"os"

	"github.com/clevyr/kubedb/internal/command"
	"github.com/clevyr/kubedb/internal/config"
	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/tools/remotecommand"
	"k8s.io/kubectl/pkg/util/term"
)

type Exec struct {
	config.Exec `mapstructure:",squash"`
	Args        []string
}

func (action Exec) Run(ctx context.Context) error {
	log.WithFields(log.Fields{
		"namespace": action.Client.Namespace,
		"name":      "pod/" + action.Pod.Name,
	}).Info("exec into pod")

	t := term.TTY{
		In:  os.Stdin,
		Out: os.Stdout,
	}
	t.Raw = t.IsTerminalIn()
	var sizeQueue remotecommand.TerminalSizeQueue
	if t.Raw {
		sizeQueue = t.MonitorSize(t.GetSize())
	}

	podCmd := action.buildCommand(action.Args)
	return t.Safe(func() error {
		return action.Client.Exec(
			ctx,
			action.Pod,
			podCmd.String(),
			t.In,
			t.Out,
			os.Stderr,
			t.IsTerminalIn(),
			sizeQueue,
		)
	})
}

func (action Exec) buildCommand(args []string) (cmd *command.Builder) {
	if len(args) == 0 {
		cmd = action.Dialect.ExecCommand(action.Exec)
	} else {
		cmd = command.NewBuilder("exec")
		for _, arg := range args {
			cmd.Push(arg)
		}
	}
	log.WithField("cmd", cmd).Trace("finished building command")
	return cmd
}

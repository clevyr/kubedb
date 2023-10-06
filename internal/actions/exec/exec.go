package exec

import (
	"context"
	"os"
	"time"

	"github.com/clevyr/kubedb/internal/command"
	"github.com/clevyr/kubedb/internal/config"
	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/tools/remotecommand"
	"k8s.io/kubectl/pkg/util/term"
)

type Exec struct {
	config.Exec `mapstructure:",squash"`
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

	return t.Safe(func() error {
		return action.Client.Exec(
			ctx,
			action.Pod,
			action.buildCommand().String(),
			t.In,
			t.Out,
			os.Stderr,
			t.IsTerminalIn(),
			sizeQueue,
			5*time.Second,
		)
	})
}

func (action Exec) buildCommand() *command.Builder {
	cmd := action.Dialect.ExecCommand(action.Exec)
	log.WithField("cmd", cmd).Trace("finished building command")
	return cmd
}

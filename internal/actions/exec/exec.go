package exec

import (
	"context"
	"os"

	"github.com/clevyr/kubedb/internal/command"
	"github.com/clevyr/kubedb/internal/config"
	"github.com/clevyr/kubedb/internal/kubernetes"
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
		"name":      "pod/" + action.DbPod.Name,
	}).Info("exec into database")

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
		return action.Client.Exec(ctx, kubernetes.ExecOptions{
			Pod:       action.JobPod,
			Cmd:       action.buildCommand().String(),
			Stdin:     t.In,
			Stdout:    t.Out,
			Stderr:    os.Stderr,
			TTY:       t.IsTerminalIn(),
			SizeQueue: sizeQueue,
		})
	})
}

func (action Exec) buildCommand() *command.Builder {
	cmd := action.Dialect.ExecCommand(action.Exec)
	log.WithField("cmd", cmd).Trace("finished building command")
	return cmd
}

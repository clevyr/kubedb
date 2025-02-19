package exec

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"gabe565.com/utils/slogx"
	"github.com/clevyr/kubedb/internal/command"
	"github.com/clevyr/kubedb/internal/config"
	"github.com/clevyr/kubedb/internal/consts"
	"github.com/clevyr/kubedb/internal/kubernetes"
	"github.com/clevyr/kubedb/internal/util"
	"github.com/spf13/viper"
	"k8s.io/client-go/tools/remotecommand"
	"k8s.io/kubectl/pkg/util/term"
)

type Exec struct {
	config.Exec `mapstructure:",squash"`
}

func (action Exec) Run(ctx context.Context) error {
	slog.Info("Exec into database",
		"namespace", action.Client.Namespace,
		"pod", action.DBPod.Name,
	)

	t := term.TTY{
		In:  os.Stdin,
		Out: os.Stdout,
	}
	stderr := os.Stderr
	t.Raw = t.IsTerminalIn()
	var sizeQueue remotecommand.TerminalSizeQueue
	if t.Raw {
		sizeQueue = t.MonitorSize(t.GetSize())
		// unset stderr because both stdout and stderr go over t.Out in raw mode
		stderr = nil
	}

	cmd, err := action.buildCommand()
	if err != nil {
		return err
	}

	return t.Safe(func() error {
		return action.Client.Exec(ctx, kubernetes.ExecOptions{
			Pod:       action.JobPod,
			Cmd:       cmd.String(),
			Stdin:     t.In,
			Stdout:    t.Out,
			Stderr:    stderr,
			TTY:       t.IsTerminalIn(),
			SizeQueue: sizeQueue,
		})
	})
}

func (action Exec) buildCommand() (*command.Builder, error) {
	db, ok := action.Dialect.(config.DBExecer)
	if !ok {
		return nil, fmt.Errorf("%w: %s", util.ErrNoExec, action.Dialect.Name())
	}

	cmd := db.ExecCommand(action.Exec)
	if opts := viper.GetString(consts.KeyOpts); opts != "" {
		cmd.Push(command.Split(opts))
	}

	slogx.Trace("Finished building command", "cmd", cmd)
	return cmd, nil
}

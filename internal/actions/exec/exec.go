package exec

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/clevyr/kubedb/internal/command"
	"github.com/clevyr/kubedb/internal/config"
	"github.com/clevyr/kubedb/internal/consts"
	"github.com/clevyr/kubedb/internal/kubernetes"
	"github.com/clevyr/kubedb/internal/util"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"k8s.io/client-go/tools/remotecommand"
	"k8s.io/kubectl/pkg/util/term"
)

type Exec struct {
	config.Exec `mapstructure:",squash"`
}

func (action Exec) Run(ctx context.Context) error {
	log.Info().
		Str("namespace", action.Client.Namespace).
		Str("pod", action.DBPod.Name).
		Msg("exec into database")

	t := term.TTY{
		In:  os.Stdin,
		Out: os.Stdout,
	}
	t.Raw = t.IsTerminalIn()
	var sizeQueue remotecommand.TerminalSizeQueue
	if t.Raw {
		sizeQueue = t.MonitorSize(t.GetSize())
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
			Stderr:    os.Stderr,
			TTY:       t.IsTerminalIn(),
			SizeQueue: sizeQueue,
		})
	})
}

func (action Exec) buildCommand() (*command.Builder, error) {
	db, ok := action.Dialect.(config.DatabaseExec)
	if !ok {
		return nil, fmt.Errorf("%w: %s", util.ErrNoExec, action.Dialect.Name())
	}

	cmd := db.ExecCommand(action.Exec)
	if opts := viper.GetString(consts.OptsKey); opts != "" {
		cmd.Push(command.Split(opts))
	}

	sanitized := strings.ReplaceAll(cmd.String(), action.Password, "***")
	log.Trace().Str("cmd", sanitized).Msg("finished building command")
	return cmd, nil
}

package redis

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/clevyr/kubedb/internal/command"
	"github.com/clevyr/kubedb/internal/config"
	"github.com/clevyr/kubedb/internal/kubernetes"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
)

var (
	_ config.DatabaseExec     = Redis{}
	_ config.DatabasePort     = Redis{}
	_ config.DatabasePassword = Redis{}
	_ config.DatabaseDb       = Redis{}
)

type Redis struct{}

func (Redis) Name() string {
	return "redis"
}

func (Redis) PortEnvNames() kubernetes.ConfigLookups {
	return kubernetes.ConfigLookups{kubernetes.LookupEnv{"REDIS_PORT"}}
}

func (Redis) DefaultPort() uint16 {
	return 6379
}

func (Redis) DatabaseEnvNames() kubernetes.ConfigLookups {
	return kubernetes.ConfigLookups{kubernetes.LookupEnv{"REDIS_DB"}}
}

func (db Redis) PodLabels() []kubernetes.LabelQueryable {
	return []kubernetes.LabelQueryable{
		kubernetes.LabelQueryAnd{
			kubernetes.LabelQuery{Name: "app.kubernetes.io/name", Value: "redis"},
			kubernetes.LabelQuery{Name: "app.kubernetes.io/component", Value: "master"},
		},
		db.sentinelQuery(),
		kubernetes.LabelQueryAnd{
			kubernetes.LabelQuery{Name: "app", Value: "redis"},
			kubernetes.LabelQuery{Name: "role", Value: "master"},
		},
	}
}

func (db Redis) FilterPods(ctx context.Context, client kubernetes.KubeClient, pods []v1.Pod) ([]v1.Pod, error) {
	preferred := make([]v1.Pod, 0, len(pods))

	if matched := db.sentinelQuery().FindPods(pods); len(matched) != 0 {
		log.Debug("querying Sentinel for primary instance")
		cmd := command.NewBuilder(
			command.Raw(`REDISCLI_AUTH="$REDIS_PASSWORD"`),
			"redis-cli", "-p", command.Raw(`"$REDIS_SENTINEL_PORT"`), "--raw",
			"SENTINEL", "MASTERS",
		)

		var buf bytes.Buffer
		var errBuf strings.Builder
		if err := client.Exec(ctx, kubernetes.ExecOptions{
			Pod:       matched[0],
			Container: "sentinel",
			Cmd:       cmd.String(),
			Stdout:    &buf,
			Stderr:    &errBuf,
		}); err != nil {
			return pods, fmt.Errorf("%w: %s", err, errBuf.String())
		}

		scanner := bufio.NewScanner(&buf)
		var primary string
		for scanner.Scan() {
			if scanner.Text() == "ip" && scanner.Scan() {
				primary = scanner.Text()
				break
			}
		}
		if scanner.Err() != nil {
			return pods, scanner.Err()
		}

		for _, pod := range pods {
			if strings.HasPrefix(primary, pod.Name+".") {
				preferred = append(preferred, pod)
				break
			}
		}
	}

	return preferred, nil
}

func (db Redis) PasswordEnvNames(c config.Global) kubernetes.ConfigLookups {
	return kubernetes.ConfigLookups{
		kubernetes.LookupEnv{"REDIS_PASSWORD"},
	}
}

func (Redis) ExecCommand(conf config.Exec) *command.Builder {
	cmd := command.NewBuilder(
		command.NewEnv("REDISCLI_AUTH", conf.Password),
		"exec", "redis-cli", "-h", conf.Host,
	)
	if conf.Port != 0 {
		cmd.Push("-p", strconv.Itoa(int(conf.Port)))
	}
	if conf.Database != "" {
		cmd.Push("-n", conf.Database)
	}
	if conf.DisableHeaders {
		cmd.Push("--raw")
	}
	if conf.Command != "" {
		cmd.Push(command.Split(conf.Command))
	}
	return cmd
}

func (Redis) sentinelQuery() kubernetes.LabelQueryAnd {
	return kubernetes.LabelQueryAnd{
		kubernetes.LabelQuery{Name: "app.kubernetes.io/name", Value: "redis"},
		kubernetes.LabelQuery{Name: "app.kubernetes.io/component", Value: "node"},
	}
}

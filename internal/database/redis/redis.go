package redis

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	"github.com/clevyr/kubedb/internal/command"
	"github.com/clevyr/kubedb/internal/config"
	"github.com/clevyr/kubedb/internal/kubernetes"
	"github.com/clevyr/kubedb/internal/kubernetes/filter"
	v1 "k8s.io/api/core/v1"
)

var (
	_ config.DBExecer      = Redis{}
	_ config.DBHasPort     = Redis{}
	_ config.DBHasPassword = Redis{}
	_ config.DBHasDatabase = Redis{}
)

type Redis struct{}

func (Redis) Name() string { return "redis" }

func (Redis) PrettyName() string { return "Redis" }

func (Redis) PortEnvs(_ config.Global) kubernetes.ConfigLookups {
	return kubernetes.ConfigLookups{kubernetes.LookupEnv{"REDIS_PORT"}}
}

func (Redis) PortDefault() uint16 { return 6379 }

func (Redis) DatabaseEnvs(_ config.Global) kubernetes.ConfigLookups {
	return kubernetes.ConfigLookups{kubernetes.LookupEnv{"REDIS_DB"}}
}

func (db Redis) PodFilters() filter.Filter {
	return filter.Or{
		filter.And{
			filter.Label{Name: "app.kubernetes.io/name", Value: "redis"},
			filter.Label{Name: "app.kubernetes.io/component", Value: "master"},
		},
		db.sentinelQuery(),
		filter.And{
			filter.Label{Name: "app", Value: "redis"},
			filter.Label{Name: "role", Value: "master"},
		},
		filter.Label{Name: "app.kubernetes.io/component", Value: "keydb"},
	}
}

func (db Redis) FilterPods(ctx context.Context, client kubernetes.KubeClient, pods []v1.Pod) ([]v1.Pod, error) {
	preferred := make([]v1.Pod, 0, len(pods))

	if matched := filter.Pods(pods, db.sentinelQuery()); len(matched) != 0 {
		slog.Debug("Querying Sentinel for primary instance")
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

func (db Redis) PasswordEnvs(_ config.Global) kubernetes.ConfigLookups {
	return kubernetes.ConfigLookups{
		kubernetes.LookupEnv{"REDIS_PASSWORD", "KEYDB_PASSWORD"},
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

func (Redis) sentinelQuery() filter.And {
	return filter.And{
		filter.Label{Name: "app.kubernetes.io/name", Value: "redis"},
		filter.Label{Name: "app.kubernetes.io/component", Value: "node"},
	}
}

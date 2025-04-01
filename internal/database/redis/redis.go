package redis

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	"github.com/clevyr/kubedb/internal/command"
	"github.com/clevyr/kubedb/internal/config/conftypes"
	"github.com/clevyr/kubedb/internal/kubernetes"
	"github.com/clevyr/kubedb/internal/kubernetes/filter"
	corev1 "k8s.io/api/core/v1"
)

var (
	_ conftypes.DBAliaser     = Redis{}
	_ conftypes.DBExecer      = Redis{}
	_ conftypes.DBHasPort     = Redis{}
	_ conftypes.DBHasPassword = Redis{}
	_ conftypes.DBHasDatabase = Redis{}
)

type Redis struct{}

func (Redis) Name() string { return "redis" }

func (Redis) PrettyName() string { return "Redis" }

func (Redis) Aliases() []string { return []string{"valkey", "keydb"} }

func (Redis) PortEnvs(_ *conftypes.Global) kubernetes.ConfigLookups {
	return kubernetes.ConfigLookups{kubernetes.LookupEnv{"REDIS_PORT", "VALKEY_PORT"}}
}

func (Redis) PortDefault() uint16 { return 6379 }

func (Redis) DatabaseEnvs(_ *conftypes.Global) kubernetes.ConfigLookups {
	return kubernetes.ConfigLookups{kubernetes.LookupEnv{"REDIS_DB"}}
}

func (db Redis) PodFilters() filter.Filter {
	return filter.Or{
		filter.And{
			filter.Label{Name: "app.kubernetes.io/name", Value: "redis"},
			filter.Label{Name: "app.kubernetes.io/component", Value: "master"},
		},
		filter.And{
			filter.Label{Name: "app.kubernetes.io/name", Value: "valkey"},
			filter.Label{Name: "app.kubernetes.io/component", Value: "primary"},
		},
		db.sentinelQuery(),
		filter.And{
			filter.Label{Name: "app", Value: "redis"},
			filter.Label{Name: "role", Value: "master"},
		},
		filter.Label{Name: "app.kubernetes.io/component", Value: "keydb"},
	}
}

func (db Redis) FilterPods(ctx context.Context, client kubernetes.KubeClient, pods []corev1.Pod) ([]corev1.Pod, error) {
	preferred := make([]corev1.Pod, 0, len(pods))

	if matched := filter.Pods(pods, db.sentinelQuery()); len(matched) != 0 {
		slog.Debug("Querying Sentinel for primary instance")
		cmd := command.NewBuilder(
			command.Raw(`REDISCLI_AUTH="${REDIS_PASSWORD:-$VALKEY_PASSWORD}"`),
			command.Raw(
				`"$(which redis-cli || which valkey-cli)"`,
			),
			"-p",
			command.Raw(`"${REDIS_SENTINEL_PORT:-$VALKEY_SENTINEL_PORT}"`),
			"--raw",
			"SENTINEL",
			"MASTERS",
		)

		var buf strings.Builder
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

		var primary string
		var returnNext bool
		for line := range strings.Lines(buf.String()) {
			if returnNext {
				primary = strings.TrimSuffix(line, "\n")
				break
			}
			if line == "ip\n" {
				returnNext = true
			}
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

func (db Redis) PasswordEnvs(_ *conftypes.Global) kubernetes.ConfigLookups {
	return kubernetes.ConfigLookups{
		kubernetes.LookupEnv{"REDIS_PASSWORD", "VALKEY_PASSWORD", "KEYDB_PASSWORD"},
		kubernetes.LookupSecretVolume{Name: "redis-password", Key: "redis-password"},
		kubernetes.LookupSecretVolume{Name: "valkey-password", Key: "valkey-password"},
	}
}

func (Redis) ExecCommand(conf *conftypes.Exec) *command.Builder {
	cmd := command.NewBuilder(
		"exec", command.Raw(`"$(which redis-cli || which valkey-cli)"`), "-h", conf.Host,
	)
	if conf.Password != "" {
		cmd.Unshift(command.NewEnv("REDISCLI_AUTH", conf.Password))
	}
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
		filter.Or{
			filter.Label{Name: "app.kubernetes.io/name", Value: "redis"},
			filter.Label{Name: "app.kubernetes.io/name", Value: "valkey"},
		},
		filter.Label{Name: "app.kubernetes.io/component", Value: "node"},
	}
}

package redis

import (
	"strconv"

	"github.com/clevyr/kubedb/internal/command"
	"github.com/clevyr/kubedb/internal/config"
	"github.com/clevyr/kubedb/internal/kubernetes"
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

func (Redis) PodLabels() []kubernetes.LabelQueryable {
	return []kubernetes.LabelQueryable{
		kubernetes.LabelQueryAnd{
			{Name: "app.kubernetes.io/name", Value: "redis"},
			{Name: "app.kubernetes.io/component", Value: "master"},
		},
		kubernetes.LabelQueryAnd{
			{Name: "app", Value: "redis"},
			{Name: "role", Value: "master"},
		},
	}
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

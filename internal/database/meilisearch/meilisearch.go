package meilisearch

import (
	_ "embed"
	"strconv"

	"github.com/clevyr/kubedb/internal/command"
	"github.com/clevyr/kubedb/internal/config/conftypes"
	"github.com/clevyr/kubedb/internal/database/sqlformat"
	"github.com/clevyr/kubedb/internal/kubernetes"
	"github.com/clevyr/kubedb/internal/kubernetes/filter"
)

var (
	_ conftypes.DBDumper        = Meilisearch{}
	_ conftypes.DBRestorer      = Meilisearch{}
	_ conftypes.DBHasPort       = Meilisearch{}
	_ conftypes.DBHasPassword   = Meilisearch{}
	_ conftypes.DBCanDisableJob = Meilisearch{}
)

type Meilisearch struct{}

func (Meilisearch) Name() string { return "meilisearch" }

func (Meilisearch) PrettyName() string { return "Meilisearch" }

func (Meilisearch) PortEnvs(_ *conftypes.Global) kubernetes.ConfigLookups {
	return kubernetes.ConfigLookups{}
}

func (Meilisearch) PortDefault() uint16 { return 7700 }

func (Meilisearch) PodFilters() filter.Filter {
	return filter.Label{Name: "app.kubernetes.io/name", Value: "meilisearch"}
}

func (Meilisearch) PasswordEnvs(_ *conftypes.Global) kubernetes.ConfigLookups {
	return kubernetes.ConfigLookups{
		kubernetes.LookupEnv{"MEILI_MASTER_KEY"},
	}
}

//go:embed dump.sh
var dumpScript string

func (Meilisearch) DumpCommand(conf *conftypes.Dump) *command.Builder {
	url := "http://" + conf.Host + ":" + strconv.Itoa(int(conf.Port))
	cmd := command.NewBuilder(
		command.NewEnv("API_HOST", url),
		"sh", "-c", dumpScript,
	)
	if conf.Password != "" {
		cmd.Unshift(command.NewEnv("MEILI_MASTER_KEY", conf.Password))
	}
	return cmd
}

//go:embed restore.sh
var restoreScript string

func (Meilisearch) RestoreCommand(conf *conftypes.Restore, _ sqlformat.Format) *command.Builder {
	cmd := command.NewBuilder("sh", "-c", restoreScript)
	if conf.Password != "" {
		cmd.Unshift(command.NewEnv("MEILI_MASTER_KEY", conf.Password))
	}
	return cmd
}

func (Meilisearch) Formats() map[sqlformat.Format]string {
	return map[sqlformat.Format]string{
		sqlformat.Plain: ".dump",
		sqlformat.Gzip:  ".dump.gz",
	}
}

func (Meilisearch) DisableJob() bool {
	return true
}

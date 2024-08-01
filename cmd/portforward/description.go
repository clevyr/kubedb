package portforward

import (
	"strings"

	"github.com/clevyr/kubedb/internal/config"
	"github.com/clevyr/kubedb/internal/database"
)

func newDescription() string {
	dbs := database.NamesForInterface[config.DBHasPort]()

	return `Set up a local port forward

Databases: ` + strings.Join(dbs, ", ")
}

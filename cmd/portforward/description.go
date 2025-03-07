package portforward

import (
	"strings"

	"github.com/clevyr/kubedb/internal/config/conftypes"
	"github.com/clevyr/kubedb/internal/database"
)

func newDescription() string {
	dbs := database.NamesForInterface[conftypes.DBHasPort]()

	return `Set up a local port forward.

Supported Databases:
  ` + strings.Join(dbs, ", ")
}

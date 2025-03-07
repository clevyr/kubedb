package exec

import (
	"strings"

	"github.com/clevyr/kubedb/internal/config/conftypes"
	"github.com/clevyr/kubedb/internal/database"
)

func newDescription() string {
	dbs := database.NamesForInterface[conftypes.DBExecer]()

	return `Connect to an interactive shell.

Supported Databases:
  ` + strings.Join(dbs, ", ")
}

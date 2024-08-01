package exec

import (
	"strings"

	"github.com/clevyr/kubedb/internal/config"
	"github.com/clevyr/kubedb/internal/database"
)

func newDescription() string {
	dbs := database.NamesForInterface[config.DBExecer]()

	return `Connect to an interactive shell.

Supported Databases:
  ` + strings.Join(dbs, ", ")
}

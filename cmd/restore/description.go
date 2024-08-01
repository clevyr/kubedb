package restore

import (
	"strings"

	"github.com/clevyr/kubedb/internal/config"
	"github.com/clevyr/kubedb/internal/database"
)

func newDescription() string {
	dbs := database.NamesForInterface[config.DBRestorer]()

	return `Restore a sql file to a database.

Databases: ` + strings.Join(dbs, ", ") + `

Supported Input Filetypes:
  - Raw sql file. Typically with the ` + "`" + `.sql` + "`" + ` extension
  - Gzipped sql file. Typically with the ".sql.gz" extension
  - For Postgres: custom dump file. Typically with the ".dmp" extension`
}

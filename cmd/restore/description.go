package restore

import (
	"strings"

	"github.com/clevyr/kubedb/internal/config/conftypes"
	"github.com/clevyr/kubedb/internal/database"
)

func newDescription() string {
	dbs := database.NamesForInterface[conftypes.DBRestorer]()

	return `Restore a sql file to a database.

Supported Databases:
  ` + strings.Join(dbs, ", ") + `

File Path:
  - Raw sql file. Typically with a ".sql" file extension
  - Gzipped sql file. Typically with a ".sql.gz" file extension
  - For Postgres: custom dump file. Typically with a ".dmp" file extension

Cloud Download:
  - Use "s3://" for S3, "gs://" for GCS, or "b2://" for Backblaze B2.
  - Cloud config is loaded from the environment (similar to the aws and gcloud tools).`
}

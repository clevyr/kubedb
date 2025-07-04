package dump

import (
	"strings"

	"github.com/clevyr/kubedb/internal/config/conftypes"
	"github.com/clevyr/kubedb/internal/database"
)

func newDescription() string {
	dbs := database.NamesForInterface[conftypes.DBDumper]()

	return `Dump a database to a sql file.

Supported Databases:
  ` + strings.Join(dbs, ", ") + `

File Path:
  - If the path is not provided, a filename will be generated.
  - If the path is a file, the database will be dumped there.
  - If the path is a directory, the database will be dumped to a generated filename in that directory.
  - Filenames are autogenerated based on the namespace and timestamp.

Cloud Upload:
	- Use "s3://" for S3, "gs://" for GCS, or "b2://" for Backblaze B2.
  - If the URL only contains a bucket name or if the path ends with "/", then filenames are autogenerated similarly to local dumps.
  - Cloud config is loaded from the environment (similar to the aws and gcloud tools).
`
}

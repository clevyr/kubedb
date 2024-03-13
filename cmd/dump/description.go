package dump

import (
	"strings"
	"time"

	"github.com/clevyr/kubedb/internal/actions/dump"
	"github.com/clevyr/kubedb/internal/config"
	"github.com/clevyr/kubedb/internal/database"
)

func newDescription() string {
	helpFilename := dump.Filename{
		Namespace: "clevyr",
		Ext:       ".sql.gz",
		Date:      time.Date(2022, 1, 9, 9, 41, 0, 0, time.UTC),
	}.Generate()

	dbs := database.NamesForInterface[config.DatabaseDump]()

	return `Dump a database to a sql file.

Databases: ` + strings.Join(dbs, ", ") + `

Filenames:  
  If a filename is provided, and it does not end with a "/", then it will be used verbatim.
  Otherwise, the filename will be generated and appended to the given path.
  For example, if a dump is performed in the namespace "clevyr" with no extra flags,
  the generated filename might look like "` + helpFilename + `".
  Similarly, if the filename is passed as "backups/", then the generated path might look like
  "backups/` + helpFilename + `".

Cloud Upload:  
  KubeDB will directly upload the dump to a cloud storage bucket if the output path starts with a URL scheme:
    - S3 upload schema is "s3://".
    - GCS upload schema is "gs://".
  Cloud configuration will be loaded from the environment, similarly to the official aws and gcloud tools.

  Note the above section on filenames. For example, if the filename is set to "s3://clevyr-backups/dev/",
  then the resulting filename might look like "s3://clevyr-backups/dev/` + helpFilename + `".
  The only exception is if a bucket name is provided without any sub-path (like "s3://backups"), then
  the generated filename will be appended without requiring an ending "/".
`
}

package dump

import "github.com/clevyr/kubedb/internal/actions/dump"

func newDescription() string {
	return `Dump a database to a sql file.

Filenames:  
  If a filename is provided, and it does not end with a "/", then it will be used verbatim.
  Otherwise, the filename will be generated and appended to the given path.
  For example, if a dump is performed in the namespace "clevyr" with no extra flags,
  the generated filename might look like "` + dump.HelpFilename() + `".
  Similarly, if the filename is passed as "backups/", then the generated path might look like
  "backups/` + dump.HelpFilename() + `".

Cloud Upload:  
  KubeDB will directly upload the dump to a cloud storage bucket if the output path starts with a URL scheme:
    - S3 upload schema is "s3://".
    - GCS upload schema is "gs://".
  Cloud configuration will be loaded from the environment, similarly to the official aws and gcloud tools.

  Note the above section on filenames. For example, if the filename is set to "s3://clevyr-backups/dev/",
  then the resulting filename might look like "s3://clevyr-backups/dev/` + dump.HelpFilename() + `".
  The only exception is if a bucket name is provided without any sub-path (like "s3://backups"), then
  the generated filename will be appended without requiring an ending "/".
`
}

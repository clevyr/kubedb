package restore

func newDescription() string {
	return `Restore a sql file to a database.

Supported Input Filetypes:
  - Raw sql file. Typically with the ` + "`" + `.sql` + "`" + ` extension
  - Gzipped sql file. Typically with the ".sql.gz" extension
  - For Postgres: custom dump file. Typically with the ".dmp" extension`
}

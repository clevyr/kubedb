## kubedb dump

Dump a database to a sql file

### Synopsis

Dump a database to a sql file.

If no filename is provided, the filename will be generated.
For example, if a dump is performed in the namespace "clevyr" with no extra flags,
the generated filename might look like "clevyr_2022-01-09_094100.sql.gz"

```
kubedb dump [filename] [flags]
```

### Options

```
  -c, --clean                           Clean (drop) database objects before recreating (default true)
  -d, --dbname string                   Database name to use (default discovered)
  -C, --directory string                Directory to dump to (default ".")
  -T, --exclude-table strings           Do NOT dump the specified table(s)
  -D, --exclude-table-data strings      Do NOT dump data for the specified table(s)
  -F, --format string                   Output file format One of (gzip|custom|plain) (default "gzip")
  -h, --help                            help for dump
      --if-exists                       Use IF EXISTS when dropping objects (default true)
      --job-pod-labels stringToString   Pod labels to add to the job (default [])
      --no-job                          Database commands will be run in the database pod instead of a dedicated job
  -O, --no-owner                        Skip restoration of object ownership in plain-text format (default true)
  -p, --password string                 Database password (default discovered)
      --port uint16                     Database port (default discovered)
  -q, --quiet                           Silence remote log output
      --remote-gzip                     Compress data over the wire. Results in lower bandwidth usage, but higher database load. May improve speed on slow connections. (default true)
  -t, --table strings                   Dump the specified table(s) only
  -U, --username string                 Database username (default discovered)
```

### Options inherited from parent commands

```
      --context string      Kubernetes context name
      --dialect string      Database dialect. One of (postgres|mariadb|mongodb) (default discovered)
      --kubeconfig string   Paths to the kubeconfig file (default "$HOME/.kube/config")
      --log-format string   Log formatter. One of (text|json) (default "text")
      --log-level string    Log level. One of (trace|debug|info|warning|error|fatal|panic) (default "info")
  -n, --namespace string    Kubernetes namespace
      --pod string          Perform detection from a pod instead of searching the namespace
```

### SEE ALSO

* [kubedb](kubedb.md)	 - Painlessly work with databases in Kubernetes.


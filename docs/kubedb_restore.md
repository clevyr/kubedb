## kubedb restore

Restore a sql file to a database

### Synopsis

Restore a sql file to a database.

Supported Input Filetypes:
  - Raw sql file. Typically with the `.sql` extension
  - Gzipped sql file. Typically with the ".sql.gz" extension
  - For Postgres: custom dump file. Typically with the ".dmp" extension

```
kubedb restore filename [flags]
```

### Options

```
      --analyze                         Run an analyze query after restore (default true)
  -c, --clean                           Clean (drop) database objects before recreating (default true)
  -d, --dbname string                   Database name to use (default discovered)
  -f, --force                           Do not prompt before restore
  -F, --format string                   Output file format One of (gzip|custom|plain) (default "gzip")
      --halt-on-error                   Halt on error (Postgres only) (default true)
  -h, --help                            help for restore
      --job-pod-labels stringToString   Pod labels to add to the job (default [])
      --no-job                          Database commands will be run in the database pod instead of a dedicated job
  -O, --no-owner                        Skip restoration of object ownership in plain-text format (default true)
      --opts string                     Additional options to pass to the database client command
  -p, --password string                 Database password (default discovered)
      --port uint16                     Database port (default discovered)
  -q, --quiet                           Silence remote log output
      --remote-gzip                     Compress data over the wire. Results in lower bandwidth usage, but higher database load. May improve speed on slow connections. (default true)
  -1, --single-transaction              Restore as a single transaction (default true)
  -U, --username string                 Database username (default discovered)
```

### Options inherited from parent commands

```
      --context string                 Kubernetes context name
      --dialect string                 Database dialect. One of (postgres|mariadb|mongodb) (default discovered)
      --healthchecks-ping-url string   Notification handler URL
      --kubeconfig string              Paths to the kubeconfig file (default "$HOME/.kube/config")
      --log-format string              Log formatter. One of (text|json) (default "text")
      --log-level string               Log level. One of (trace|debug|info|warning|error|fatal|panic) (default "info")
  -n, --namespace string               Kubernetes namespace
      --pod string                     Perform detection from a pod instead of searching the namespace
```

### SEE ALSO

* [kubedb](kubedb.md)	 - Painlessly work with databases in Kubernetes.


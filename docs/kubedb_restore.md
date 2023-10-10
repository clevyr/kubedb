## kubedb restore

Restore a database from a sql file

### Synopsis

The "restore" command restores a given sql file to a running database pod.

Supported Input Filetypes:
  - Raw sql file. Typically with the ".sql" extension
  - Gzipped sql file. Typically with the ".sql.gz" extension
  - Postgres custom dump file. Typically with the ".dmp" extension (Only if the target database is Postgres)

```
kubedb restore filename [flags]
```

### Options

```
      --analyze                         Run an analyze query after restore (default true)
  -c, --clean                           Clean (drop) database objects before recreating (default true)
  -d, --dbname string                   Database name to connect to
  -f, --force                           Do not prompt before restore
  -F, --format string                   Output file format ([g]zip, [c]ustom, [p]lain) (default "gzip")
  -h, --help                            help for restore
      --job-pod-labels stringToString   Pod labels to add to the job (default [])
      --no-job                          Database commands will be run in the database pod instead of a dedicated job
  -O, --no-owner                        Skip restoration of object ownership in plain-text format (default true)
  -p, --password string                 Database password
  -q, --quiet                           Silence remote log output
      --remote-gzip                     Compress data over the wire. Results in lower bandwidth usage, but higher database load. May improve speed on fast connections. (default true)
  -1, --single-transaction              Restore as a single transaction (default true)
  -U, --username string                 Database username
```

### Options inherited from parent commands

```
      --context string      The name of the kubeconfig context to use
      --dialect string      Database dialect. Detected if not set. (postgres, mariadb, mongodb)
      --kubeconfig string   Path to the kubeconfig file (default "$HOME/.kube/config")
      --log-format string   Log formatter (text, json) (default "text")
      --log-level string    Log level (trace, debug, info, warning, error, fatal, panic) (default "info")
  -n, --namespace string    The Kubernetes namespace scope
      --pod string          Force a specific pod. If this flag is set, dialect is required.
```

### SEE ALSO

* [kubedb](kubedb.md)	 - interact with a database inside of Kubernetes


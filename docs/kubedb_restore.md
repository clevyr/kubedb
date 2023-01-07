## kubedb restore

restore a database from a sql file

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
  -c, --clean                clean (drop) database objects before recreating (default true)
  -f, --force                do not prompt before restore
  -F, --format string        output file format ([g]zip, [c]ustom, [p]lain) (default "gzip")
  -h, --help                 help for restore
  -O, --no-owner             skip restoration of object ownership in plain-text format (default true)
  -q, --quiet                Silence remote log output
  -1, --single-transaction   restore as a single transaction (default true)
```

### Options inherited from parent commands

```
      --context string      name of the kubeconfig context to use
  -d, --dbname string       database name to connect to
      --dialect string      database dialect. detected if not set. (postgres, mariadb, mongodb)
      --kubeconfig string   absolute path to the kubeconfig file (default "$HOME/.kube/config")
      --log-format string   log formatter (text, json) (default "text")
      --log-level string    log level (trace, debug, info, warning, error, fatal, panic) (default "info")
  -n, --namespace string    the namespace scope for this CLI request
  -p, --password string     database password
      --pod string          force a specific pod. if this flag is set, dialect is required.
  -U, --username string     database username
```

### SEE ALSO

* [kubedb](kubedb.md)	 - interact with a database inside of Kubernetes


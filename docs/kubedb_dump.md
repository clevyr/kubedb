## kubedb dump

dump a database to a sql file

### Synopsis

The "dump" command dumps a running database to a sql file.

If no filename is provided, the filename will be generated.
For example, if a dump is performed in the namespace "clevyr" with no extra flags,
the generated filename might look like "clevyr_2022-01-09_094100.sql.gz"

```
kubedb dump [filename] [flags]
```

### Options

```
  -c, --clean                        clean (drop) database objects before recreating (default true)
  -T, --exclude-table strings        do NOT dump the specified table(s)
      --exclude-table-data strings   do NOT dump data for the specified table(s)
  -F, --format string                output file format ([g]zip, [c]ustom, [p]lain) (default "g")
  -h, --help                         help for dump
      --if-exists                    use IF EXISTS when dropping objects (default true)
  -O, --no-owner                     skip restoration of object ownership in plain-text format (default true)
  -t, --table strings                dump the specified table(s) only
```

### Options inherited from parent commands

```
      --context string      name of the kubeconfig context to use
  -d, --dbname string       database name to connect to
  -C, --directory string    dir to hold the generated config (default "./docs")
      --grammar string      database grammar. detected if not set. (postgres, mariadb)
      --kubeconfig string   absolute path to the kubeconfig file (default "/Users/gabe565/.kube/config")
      --log-format string   log formatter (text, json) (default "text")
      --log-level string    log level (trace, debug, info, warning, error, fatal, panic) (default "info")
  -n, --namespace string    the namespace scope for this CLI request
  -p, --password string     database password
      --pod string          force a specific pod. if this flag is set, grammar is required.
  -U, --username string     database username
```

### SEE ALSO

* [kubedb](kubedb.md)	 - interact with a database inside of Kubernetes

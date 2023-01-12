## kubedb ui

launch a terminal UI

```
kubedb ui [flags]
```

### Options

```
  -h, --help   help for ui
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


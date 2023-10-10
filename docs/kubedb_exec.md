## kubedb exec

Connect to an interactive shell

```
kubedb exec [flags]
```

### Options

```
  -c, --command string                  Run a single command and exit
  -d, --dbname string                   Database name to connect to
  -h, --help                            help for exec
      --job-pod-labels stringToString   Pod labels to add to the job (default [])
      --no-job                          Database commands will be run in the database pod instead of a dedicated job
  -p, --password string                 Database password
      --port uint16                     Database port
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


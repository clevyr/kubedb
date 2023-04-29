## kubedb exec

Connect to an interactive shell

```
kubedb exec [flags]
```

### Options

```
  -h, --help   help for exec
```

### Options inherited from parent commands

```
      --context string      The name of the kubeconfig context to use
  -d, --dbname string       Database name to connect to
      --dialect string      Database dialect. Detected if not set. (postgres, mariadb, mongodb)
      --kubeconfig string   Path to the kubeconfig file (default "$HOME/.kube/config")
      --log-format string   Log formatter (text, json) (default "text")
      --log-level string    Log level (trace, debug, info, warning, error, fatal, panic) (default "info")
  -n, --namespace string    The Kubernetes namespace scope
  -p, --password string     Database password
      --pod string          Force a specific pod. If this flag is set, dialect is required.
  -U, --username string     Database username
```

### SEE ALSO

* [kubedb](kubedb.md)	 - interact with a database inside of Kubernetes


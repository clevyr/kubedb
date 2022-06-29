## kubedb port-forward

set up a local port forward

```
kubedb port-forward [local_port] [flags]
```

### Options

```
      --address strings   addresses to listen on (comma separated) (default [127.0.0.1,::1])
  -h, --help              help for port-forward
```

### Options inherited from parent commands

```
      --context string      name of the kubeconfig context to use
  -d, --dbname string       database name to connect to
      --dialect string      database dialect. detected if not set. (postgres, mariadb, mongodb)
  -C, --directory string    dir to hold the generated config (default "./docs")
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


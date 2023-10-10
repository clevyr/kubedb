## kubedb port-forward

Set up a local port forward

```
kubedb port-forward [local_port] [flags]
```

### Options

```
      --address strings   Addresses to listen on (comma separated) (default [127.0.0.1,::1])
  -h, --help              help for port-forward
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


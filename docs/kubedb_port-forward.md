## kubedb port-forward

Set up a local port forward

### Synopsis

Set up a local port forward.

Supported Databases:
  postgres, mariadb, mongodb, redis, meilisearch

```
kubedb port-forward [local_port] [flags]
```

### Options

```
      --address strings      Local listen address (default [127.0.0.1,::1])
  -h, --help                 help for port-forward
      --listen-port uint16   Local listen port (default discovered)
      --port uint16          Database port (default discovered)
```

### Options inherited from parent commands

```
      --context string                 Kubernetes context name
      --dialect string                 Database dialect. (one of postgres, mariadb, mongodb, redis, meilisearch) (default discovered)
      --healthchecks-ping-url string   Notification handler URL
      --kubeconfig string              Paths to the kubeconfig file (default "$HOME/.kube/config")
      --log-format string              Log format (one of auto, color, plain, json) (default "auto")
      --log-level string               Log level (one of trace, debug, info, warn, error) (default "info")
  -n, --namespace string               Kubernetes namespace
      --pod string                     Perform detection from a pod instead of searching the namespace
```

### SEE ALSO

* [kubedb](kubedb.md)	 - Painlessly work with databases in Kubernetes.


## kubedb port-forward

Set up a local port forward

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


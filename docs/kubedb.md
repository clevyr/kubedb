## kubedb

Painlessly work with databases in Kubernetes.

### Synopsis

Painlessly work with databases in Kubernetes.

Supported Databases:
  postgres, mariadb, mongodb, redis, meilisearch

### Options

```
      --context string                 Kubernetes context name
      --dialect string                 Database dialect. (one of postgres, mariadb, mongodb, redis, meilisearch) (default discovered)
      --healthchecks-ping-url string   Notification handler URL
  -h, --help                           help for kubedb
      --kubeconfig string              Paths to the kubeconfig file (default "$HOME/.kube/config")
      --log-format string              Log format (one of auto, color, plain, json) (default "auto")
      --log-level string               Log level (one of trace, debug, info, warn, error) (default "info")
  -n, --namespace string               Kubernetes namespace
      --pod string                     Perform detection from a pod instead of searching the namespace
  -v, --version                        version for kubedb
```

### SEE ALSO

* [kubedb dump](kubedb_dump.md)	 - Dump a database to a sql file
* [kubedb exec](kubedb_exec.md)	 - Connect to an interactive shell
* [kubedb port-forward](kubedb_port-forward.md)	 - Set up a local port forward
* [kubedb restore](kubedb_restore.md)	 - Restore a sql file to a database
* [kubedb status](kubedb_status.md)	 - View connection status


## kubedb

Painlessly work with databases in Kubernetes.

### Options

```
      --context string      Kubernetes context name
      --dialect string      Database dialect. One of (postgres|mariadb|mongodb) (default discovered)
  -h, --help                help for kubedb
      --kubeconfig string   Paths to the kubeconfig file (default "$HOME/.kube/config")
      --log-format string   Log formatter. One of (text|json) (default "text")
      --log-level string    Log level. One of (trace|debug|info|warning|error|fatal|panic) (default "info")
  -n, --namespace string    Kubernetes namespace
      --pod string          Perform detection from a pod instead of searching the namespace
  -v, --version             version for kubedb
```

### SEE ALSO

* [kubedb dump](kubedb_dump.md)	 - Dump a database to a sql file
* [kubedb exec](kubedb_exec.md)	 - Connect to an interactive shell
* [kubedb port-forward](kubedb_port-forward.md)	 - Set up a local port forward
* [kubedb restore](kubedb_restore.md)	 - Restore a sql file to a database


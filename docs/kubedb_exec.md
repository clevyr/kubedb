## kubedb exec

Connect to an interactive shell

### Synopsis

Connect to an interactive shell

Databases: postgres, mariadb, mongodb, redis

```
kubedb exec [flags]
```

### Options

```
  -c, --command string                  Run a single command and exit
      --create-job                      Create a job that will run the database client (default true)
      --create-network-policy           Creates a network policy allowing the KubeDB job to talk to the database. (default true)
  -d, --dbname string                   Database name to use (default discovered)
  -h, --help                            help for exec
      --job-pod-labels stringToString   Pod labels to add to the job (default [])
      --opts string                     Additional options to pass to the database client command
  -p, --password string                 Database password (default discovered)
      --port uint16                     Database port (default discovered)
  -U, --username string                 Database username (default discovered)
```

### Options inherited from parent commands

```
      --context string                 Kubernetes context name
      --dialect string                 Database dialect. One of (postgres|mariadb|mongodb) (default discovered)
      --healthchecks-ping-url string   Notification handler URL
      --kubeconfig string              Paths to the kubeconfig file (default "$HOME/.kube/config")
      --log-format string              Log formatter. One of (text|json) (default "text")
      --log-level string               Log level. One of (trace|debug|info|warn|error|fatal|panic) (default "info")
  -n, --namespace string               Kubernetes namespace
      --pod string                     Perform detection from a pod instead of searching the namespace
```

### SEE ALSO

* [kubedb](kubedb.md)	 - Painlessly work with databases in Kubernetes.


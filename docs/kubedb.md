## kubedb

interact with a database inside of Kubernetes

### Synopsis

kubedb is a command to interact with a database running in a Kubernetes cluster.

Multiple database types (referred to as the "dialect") are supported.
If the dialect is not configured via flag, it will be detected dynamically.

Supported Database Dialects:
  - PostgreSQL
  - MariaDB
  - MongoDB

If not configured via flag, some configuration variables will be loaded from the target pod's env vars.

Dynamic Env Var Variables:
  - Port
  - Database
  - Username (fallback value: "postgres" if PostgreSQL, "mariadb" if MariaDB, "root" if MongoDB)
  - Password


### Options

```
      --context string      The name of the kubeconfig context to use
      --dialect string      Database dialect. Detected if not set. (postgres, mariadb, mongodb)
  -h, --help                help for kubedb
      --kubeconfig string   Path to the kubeconfig file (default "$HOME/.kube/config")
      --log-format string   Log formatter (text, json) (default "text")
      --log-level string    Log level (trace, debug, info, warning, error, fatal, panic) (default "info")
  -n, --namespace string    The Kubernetes namespace scope
      --pod string          Force a specific pod. If this flag is set, dialect is required.
  -v, --version             version for kubedb
```

### SEE ALSO

* [kubedb dump](kubedb_dump.md)	 - Dump a database to a sql file
* [kubedb exec](kubedb_exec.md)	 - Connect to an interactive shell
* [kubedb port-forward](kubedb_port-forward.md)	 - Set up a local port forward
* [kubedb restore](kubedb_restore.md)	 - Restore a database from a sql file


## kubedb

interact with a database inside of Kubernetes

### Synopsis

kubedb is a command to interact with a database running in a Kubernetes cluster.

Multiple database types (referred to as the "grammar") are supported.
If the grammar is not configured via flag, it will be detected dynamically.

Supported Database Grammars:
  - PostgreSQL
  - MariaDB

If not configured via flag, some configuration variables will be loaded from the target pod's env vars.

Dynamic Env Var Variables:
  - Database (fallback value: "db")
  - Username (fallback value: "db" if PostgreSQL, "mariadb" if MariaDB)
  - Password (required)


### Options

```
      --context string      name of the kubeconfig context to use
  -d, --dbname string       database name to connect to
  -C, --directory string    dir to hold the generated config (default "./docs")
      --grammar string      database grammar. detected if not set. (postgres, mariadb)
  -h, --help                help for kubedb
      --kubeconfig string   absolute path to the kubeconfig file (default "/Users/gabe565/.kube/config")
      --log-format string   log formatter (text, json) (default "text")
      --log-level string    log level (trace, debug, info, warning, error, fatal, panic) (default "info")
  -n, --namespace string    the namespace scope for this CLI request
  -p, --password string     database password
      --pod string          force a specific pod. if this flag is set, grammar is required.
  -U, --username string     database username
```

### SEE ALSO

* [kubedb dump](kubedb_dump.md)	 - dump a database to a sql file
* [kubedb exec](kubedb_exec.md)	 - connect to an interactive shell
* [kubedb port-forward](kubedb_port-forward.md)	 - set up a local port forward
* [kubedb restore](kubedb_restore.md)	 - restore a database from a sql file

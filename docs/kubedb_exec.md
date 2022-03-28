## kubedb exec

connect to an interactive shell

```
kubedb exec [flags]
```

### Options

```
  -h, --help   help for exec
```

### Options inherited from parent commands

```
      --context string      name of the kubeconfig context to use
  -d, --dbname string       database name to connect to
  -C, --directory string    dir to hold the generated config (default "./docs")
      --grammar string      database grammar. detected if not set. (postgres, mariadb)
      --kubeconfig string   absolute path to the kubeconfig file (default "$HOME/.kube/config")
      --log-format string   log formatter (text, json) (default "text")
      --log-level string    log level (trace, debug, info, warning, error, fatal, panic) (default "info")
  -n, --namespace string    the namespace scope for this CLI request
  -p, --password string     database password
      --pod string          force a specific pod. if this flag is set, grammar is required.
  -U, --username string     database username
```

### SEE ALSO

* [kubedb](kubedb.md)	 - interact with a database inside of Kubernetes


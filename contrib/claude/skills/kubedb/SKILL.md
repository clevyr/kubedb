---
name: kubedb
description: Use kubedb (not kubectl exec one-liners) whenever you need to query, inspect, dump, restore, or port-forward a database running in Kubernetes. Triggers on any request touching a Postgres, MariaDB/MySQL, MongoDB, Redis/Valkey, or Meilisearch instance in a cluster, including "run this SQL in the cluster", "check a table in prod", "dump the staging DB", "restore this backup", or "connect to the database pod".
---

# kubedb

`kubedb` is the default way to work with databases in Kubernetes. It discovers the pod,
dialect, and credentials itself. One-liners that read and decode secrets are usually
denied by permission checks, so don't read database secrets yourself. Try kubedb first,
and fall back to kubectl only if kubedb can't do the task, saying why.

## Rule

If a command would pair `kubectl exec` with a database client (`psql`, `pg_dump`, `mysql`,
`mariadb`, `mongosh`, `redis-cli`), or pipe a secret into one, use `kubedb` instead.

| Instead of                                                 | Use                                   |
| ---------------------------------------------------------- | ------------------------------------- |
| `kubectl exec -n ns pg-0 -- psql -U app -c "SELECT..."`    | `kubedb exec -n ns -c "SELECT..."`    |
| `kubectl exec ... -- pg_dump ... \| gzip > dump.sql.gz`    | `kubedb dump -n ns dump.sql.gz`       |
| `gunzip -c dump.sql.gz \| kubectl exec -i ... -- psql ...` | `kubedb restore -n ns dump.sql.gz`    |
| `kubectl port-forward svc/postgres 5432:5432`              | `kubedb port-forward -n ns`           |
| `kubectl get secret ... \| base64 -d`                      | nothing, kubedb finds credentials     |

If `command -v kubedb` finds nothing, install it with `brew install clevyr/tap/kubedb`
or ask the user before reaching for kubectl.

## Common flags

- Always pass `-n` and `--context` so the transcript shows which cluster was touched.
- `--pod` picks a specific pod when discovery chooses the wrong one.
- `--dialect` forces the database type when a namespace has more than one.
- `--log-level warn` hides kubedb's own log lines when you need to parse output.

## Check the connection

```sh
kubedb status --context <ctx> -n <namespace>
```

Shows the detected pod and dialect, and whether kubedb can connect. You don't need it
before every command. Use it when the target is ambiguous, or to diagnose a failed command.

## Run a query

```sh
kubedb exec --context <ctx> -n <namespace> -c "<query>"
```

- Always pass `-c`. A bare `kubedb exec` opens an interactive shell and hangs.
- Stdin is forwarded, so `kubedb exec -n ns < script.sql` runs a script.
- `-d` picks a database other than the discovered one.
- `--opts` passes raw client flags. For clean output use `--opts=-tA` with Postgres,
  `--opts="--skip-column-names --batch"` with MariaDB, or `--opts=--quiet` with MongoDB.

```sh
kubedb exec -n prod --log-level warn --opts=-tA -c "SELECT count(*) FROM users"
kubedb exec -n staging -c "SHOW TABLES"
kubedb exec -n prod -c 'db.orders.find({status:"failed"}).limit(5)'
kubedb exec -n cache -c "INFO memory"
```

## Read-only queries on CloudNativePG: use a replica

kubedb targets the primary by default. For read-only queries on a CloudNativePG (CNPG)
cluster, target a replica and skip the Job:

```sh
kubectl --context <ctx> -n <namespace> get clusters.postgresql.cnpg.io   # cluster name
REPLICA=$(kubectl --context <ctx> -n <namespace> get pods \
  -l cnpg.io/cluster=<cluster>,cnpg.io/instanceRole=replica \
  --field-selector=status.phase=Running -o jsonpath='{.items[0].metadata.name}')
kubedb exec --context <ctx> -n <namespace> --pod "$REPLICA" --create-job=false -c "SELECT ..."
```

- Look up the replica every time. Pod names don't show the role, and failovers move the primary.
- If the lookup is empty, the cluster has one instance. Drop `--pod` and `--create-job=false`.
- Replicas reject writes and can't block migrations on the primary, so skipping the Job is safe here.
- Add a `LIMIT` to exploratory queries. Without a Job, `psql` runs inside the replica's
  container and a huge result set uses its memory.
- Replicas can lag slightly. Use the primary when you need a write that just happened.

Everything else keeps the default Job: primaries, writes, dumps, restores, and other databases.

## Dump

```sh
kubedb dump --context <ctx> -n <namespace> [path | s3://bucket/ | gs://bucket/ | b2://bucket/]
```

- With no path, or a directory or bucket ending in `/`, the file name is generated.
- `-F` is `gzip` by default. `plain` and Postgres-only `custom` are also available.
- `-t` dumps only listed tables, `-T` skips tables, and `-D` keeps schema but skips data.
- `-q --progress=false` keeps output short.

## Restore

```sh
kubedb restore --context <ctx> -n <namespace> <file.sql | file.sql.gz | file.dmp | s3://...>
```

Restore is destructive because `--clean` is on by default. Confirm the exact context and
namespace with the user first, and pass `-f` to skip the prompt only after they approve.

## Port-forward

```sh
kubedb port-forward --context <ctx> -n <namespace> [local_port]
```

Use this when a local client or GUI needs a connection. The local port is deliberately not
the native port, to avoid conflicts and accidental connections. It defaults to 30000 plus the
native port, such as 35432 for Postgres, and is logged at startup. Read it from the log or
pass one explicitly. Run it in the background and stop it when done.

## Why the Job

By default kubedb runs the client in a short-lived Job with a NetworkPolicy, not inside the
database pod. This keeps a stuck or disconnected command from lingering in the database
container, where a zombie `pg_dump` has blocked app migrations before. Only pass
`--create-job=false` for replica reads as above, or when the cluster forbids Jobs and the
user agrees.

Config lives at `~/.config/kubedb/config.yaml`, and any flag can be set with a `KUBEDB_`
environment variable. Full reference: https://github.com/clevyr/kubedb/blob/main/docs/kubedb.md

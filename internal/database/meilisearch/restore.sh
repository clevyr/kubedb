#!/usr/bin/env sh
set -euo pipefail

cleanup_fs() {
  echo 'Cleaning up' >&2
  rm -rf data.ms.restore
  if [[ -n "$dump" ]]; then
    rm -f "$dump"
  fi
}

trap 'cleanup_fs' EXIT

echo 'Uploading dump' >&2
now="$(date +%Y%m%d-%H%M%S)"
dump="dumps/$now.dump"
cat > "$dump"

restore="restore_${now}_data.ms"
printf 'Creating new database "%s" from dump\n' "$restore" >&2
port="$(( RANDOM + 32767 ))"
meilisearch --import-dump "$dump" --db-path "$restore" --http-addr 127.0.0.1:"$port" &
restore_pid="$!"
while ! nc -z 127.0.0.1 "$port"; do
  if ! kill -0 "$restore_pid"; then
    exit 1
  fi
  sleep 1
done
echo 'Restore finished' >&2

echo 'Moving "data.ms" to "old_data.ms"' >&2
rm -rf old_data.ms
mv data.ms old_data.ms
printf 'Moving "%s" to "data.ms"\n' "$restore" >&2
mv "$restore" data.ms
trap - EXIT
cleanup_fs
echo 'Restarting Meilisearch' >&2
kill 1

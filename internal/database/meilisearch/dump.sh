#!/usr/bin/env sh
set -euo pipefail

curl_api() {
  curl -sSfX "${2:-GET}" -H "Authorization: Bearer $MEILI_MASTER_KEY" "$API_HOST/$1"
}

get_dump_uid() {
  curl_api "tasks/$task_uid" | grep -Eo '"dumpUid":"[^,}]+' | cut -d: -f2 | cut -d\" -f2
}

echo 'Creating dump' >&2
if command -v meilitool &>/dev/null; then
  meilitool export-a-dump
  dump="dumps/$(ls -t1 dumps | head -n1)"
else
  task_uid="$(curl_api dumps POST | grep -Eo '"taskUid":[^,}]+' | cut -d: -f2)"
  echo 'Waiting for dump to complete' >&2
  while [ -z "${dump_uid:-}" ]; do
    dump_uid="$(get_dump_uid || true)"
    sleep 1
  done
  dump="dumps/$dump_uid.dump"
fi

echo Downloading dump >&2
cat "$dump"
echo 'Cleaning up' >&2
rm -f "$dump"

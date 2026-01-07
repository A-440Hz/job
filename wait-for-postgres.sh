#!/bin/sh
set -e

: "${PGTIMEOUT:=300}" # total seconds to wait

echo "Waiting up to ${PGTIMEOUT}s for Postgres at ${PGHOST}:${PGPORT}..."

elapsed=0
interval=2
while ! pg_isready -h "$PGHOST" -p "$PGPORT" -U "$PGUSER" -d "$PGDATABASE" >/dev/null 2>&1; do
  sleep $interval
  elapsed=$((elapsed + interval))
  echo "elapsed time: $elapsed seconds"
  if [ "$elapsed" -ge "$PGTIMEOUT" ]; then
    echo "Timed out waiting for Postgres after $elapsed seconds" >&2
    exit 1
  fi
done

echo "Postgres ready, starting app"
exec ./out -production=true
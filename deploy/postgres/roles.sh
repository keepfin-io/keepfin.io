#!/bin/sh
set -eu
: "${KEEPFIN_DB_PASSWORD:?Required}"
: "${KEEPFIN_MIGRATION_DB_PASSWORD:?Required}"
: "${AUTH_DB_PASSWORD:?Required}"
: "${AUTH_MIGRATION_DB_PASSWORD:?Required}"
for password in "$KEEPFIN_DB_PASSWORD" "$KEEPFIN_MIGRATION_DB_PASSWORD" "$AUTH_DB_PASSWORD" "$AUTH_MIGRATION_DB_PASSWORD"; do
  [ "${#password}" -ge 16 ] || { echo 'Database passwords must contain at least 16 characters' >&2; exit 1; }
  count=0
  for other in "$KEEPFIN_DB_PASSWORD" "$KEEPFIN_MIGRATION_DB_PASSWORD" "$AUTH_DB_PASSWORD" "$AUTH_MIGRATION_DB_PASSWORD"; do
    if [ "$password" = "$other" ]; then count=$((count + 1)); fi
  done
  [ "$count" -eq 1 ] && [ "$password" != "${PGPASSWORD:-}" ] || { echo 'Use different database passwords for every role' >&2; exit 1; }
done
exec psql -X --no-password -f /security/roles.sql

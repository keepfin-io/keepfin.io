#!/bin/sh
# Uses a disposable PostgreSQL container/volume. Never targets the application DB.
set -eu
workers=false
case "${1:-}" in
  --workers) workers=true ;;
  '') ;;
  *) echo 'Usage: sh scripts/test-db-security.sh [--workers]' >&2; exit 1 ;;
esac
repo=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
cd "$repo"
test_name="keepfin-security-test-$(date +%s)-$$"
test_dir=$(mktemp -d)
cleanup() {
  if [ "$workers" = true ]; then docker rm -fv "$test_name-search" >/dev/null 2>&1 || true; fi
  docker rm -fv "$test_name" >/dev/null 2>&1 || true
  rm -rf "$test_dir"
}
trap cleanup EXIT INT TERM
export GOCACHE=${GOCACHE:-/private/tmp/keepfin-controller-go-cache}
go build -o "$test_dir/keepnode" .
npm --prefix auth run build
docker run -d --name "$test_name" -e POSTGRES_PASSWORD=security-test-admin \
  -e POSTGRES_DB=keepfin_security_test -p 127.0.0.1::5432 \
  -v "$repo/deploy/postgres:/security:ro" postgres:18.6 >/dev/null
attempt=0
until docker exec "$test_name" pg_isready -U postgres -d keepfin_security_test >/dev/null 2>&1; do
  attempt=$((attempt + 1)); [ "$attempt" -lt 30 ] || exit 1
  sleep 1
done
port=$(docker port "$test_name" 5432/tcp | sed 's/.*://')
export DB_HOST=127.0.0.1 DB_PORT="$port" DB_NAME=keepfin_security_test DB_SSL_MODE=disable
export DB_USERNAME=postgres DB_PASSWORD=security-test-admin
# Better Auth intentionally relaxes origin checks under process NODE_ENV=test.
# Keep the real middleware active; individual test configs still use test mode.
export NODE_ENV=development APP_ORIGIN=http://localhost:8080 BETTER_AUTH_SECRET=security-test-secret-012345678901234567890123456789
export PGPASSWORD=security-test-admin DATABASE_URL="postgres://postgres@127.0.0.1:$port/keepfin_security_test"

# Reproduce the legacy owner layout, then upgrade it in place.
"$test_dir/keepnode" migrate >"$test_dir/go-migrate.log" 2>&1 || { cat "$test_dir/go-migrate.log"; exit 1; }
node auth/dist/migrate.js
docker exec "$test_name" psql -U postgres -d keepfin_security_test -v ON_ERROR_STOP=1 -c \
  "INSERT INTO users(id,email,user_status) VALUES ('10000000-0000-4000-8000-000000000001','legacy-security@example.test',1)" >/dev/null
docker exec "$test_name" psql -U postgres -d keepfin_security_test -v ON_ERROR_STOP=1 -c \
  "INSERT INTO accounts(id,user_id,account_name,currency,initial_balance,ledger_ready) VALUES ('20000000-0000-4000-8000-000000000001','10000000-0000-4000-8000-000000000001','Retained account','USD',775.50,true)" >/dev/null

roles() {
  docker exec -e PGUSER=postgres -e PGDATABASE="${1:-keepfin_security_test}" \
    -e KEEPFIN_DB_PASSWORD=security-test-runtime -e KEEPFIN_MIGRATION_DB_PASSWORD=security-test-migrator \
    -e AUTH_DB_PASSWORD=security-test-auth -e AUTH_MIGRATION_DB_PASSWORD=security-test-auth-migrator \
    "$test_name" sh /security/roles.sh >"$test_dir/roles.log" 2>&1 || { cat "$test_dir/roles.log"; exit 1; }
}
permissions() {
  docker exec "$test_name" psql -U postgres -d "${1:-keepfin_security_test}" -X -f /security/permissions.sql \
    >"$test_dir/permissions.log" 2>&1 || { cat "$test_dir/permissions.log"; exit 1; }
}
roles
DB_USERNAME=keepfin_migrator DB_PASSWORD=security-test-migrator "$test_dir/keepnode" migrate \
  >"$test_dir/go-upgrade.log" 2>&1 || { cat "$test_dir/go-upgrade.log"; exit 1; }
DATABASE_URL="postgres://auth_migrator@127.0.0.1:$port/keepfin_security_test" PGPASSWORD=security-test-auth-migrator node auth/dist/migrate.js
permissions
# A second deployment must not change existing IDs or undo restrictions.
roles
permissions
# Also verify first installation: role setup must work before any tables exist.
docker exec "$test_name" createdb -U postgres keepfin_security_fresh
roles keepfin_security_fresh
DB_NAME=keepfin_security_fresh DB_USERNAME=keepfin_migrator DB_PASSWORD=security-test-migrator "$test_dir/keepnode" migrate \
  >"$test_dir/go-fresh.log" 2>&1 || { cat "$test_dir/go-fresh.log"; exit 1; }
DATABASE_URL="postgres://auth_migrator@127.0.0.1:$port/keepfin_security_fresh" PGPASSWORD=security-test-auth-migrator node auth/dist/migrate.js
permissions keepfin_security_fresh
export AUTH_SECURITY_DATABASE_URL="postgres://postgres:security-test-admin@127.0.0.1:$port/keepfin_security_test"
export AUTH_TEST_DATABASE_URL="$AUTH_SECURITY_DATABASE_URL"
export SECURITY_RUNTIME_DATABASE_URL="postgres://keepfin_runtime:security-test-runtime@127.0.0.1:$port/keepfin_security_test"
if [ "$workers" = true ]; then
  export PORTFOLIO_TEST_DATABASE_URL="$AUTH_SECURITY_DATABASE_URL"
  export TYPESENSE_API_KEY=worker-test-key
  docker run -d --name "$test_name-search" --tmpfs /data -p 127.0.0.1::8108 \
    typesense/typesense:30.2 --data-dir=/data --api-key="$TYPESENSE_API_KEY" >/dev/null
  search_port=$(docker port "$test_name-search" 8108/tcp | sed 's/.*://')
  export HOLDINGS_SEARCH_TEST_URL="http://127.0.0.1:$search_port"
  attempt=0
  until curl -fsS "$HOLDINGS_SEARCH_TEST_URL/health" >/dev/null 2>&1; do
    attempt=$((attempt + 1)); [ "$attempt" -lt 30 ] || exit 1
    sleep 1
  done
fi
npm --prefix auth test
go test -race ./... -count=1
echo 'Database role isolation, identity audit, auth and Go integration checks passed.'

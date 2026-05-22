#!/usr/bin/env sh
set -eu

MIGRATIONS_DIR="${MIGRATIONS_DIR:-/migrations}"

echo "Waiting for PostgreSQL..."
until pg_isready -h "${PGHOST:-postgres}" -p "${PGPORT:-5432}" -U "${PGUSER}" -d "${PGDATABASE}" >/dev/null 2>&1; do
  sleep 1
done

psql -v ON_ERROR_STOP=1 -c "CREATE TABLE IF NOT EXISTS schema_migrations (version TEXT PRIMARY KEY, applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW());"

existing_users="$(psql -At -v ON_ERROR_STOP=1 -c "SELECT to_regclass('public.users') IS NOT NULL;")"
has_0001="$(psql -At -v ON_ERROR_STOP=1 -c "SELECT EXISTS (SELECT 1 FROM schema_migrations WHERE version = '0001_init');")"
if [ "$existing_users" = "t" ] && [ "$has_0001" != "t" ]; then
  psql -v ON_ERROR_STOP=1 -c "INSERT INTO schema_migrations (version) VALUES ('0001_init') ON CONFLICT DO NOTHING;"
fi

for migration in "$MIGRATIONS_DIR"/*.up.sql; do
  [ -e "$migration" ] || continue
  filename="$(basename "$migration")"
  version="${filename%.up.sql}"
  applied="$(psql -At -v ON_ERROR_STOP=1 -c "SELECT EXISTS (SELECT 1 FROM schema_migrations WHERE version = '$version');")"
  if [ "$applied" = "t" ]; then
    echo "Skipping applied migration $version"
    continue
  fi

  echo "Applying migration $version"
  psql -v ON_ERROR_STOP=1 -f "$migration"
  psql -v ON_ERROR_STOP=1 -c "INSERT INTO schema_migrations (version) VALUES ('$version');"
done

echo "Migrations complete."

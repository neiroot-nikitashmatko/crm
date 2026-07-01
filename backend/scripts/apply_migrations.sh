#!/usr/bin/env bash
set -euo pipefail

CONTAINER="${PROCLIENTS_PG_CONTAINER:-proclients-postgres}"
DB_USER="${PROCLIENTS_PG_USER:-postgres}"
DB_NAME="${PROCLIENTS_PG_DB:-proclients}"
MIGRATIONS_DIR="$(cd "$(dirname "$0")/../migrations" && pwd)"

if ! docker ps --format '{{.Names}}' | grep -qx "$CONTAINER"; then
  echo "Контейнер PostgreSQL «$CONTAINER» не запущен." >&2
  exit 1
fi

for migration in "$MIGRATIONS_DIR"/*.up.sql; do
  echo "→ $(basename "$migration")"
  docker exec -i "$CONTAINER" psql -U "$DB_USER" -d "$DB_NAME" < "$migration"
done

echo "Миграции применены."

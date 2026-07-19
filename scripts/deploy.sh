#!/usr/bin/env bash
# Деплой PROclients на сервер: pull → миграции → backend → frontend.
# Запуск на сервере:
#   cd /opt/proclients && git pull && ./scripts/deploy.sh
# Или с Mac после git push:
#   ssh root@138.16.184.152 'cd /opt/proclients && git pull && ./scripts/deploy.sh'
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

API_BIN="${PROCLIENTS_API_BIN:-/opt/proclients/proclients-api}"
ENV_FILE="${PROCLIENTS_ENV_FILE:-$ROOT/backend/.env}"

log() { printf '\n==> %s\n' "$*"; }
die() { printf 'ERROR: %s\n' "$*" >&2; exit 1; }

SKIP_PULL=0
SKIP_MIGRATE=0
SKIP_BACKEND=0
SKIP_FRONTEND=0

for arg in "$@"; do
  case "$arg" in
    --skip-pull) SKIP_PULL=1 ;;
    --skip-migrate) SKIP_MIGRATE=1 ;;
    --skip-backend) SKIP_BACKEND=1 ;;
    --skip-frontend) SKIP_FRONTEND=1 ;;
    --frontend-only) SKIP_MIGRATE=1; SKIP_BACKEND=1 ;;
    --backend-only) SKIP_FRONTEND=1 ;;
    -h|--help)
      cat <<'EOF'
Usage: ./scripts/deploy.sh [options]

  --skip-pull       не делать git pull
  --skip-migrate    не применять миграции
  --skip-backend    не собирать/рестартить API
  --skip-frontend   не собирать frontend / reload nginx
  --frontend-only   только frontend
  --backend-only    миграции + backend (без frontend)
EOF
      exit 0
      ;;
    *) die "неизвестный аргумент: $arg" ;;
  esac
done

if [[ "$SKIP_PULL" -eq 0 ]]; then
  log "git pull"
  git pull --ff-only
fi

run_migrations() {
  [[ -f "$ENV_FILE" ]] || die "нет файла $ENV_FILE"

  # shellcheck disable=SC1090
  set -a
  source "$ENV_FILE"
  set +a

  [[ -n "${DATABASE_URL:-}" ]] || die "в $ENV_FILE не задан DATABASE_URL"

  python3 - "$ROOT/backend/migrations" <<'PY'
import os, sys, subprocess
from pathlib import Path
from urllib.parse import urlparse, unquote

migrations_dir = Path(sys.argv[1])
url = os.environ["DATABASE_URL"]
u = urlparse(url)
if u.scheme not in ("postgres", "postgresql"):
    raise SystemExit(f"Неподдерживаемый DATABASE_URL: {u.scheme}")

db = (u.path or "/").lstrip("/")
if not db:
    raise SystemExit("В DATABASE_URL нет имени БД")

env = os.environ.copy()
env["PGPASSWORD"] = unquote(u.password or "")
base = [
    "psql",
    "-v", "ON_ERROR_STOP=1",
    "-h", u.hostname or "localhost",
    "-p", str(u.port or 5432),
    "-U", unquote(u.username or "postgres"),
    "-d", db,
]

def psql(sql: str) -> None:
    subprocess.run(base + ["-c", sql], check=True, env=env)

def psql_file(path: Path) -> None:
    subprocess.run(base + ["-f", str(path)], check=True, env=env)

def psql_scalar(sql: str) -> str:
    out = subprocess.check_output(base + ["-At", "-c", sql], env=env, text=True)
    return out.strip()

psql(
    """
    CREATE TABLE IF NOT EXISTS schema_migrations (
      filename TEXT PRIMARY KEY,
      applied_at TIMESTAMPTZ NOT NULL DEFAULT now()
    );
    """
)

count = int(psql_scalar("SELECT count(*) FROM schema_migrations;") or "0")
files = sorted(migrations_dir.glob("*.up.sql"))

def migration_num(name: str) -> int:
    prefix = name.split("_", 1)[0]
    return int(prefix) if prefix.isdigit() else -1

# Миграции, которые на проде уже были накатаны вручную до появления трекера
BOOTSTRAP_MAX = 21

if count == 0 and files:
    users_exists = psql_scalar("SELECT to_regclass('public.users');") == "users"
    if users_exists:
        print(
            f"schema_migrations пуста, БД уже есть — "
            f"помечаю миграции 001–{BOOTSTRAP_MAX:03d} как применённые."
        )
        for path in files:
            if 0 <= migration_num(path.name) <= BOOTSTRAP_MAX:
                name = path.name.replace("'", "''")
                psql(
                    f"INSERT INTO schema_migrations (filename) VALUES ('{name}') "
                    f"ON CONFLICT DO NOTHING;"
                )
    else:
        print("Пустая БД — применю все миграции с нуля.")

applied = set(
    filter(
        None,
        psql_scalar("SELECT filename FROM schema_migrations ORDER BY filename;").splitlines(),
    )
)
pending = [p for p in files if p.name not in applied]

if not pending:
    print("Новых миграций нет.")
    raise SystemExit(0)

for path in pending:
    print(f"→ {path.name}")
    psql_file(path)
    name = path.name.replace("'", "''")
    psql(f"INSERT INTO schema_migrations (filename) VALUES ('{name}');")

print(f"Применено миграций: {len(pending)}")
PY
}

if [[ "$SKIP_MIGRATE" -eq 0 ]]; then
  log "миграции"
  run_migrations
fi

if [[ "$SKIP_BACKEND" -eq 0 ]]; then
  log "сборка backend"
  (
    cd "$ROOT/backend"
    go build -o "$API_BIN" ./cmd/server
  )
  log "restart proclients-api"
  systemctl restart proclients-api
  systemctl --no-pager --lines=0 status proclients-api >/dev/null
  if curl -fsS http://127.0.0.1:8080/health >/dev/null; then
    echo "health: ok"
  else
    die "API не отвечает на /health"
  fi
fi

if [[ "$SKIP_FRONTEND" -eq 0 ]]; then
  log "сборка frontend"
  (
    cd "$ROOT/frontend"
    npm ci
    npm run build
  )
  log "reload nginx"
  nginx -t
  systemctl reload nginx
fi

log "деплой завершён"

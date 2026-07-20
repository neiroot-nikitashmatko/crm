#!/usr/bin/env bash
# Подписка webhook Авито на публичный URL (ngrok / Cloudflare Tunnel).
#
# Проще всего — только origin туннеля (секрет возьмётся из backend/.env):
#   ./backend/scripts/avito_subscribe_webhook.sh 'https://xxxx.trycloudflare.com'
#
# Или полный URL уже с ?secret=...
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
ENV_FILE="${PROCLIENTS_ENV_FILE:-$ROOT/backend/.env}"

# Парсим .env через Python — безопасно для секретов с $ & и т.п.
eval "$(
  ENV_FILE="$ENV_FILE" python3 - <<'PY'
import os, shlex
path = os.environ["ENV_FILE"]
if not os.path.isfile(path):
    raise SystemExit(0)
keys = ("AVITO_CLIENT_ID", "AVITO_CLIENT_SECRET", "AVITO_WEBHOOK_SECRET", "AVITO_WEBHOOK_PUBLIC_URL")
wanted = set(keys)
found = {}
with open(path, encoding="utf-8") as f:
    for raw in f:
        line = raw.strip()
        if not line or line.startswith("#") or "=" not in line:
            continue
        key, value = line.split("=", 1)
        key = key.strip()
        if key not in wanted:
            continue
        value = value.strip()
        if len(value) >= 2 and value[0] == value[-1] and value[0] in "'\"":
            value = value[1:-1]
        found[key] = value
for key in keys:
    if key in found:
        print(f"{key}={shlex.quote(found[key])}")
PY
)"

CLIENT_ID="${AVITO_CLIENT_ID:-}"
CLIENT_SECRET="${AVITO_CLIENT_SECRET:-}"
WEBHOOK_SECRET="${AVITO_WEBHOOK_SECRET:-}"
INPUT_URL="${1:-${AVITO_WEBHOOK_PUBLIC_URL:-}}"

if [[ -z "$CLIENT_ID" || -z "$CLIENT_SECRET" ]]; then
  echo "Нужны AVITO_CLIENT_ID и AVITO_CLIENT_SECRET (в backend/.env)" >&2
  exit 1
fi
if [[ -z "$INPUT_URL" ]]; then
  echo "Укажите URL туннеля:" >&2
  echo "  ./backend/scripts/avito_subscribe_webhook.sh 'https://xxxx.trycloudflare.com'" >&2
  exit 1
fi
if [[ -z "$WEBHOOK_SECRET" ]]; then
  echo "Нужен AVITO_WEBHOOK_SECRET в backend/.env" >&2
  exit 1
fi

PUBLIC_URL="$(
  INPUT_URL="$INPUT_URL" WEBHOOK_SECRET="$WEBHOOK_SECRET" python3 - <<'PY'
import os
from urllib.parse import urlsplit, urlunsplit, parse_qsl, urlencode

raw = os.environ["INPUT_URL"].strip()
secret = os.environ["WEBHOOK_SECRET"]
parts = urlsplit(raw)
path = parts.path.rstrip("/")
if not path or path == "":
    path = "/api/v1/integrations/avito/webhook"
elif not path.endswith("/api/v1/integrations/avito/webhook"):
    if path.endswith("/api/v1/integrations/avito/webhook/"):
        path = path.rstrip("/")
    else:
        path = path + "/api/v1/integrations/avito/webhook"

query = dict(parse_qsl(parts.query, keep_blank_values=True))
query["secret"] = secret
print(urlunsplit((parts.scheme, parts.netloc, path, urlencode(query), "")))
PY
)"

TOKEN="$(
  curl -sS -X POST 'https://api.avito.ru/token' \
    -H 'Content-Type: application/x-www-form-urlencoded' \
    --data-urlencode "grant_type=client_credentials" \
    --data-urlencode "client_id=${CLIENT_ID}" \
    --data-urlencode "client_secret=${CLIENT_SECRET}" \
  | python3 -c 'import sys,json; print(json.load(sys.stdin)["access_token"])'
)"

BODY="$(PUBLIC_URL="$PUBLIC_URL" python3 - <<'PY'
import json, os
print(json.dumps({"url": os.environ["PUBLIC_URL"]}, ensure_ascii=False))
PY
)"

echo "Подписываю webhook:"
echo "  $PUBLIC_URL"

curl -sS -X POST 'https://api.avito.ru/messenger/v3/webhook' \
  -H "Authorization: Bearer ${TOKEN}" \
  -H 'Content-Type: application/json' \
  -d "$BODY"
echo
echo
echo "Текущие подписки:"
curl -sS 'https://api.avito.ru/messenger/v1/subscriptions' \
  -H "Authorization: Bearer ${TOKEN}"
echo

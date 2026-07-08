#!/usr/bin/env bash
set -euo pipefail

# Renews all Beeline XSI-Events subscriptions.
# Reads BEELINE_API_TOKEN and BEELINE_WEBHOOK_SECRET from backend/.env by default.

ROOT_DIR="$(cd "$(dirname "$0")/../.." && pwd)"
ENV_FILE="${BEELINE_ENV_FILE:-$ROOT_DIR/backend/.env}"
LOG_FILE="${BEELINE_RENEW_LOG:-/var/log/proclients-beeline-renew.log}"

if [[ -f "$ENV_FILE" ]]; then
  set -a
  # shellcheck disable=SC1090
  source "$ENV_FILE"
  set +a
fi

mkdir -p "$(dirname "$LOG_FILE")"

{
  echo "=== $(date -Is) renew start ==="
  "$ROOT_DIR/backend/scripts/beeline_subscribe_all.sh"
  echo "=== $(date -Is) renew done ==="
} >>"$LOG_FILE" 2>&1

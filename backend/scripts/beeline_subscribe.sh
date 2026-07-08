#!/usr/bin/env bash
set -euo pipefail

# Usage:
#   BEELINE_API_TOKEN=... BEELINE_PATTERN=200 BEELINE_CALLBACK_URL=https://crm.neiroot.ru/api/v1/integrations/beeline/xsi-events ./backend/scripts/beeline_subscribe.sh
#
# Optional:
#   BEELINE_EXPIRES=3600
#   BEELINE_SUBSCRIPTION_TYPE=BASIC_CALL

TOKEN="${BEELINE_API_TOKEN:-}"
PATTERN="${BEELINE_PATTERN:-}"
CALLBACK_URL="${BEELINE_CALLBACK_URL:-}"
EXPIRES="${BEELINE_EXPIRES:-3600}"
SUB_TYPE="${BEELINE_SUBSCRIPTION_TYPE:-BASIC_CALL}"

if [[ -z "$TOKEN" ]]; then
  echo "BEELINE_API_TOKEN is required" >&2
  exit 1
fi
if [[ -z "$CALLBACK_URL" ]]; then
  echo "BEELINE_CALLBACK_URL is required" >&2
  exit 1
fi

payload=$(cat <<EOF
{
  "pattern": "${PATTERN}",
  "expires": ${EXPIRES},
  "subscriptionType": "${SUB_TYPE}",
  "url": "${CALLBACK_URL}"
}
EOF
)

curl -sS -X PUT \
  -H "X-MPBX-API-AUTH-TOKEN: ${TOKEN}" \
  -H "Content-Type: application/json" \
  -d "${payload}" \
  "https://cloudpbx.beeline.ru/apis/portal/subscription"

echo

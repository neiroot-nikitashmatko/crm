#!/usr/bin/env bash
set -euo pipefail

# Subscribes all multi-channel Beeline numbers with traffic source per callback URL.
#
# Required env:
#   BEELINE_API_TOKEN
#   BEELINE_WEBHOOK_SECRET
#
# Optional env:
#   BEELINE_CALLBACK_BASE=https://crm.neiroot.ru/api/v1/integrations/beeline/xsi-events
#   BEELINE_NUMBER_DOMAIN=@rnd.so.ims.mnc099.mcc250.3gppnetwork.org
#   BEELINE_EXPIRES=3600
#   BEELINE_SUBSCRIPTION_TYPE=BASIC_CALL

TOKEN="${BEELINE_API_TOKEN:-}"
SECRET="${BEELINE_WEBHOOK_SECRET:-}"
BASE_URL="${BEELINE_CALLBACK_BASE:-https://crm.neiroot.ru/api/v1/integrations/beeline/xsi-events}"
DOMAIN="${BEELINE_NUMBER_DOMAIN:-@rnd.so.ims.mnc099.mcc250.3gppnetwork.org}"
EXPIRES="${BEELINE_EXPIRES:-3600}"
SUB_TYPE="${BEELINE_SUBSCRIPTION_TYPE:-BASIC_CALL}"

if [[ -z "$TOKEN" ]]; then
  echo "BEELINE_API_TOKEN is required" >&2
  exit 1
fi
if [[ -z "$SECRET" ]]; then
  echo "BEELINE_WEBHOOK_SECRET is required" >&2
  exit 1
fi

urlencode() {
  python3 -c 'import urllib.parse, sys; print(urllib.parse.quote(sys.argv[1], safe=""))' "$1"
}

subscribe_number() {
  local phone="$1"
  local source="$2"
  local encoded_source pattern callback_url payload

  encoded_source="$(urlencode "$source")"
  pattern="${phone}${DOMAIN}"
  callback_url="${BASE_URL}/${SECRET}?trafficSource=${encoded_source}"

  payload=$(cat <<EOF
{
  "pattern": "${pattern}",
  "expires": ${EXPIRES},
  "subscriptionType": "${SUB_TYPE}",
  "url": "${callback_url}"
}
EOF
)

  echo "==> ${phone} (${source})"
  curl -sS -X PUT \
    -H "X-MPBX-API-AUTH-TOKEN: ${TOKEN}" \
    -H "Content-Type: application/json" \
    -d "${payload}" \
    "https://cloudpbx.beeline.ru/apis/portal/subscription"
  echo
}

subscribe_number "9613001616" "Знал о производстве"
subscribe_number "9613015050" "Знал о производстве"
subscribe_number "9662066959" "Визитка(авточехлы)"
subscribe_number "9613195219" "Авито (Автоатрибут)"
subscribe_number "9064545834" "Авито (AutoFactory)"
subscribe_number "9064545866" "Авито (Автоателье)"
subscribe_number "9034306767" "Яндекс карты"
subscribe_number "9034363336" "Instagram"
subscribe_number "9613011458" "Вконтакте"
subscribe_number "9613011460" "2gis"

echo "Done."

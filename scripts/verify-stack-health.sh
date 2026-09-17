#!/usr/bin/env bash
# Verify API + UI respond before Playwright / live probes.
set -euo pipefail

API_HEALTH_URL="${API_HEALTH_URL:-http://localhost:8080/actuator/health}"
UI_HEALTH_URL="${UI_HEALTH_URL:-http://localhost:3000}"
MAX_WAIT_SEC="${MAX_WAIT_SEC:-120}"
INTERVAL_SEC="${INTERVAL_SEC:-3}"

wait_url() {
  local url="$1"
  local label="$2"
  local elapsed=0
  echo "Waiting for $label at $url ..."
  until curl -sf "$url" >/dev/null 2>&1; do
    sleep "$INTERVAL_SEC"
    elapsed=$((elapsed + INTERVAL_SEC))
    if [[ "$elapsed" -ge "$MAX_WAIT_SEC" ]]; then
      echo "FAIL: $label not healthy after ${MAX_WAIT_SEC}s ($url)" >&2
      return 1
    fi
  done
  echo "OK: $label healthy"
}

wait_url "$API_HEALTH_URL" "API"
wait_url "$UI_HEALTH_URL" "UI"
echo "Stack verified — proceed with /run-tests (JUnit → Vitest → Playwright)."

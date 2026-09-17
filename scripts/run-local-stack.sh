#!/usr/bin/env bash
# Start API + UI for local test/e2e runs. Requires modules built per config.yaml.
set -euo pipefail

API_ROOT="${API_MODULE_ROOT:-modules/api}"
UI_ROOT="${UI_MODULE_ROOT:-modules/ui}"
API_PORT="${API_PORT:-8080}"
UI_PORT="${UI_PORT:-3000}"
PID_FILE="${PID_FILE:-artifacts/local-stack.pids}"

mkdir -p artifacts
: > "$PID_FILE"

if [[ ! -d "$API_ROOT" || ! -d "$UI_ROOT" ]]; then
  echo "ERROR: Set API_MODULE_ROOT and UI_MODULE_ROOT to implemented module paths." >&2
  exit 1
fi

echo "Starting API in $API_ROOT on port $API_PORT ..."
(
  cd "$API_ROOT"
  if [[ -f pom.xml ]]; then
    SPRING_PROFILES_ACTIVE="${SPRING_PROFILES_ACTIVE:-local}" \
      mvn -q spring-boot:run -Dspring-boot.run.arguments="--server.port=$API_PORT" &
  else
    echo "ERROR: No pom.xml under $API_ROOT" >&2
    exit 1
  fi
  echo $! >> "$PID_FILE"
)

echo "Starting UI in $UI_ROOT on port $UI_PORT ..."
(
  cd "$UI_ROOT"
  if [[ -f package.json ]]; then
    PORT="$UI_PORT" npm run dev &
  else
    echo "ERROR: No package.json under $UI_ROOT" >&2
    exit 1
  fi
  echo $! >> "$PID_FILE"
)

echo "Stack starting. Run scripts/verify-stack-health.sh before /run-tests E2E."

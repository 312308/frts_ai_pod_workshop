#!/usr/bin/env bash
set -euo pipefail
PID_FILE="${PID_FILE:-artifacts/local-stack.pids}"
if [[ ! -f "$PID_FILE" ]]; then
  echo "No PID file at $PID_FILE"
  exit 0
fi
while read -r pid; do
  [[ -n "$pid" ]] && kill "$pid" 2>/dev/null || true
done < "$PID_FILE"
rm -f "$PID_FILE"
echo "Local stack stopped."

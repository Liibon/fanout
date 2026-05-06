#!/usr/bin/env bash
# Wait until root is accepting connections.
set -euo pipefail
TIMEOUT="${TIMEOUT:-120}"
START=$(date +%s)
echo "Waiting for root on :50051..."
while ! nc -z localhost 50051 2>/dev/null; do
  if (( $(date +%s) - START > TIMEOUT )); then
    echo "ERROR: root not ready after ${TIMEOUT}s" >&2; exit 1
  fi
  sleep 2
done
echo "root is ready."

#!/usr/bin/env bash
# Run a QPS sweep and print a CSV latency table.
set -euo pipefail

QPS_VALUES="${QPS_VALUES:-50 100 200 400}"
WARMUP="${WARMUP:-300}"
MEASURE="${MEASURE:-2000}"

echo "qps,p50,p95,p99,max"
for qps in $QPS_VALUES; do
  out=$(docker compose run --rm loadgen \
    -qps=$qps -warmup=$WARMUP -measure=$MEASURE 2>/dev/null)
  p50=$(echo "$out"  | awk '/^p50:/{print $2}')
  p95=$(echo "$out"  | awk '/^p95:/{print $2}')
  p99=$(echo "$out"  | awk '/^p99:/{print $2}')
  max=$(echo "$out"  | awk '/^max:/{print $2}')
  echo "$qps,$p50,$p95,$p99,$max"
done

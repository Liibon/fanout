#!/usr/bin/env bash
set -euo pipefail
QPS=${1:-200}; FAN_OUT=${2:-16}
info() { echo "[demo] $*"; }
SYNTHETIC=true FAN_OUT=$FAN_OUT docker compose up -d \
    otel-collector jaeger prometheus dataset-gen \
    leaf-0 leaf-1 leaf-2 leaf-3 leaf-4 leaf-5 leaf-6 leaf-7 \
    leaf-8 leaf-9 leaf-10 leaf-11 leaf-12 leaf-13 leaf-14 leaf-15 root
info "Waiting for root..."
timeout 60 bash -c 'until docker compose logs root 2>&1 | grep -q "listening"; do sleep 2; done' || true
sleep 5
info "=== Phase 1: Baseline ==="
docker compose run --rm loadgen -root=root:50051 -qps="$QPS" -warmup=300 -measure=2000

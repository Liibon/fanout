#!/usr/bin/env bash
# demo-incast.sh - shows the incast tail-latency cliff in fanout.
#
# Runs two back-to-back load tests:
#   1. Baseline: synthetic leaves with low jitter  -> healthy p99
#   2. Incast:   synthetic leaves with high jitter -> p99 cliff visible in Jaeger
#
# Requires: docker compose v2
# Usage:    ./demo-incast.sh [qps] [fan-out]

set -euo pipefail

QPS=${1:-200}
FAN_OUT=${2:-16}
WARMUP=300
MEASURE=2000

JAEGER_UI="http://localhost:16686"
PROMETHEUS_UI="http://localhost:9090"

info()  { echo "[demo] $*"; }
warn()  { echo "[demo] $*"; }
fatal() { echo "[demo] $*" >&2; exit 1; }

# preflight
command -v docker >/dev/null 2>&1 || fatal "docker not found"
docker compose version >/dev/null 2>&1 || fatal "docker compose v2 not found"

# ensure stack is up with synthetic leaves
info "Starting stack in synthetic mode (SYNTHETIC=true FAN_OUT=$FAN_OUT)..."
SYNTHETIC=true FAN_OUT=$FAN_OUT docker compose up -d \
    otel-collector jaeger prometheus dataset-gen \
    leaf-0 leaf-1 leaf-2 leaf-3 leaf-4 leaf-5 leaf-6 leaf-7 \
    leaf-8 leaf-9 leaf-10 leaf-11 leaf-12 leaf-13 leaf-14 leaf-15 \
    root

info "Waiting for root to become ready..."
timeout 60 bash -c '
    until docker compose logs root 2>&1 | grep -q "listening"; do
        sleep 2
    done
' || true
sleep 5

# phase 1: baseline
info "=== Phase 1: Baseline (low jitter) ==="
BASELINE_LOG=$(mktemp)
docker compose run --rm \
    -e SYNTHETIC_MU=2.0 \
    -e SYNTHETIC_SIGMA=0.3 \
    -e SYNTHETIC_HEAVY_PCT=0.001 \
    loadgen \
    -root=root:50051 \
    -qps="$QPS" \
    -warmup="$WARMUP" \
    -measure="$MEASURE" \
    -dim=128 \
    -top-k=10 \
    2>&1 | tee "$BASELINE_LOG" || true

echo ""
warn "Baseline results:"
grep -E "^(p50|p90|p99|qps_achieved|samples)" "$BASELINE_LOG" || true
echo ""

# phase 2: incast
info "=== Phase 2: Incast (high jitter, straggler leaves) ==="
warn "Restarting leaves with aggressive heavy-tail distribution..."

SYNTHETIC=true \
SYNTHETIC_MU=2.5 \
SYNTHETIC_SIGMA=0.8 \
SYNTHETIC_HEAVY_PCT=0.10 \
SYNTHETIC_HEAVY_MU=6.5 \
SYNTHETIC_HEAVY_SIGMA=0.4 \
FAN_OUT=$FAN_OUT \
docker compose up -d --force-recreate \
    leaf-0 leaf-1 leaf-2 leaf-3 leaf-4 leaf-5 leaf-6 leaf-7 \
    leaf-8 leaf-9 leaf-10 leaf-11 leaf-12 leaf-13 leaf-14 leaf-15

sleep 5

INCAST_LOG=$(mktemp)
docker compose run --rm loadgen \
    -root=root:50051 \
    -qps="$QPS" \
    -warmup="$WARMUP" \
    -measure="$MEASURE" \
    -dim=128 \
    -top-k=10 \
    2>&1 | tee "$INCAST_LOG" || true

echo ""
warn "Incast results:"
grep -E "^(p50|p90|p99|qps_achieved|samples)" "$INCAST_LOG" || true
echo ""

# summary
echo ""
info "Baseline p99:"
grep "^p99:" "$BASELINE_LOG" || true
info "Incast p99:"
grep "^p99:" "$INCAST_LOG" || true
echo ""
info "Jaeger:     $JAEGER_UI  (service: usuite-root)"
info "Prometheus: $PROMETHEUS_UI"
echo ""
info "In Jaeger: look for traces where all 16 leaf spans start simultaneously."
info "The slowest leaf determines root p99 -- that is the incast effect."
echo ""
info "To reproduce:"
info "  FAN_OUT=$FAN_OUT QPS=$QPS SYNTHETIC=true SYNTHETIC_HEAVY_PCT=0.10 \\"
info "  docker compose run loadgen ..."

rm -f "$BASELINE_LOG" "$INCAST_LOG"

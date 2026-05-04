# Changelog

## [Unreleased]

### Added
- Grafana dashboard for root latency and leaf timeout visualization
- Prometheus alert rules for p99 and timeout thresholds
- `.env.example` documenting all configuration variables
- `scripts/bench.sh` QPS sweep, `scripts/teardown.sh`
- golangci-lint CI integration
- loadgen `-json` flag for machine-readable output

## v1.0.0 — 2026-05-15

### Added
- 16-shard FAISS-backed HDSearch scatter-gather service
- Root fan-out with configurable top-K merge using min-heap
- Leaf CGo FAISS bindings (IndexFlatL2, C API)
- Reproducible 1 M-vector dataset generator (seed=42, float32, dim=128)
- OTel distributed tracing via Jaeger
- Prometheus metrics (root request duration, leaf search duration, timeouts)
- Synthetic leaf mode for incast simulation (lognormal + heavy-tail)
- Hedged requests and per-leaf retry
- Open-loop loadgen with Poisson arrivals, p50/p90/p95/p99 output
- GitHub Actions CI (build, vet, race-enabled unit tests)
- Docker Compose stack: 16 leaves, root, otel-collector, Jaeger, Prometheus

### Fixed
- `libfaiss_c.so` not installed by cmake (manual copy from build tree)
- `libgomp.so.1` missing from runtime image (`libgomp1` apt package)
- Dataset-gen container write permission on Docker named volume

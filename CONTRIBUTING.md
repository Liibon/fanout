# Contributing to fanout

Thanks for your interest.  This document explains how to contribute effectively.

---

## Ground rules

1. **Results before aesthetics.** PRs that improve measurement correctness or
   reproducibility are highest priority.  Code style fixes are welcome but
   should be separate from functional changes.

2. **Methodology changes bump major version.**  If your PR changes what a
   latency number means (new percentile algorithm, different warmup logic,
   coordinated-omission handling), that is a major-version bump.  Discuss in
   an issue first.

3. **One concern per PR.**  Don't bundle a new feature with a bug fix.  It
   makes review hard and makes bisection harder.

4. **Tests for observable behavior.**  Unit tests are fine; prefer integration
   tests that actually start a subprocess over mock-heavy tests.

---

## Development setup

```bash
# prerequisites: Go 1.22+, protoc, Docker 24+
git clone https://github.com/liibon/fanout
cd fanout
make generate   # generates proto code into gen/
go build ./...  # builds root, loadgen, dataset (no FAISS needed)
# leaf requires CGO + FAISS; use Docker:
docker compose build leaf-0
```

## Running the smoke test locally

```bash
SYNTHETIC=true docker compose up -d
docker compose run loadgen -qps=50 -warmup=100 -measure=500
docker compose down -v
```

---

## PR checklist

- [ ] `go vet ./...` passes
- [ ] `make generate && go build ./...` passes
- [ ] Smoke test passes locally
- [ ] If changing load generator methodology: `METHODOLOGY.md` updated
- [ ] If changing dataset format: `METHODOLOGY.md` SHA-256 note updated
- [ ] If changing config knobs: `compose.yml` defaults and `README.md` updated

---

## Reporting issues

Use GitHub Issues.  For latency anomalies, include:
- Host CPU model and core count (`lscpu | head -20`)
- Kernel version (`uname -r`)
- Docker version (`docker version`)
- Full loadgen output (including the "Inputs" header)
- Compose logs (`docker compose logs --tail=50`)

---

## What we will not merge

- Kubernetes/Helm support (out of scope for v1)
- Additional fanout workloads beyond HDSearch (separate repo)
- Custom Grafana dashboards (Prometheus + Jaeger UI are intentionally the interface)
- Windows support
- Non-gRPC transports

---

## Triage cadence

Issues are triaged within **7 days**.  PRs get an initial review within
**14 days**.  If you haven't heard back, ping in the issue thread.

---

## Attribution

If you publish results produced by fanout, please cite:

> Sriraman, A., Daglis, A., Wenisch, T. F., & Gutierrez, J. (2018).  
> uSuite: A Benchmark Suite for Microservices. *IISWC 2018*.  
> https://doi.org/10.1109/IISWC.2018.8573523

and link to this repository.

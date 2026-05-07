# fanout / HDSearch

```bash
git clone https://github.com/liibon/fanout && cd fanout
docker compose up -d
./demo-incast.sh
```

A reproducible, containerised re-implementation of the **HDSearch** workload
from Sriraman et al., *uSuite: A Benchmark Suite for Microservices*
(IISWC 2018), updated for modern infrastructure.

---

## What is HDSearch?

HDSearch is a high-dimensional nearest-neighbour search service with a
**scatter-gather fan-out** topology:

```
         +---------+
 query ->   root   |  <- gRPC server, accepts queries
         +----+----+
   fan-out    |  (parallel RPCs to N leaves)
     +--------+--------+
     v        v        v
  leaf-0   leaf-1  ... leaf-15
  FAISS    FAISS       FAISS
  shard-0  shard-1     shard-15
     |        |           |
     +--------+-----------+
              |  aggregate top-K
              v
           response
```

Each leaf holds a shard of a 1M-vector, 128-dim corpus and answers ANN
queries using a FAISS flat-L2 index. The root fans the query out to all N
leaves simultaneously, waits (with a configurable per-leaf deadline), and
merges the top-K results.

This topology is the **canonical incast pattern**: p99 latency is determined
by the *slowest* leaf, not the average. As fan-out N grows or leaf variance
increases, tail latency degrades super-linearly.

---

## Quickstart

**Requirements:** Docker >= 24, `docker compose` v2, ~8 GB RAM, ~4 GB disk.

```bash
# 1. Clone
git clone https://github.com/liibon/fanout
cd fanout

# 2. Start the stack (builds images, generates dataset, starts all services)
docker compose up -d

# 3. Watch the incast cliff
./demo-incast.sh
```

The demo runs two load tests and prints a comparison table, then points you at
Jaeger (`http://localhost:16686`) and Prometheus (`http://localhost:9090`).

---

## Architecture

| Service | Language | Role |
|---------|----------|------|
| `root`  | Go | gRPC server; fans queries out to N leaves; merges top-K |
| `leaf`  | Go + FAISS (CGo) | gRPC server; flat-L2 ANN on a dataset shard |
| `loadgen` | Go | Open-loop Poisson load generator |
| `dataset-gen` | Go | Generates reproducible 1M-vector corpus (seed=42) |
| `otel-collector` | upstream | OTLP to Jaeger + Prometheus |
| `jaeger` | upstream | Distributed trace UI |
| `prometheus` | upstream | Metrics scraping |

---

## Configuration knobs

All config is via environment variables (or `compose.yml` overrides):

| Variable | Default | Description |
|----------|---------|-------------|
| `FAN_OUT` | 16 | Number of leaves the root fans out to |
| `TOP_K` | 10 | Top-K results returned per query |
| `PER_LEAF_TIMEOUT_MS` | 100 | Per-leaf RPC deadline (ms) |
| `HEDGING` | false | Enable hedged requests (resend to backup leaf after HedgingDelay) |
| `HEDGING_DELAY_MS` | 20 | Hedging delay threshold (ms) |
| `RETRY` | false | Enable retry on leaf failure |
| `MAX_RETRIES` | 1 | Maximum retries per leaf |
| `QPS` | 100 | Load generator target QPS |
| `SYNTHETIC` | false | Swap FAISS for sleep-with-jitter on all leaves |
| `SYNTHETIC_MU` | 2.5 | Log-mean of leaf latency distribution (ln ms) |
| `SYNTHETIC_SIGMA` | 0.6 | Log-std of leaf latency distribution |
| `SYNTHETIC_HEAVY_PCT` | 0.01 | Fraction of requests drawing from the heavy-tail distribution |
| `SYNTHETIC_HEAVY_MU` | 5.0 | Log-mean of the heavy-tail distribution (ln ms) |
| `SYNTHETIC_HEAVY_SIGMA` | 0.3 | Log-std of the heavy-tail distribution |
| `NUM_VECTORS` | 1000000 | Corpus size |
| `DIM` | 128 | Vector dimension |

Example: run with hedging enabled at 500 QPS:
```bash
FAN_OUT=8 HEDGING=true HEDGING_DELAY_MS=30 QPS=500 docker compose up -d
```

---

## Observability

### Jaeger (traces)
Open `http://localhost:16686`, select service **usuite-root**, and search.

Each root span has 16 child spans (one per leaf). The incast effect is
visible as 16 simultaneous spans, with the slowest determining the root
response time.

### Prometheus (metrics)
Open `http://localhost:9090`. Key metrics:

| Metric | Description |
|--------|-------------|
| `usuite_root_request_duration_seconds` | End-to-end latency histogram |
| `usuite_root_leaf_timeouts_total` | Per-leaf RPC timeouts |
| `usuite_root_leaf_errors_total` | Total leaf errors |
| `usuite_leaf_search_duration_seconds` | Per-leaf search latency |

Useful Prometheus query for p99:
```promql
histogram_quantile(0.99, rate(usuite_root_request_duration_seconds_bucket[1m]))
```

---

## What changed from 2018

This is **not** a port of the original fanout source. It is an independent
re-implementation of the HDSearch topology with deliberate modernisation.
Numbers from this implementation **should not** be compared to the IISWC 2018
paper.

| Dimension | Original fanout (2018) | fanout |
|-----------|----------------------|-------------|
| ANN library | MLPACK | FAISS `IndexFlatL2` |
| Distance metric | Depends on MLPACK config | L2 (Euclidean squared), always exact |
| Search type | Approximate (k-d tree) | Exact brute-force |
| Transport | Thrift | gRPC / protobuf |
| Observability | None | OpenTelemetry to Jaeger + Prometheus |
| Dataset | Proprietary | Generated, SHA-256 pinned, seed-42 reproducible |
| Load generation | Custom closed-loop | Open-loop Poisson (coordinated omission handled) |
| Language | C++ | Go (CGo for FAISS) |

**Why FAISS instead of MLPACK?** MLPACK's approximate k-d tree search has a
very different memory access pattern and branch structure compared to FAISS's
brute-force flat index. FAISS `IndexFlatL2` is the correct analog of a
"worst case" leaf that actually computes all distances. It maximises
memory bandwidth pressure, which is the bottleneck the incast experiment is
designed to amplify. Using an HNSW or IVF index would reduce per-leaf
latency and variance, making the incast effect harder to demonstrate.

**Why exact search?** The benchmark is about *fan-out latency*, not *recall
quality*. Exact search gives deterministic, reproducible latency distributions
and makes the dataset content irrelevant to the result.

---

## Reproducibility

- All Docker images are pinned by tag in `compose.yml`. Before publishing
  results, pin by digest: `docker inspect --format='{{index .RepoDigests 0}}'
  <image>`.
- The dataset is generated with a fixed seed (42) and its SHA-256 is
  written to `/data/dataset.bin.sha256` on first run.
- FAISS index parameters are hardcoded (`IndexFlatL2`, no training, no
  quantisation). See `METHODOLOGY.md` for the full parameter table.
- The load generator version is embedded in its binary via `go version`.

See `METHODOLOGY.md` for the complete measurement methodology, including
epsilon bounds for cross-hardware comparison.

---

## Citation

If you use fanout in published work, please cite the original paper:

```bibtex
@inproceedings{sriraman2018usuite,
  author    = {Sriraman, Akshitha and Daglis, Alexandros and Wenisch, Thomas F. and Gutierrez, Juan},
  title     = {{\textmu}Suite: A Benchmark Suite for Microservices},
  booktitle = {2018 IEEE International Symposium on Workload Characterization (IISWC)},
  year      = {2018},
  doi       = {10.1109/IISWC.2018.8573523}
}
```

and link to this repository.

---

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md).

---

## License

MIT. See [LICENSE](LICENSE).
Portions of the architecture are derived from the fanout paper (Apache 2.0).
See `LICENSE` for full attribution.

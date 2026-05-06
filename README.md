# fanout / HDSearch

A reproducible HDSearch scatter-gather benchmark (uSuite 2018 re-implementation).

```
         +---------+
 query ->   root   |
         +----+----+
              |  fan-out
     leaf-0 ... leaf-15
       FAISS     FAISS
              |  merge
           response
```

## Quickstart

```bash
git clone https://github.com/liibon/fanout && cd fanout
docker compose up -d
./demo-incast.sh
```

## Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `FAN_OUT` | 16 | Number of leaves |
| `TOP_K` | 10 | Top-K results |
| `PER_LEAF_TIMEOUT_MS` | 100 | Per-leaf RPC deadline (ms) |
| `HEDGING` | false | Hedged requests |
| `RETRY` | false | Retry on leaf failure |
| `QPS` | 100 | Load generator QPS |
| `SYNTHETIC` | false | Sleep instead of FAISS |
| `SYNTHETIC_HEAVY_PCT` | 0.01 | Straggler fraction |

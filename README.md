# fanout / HDSearch

A reproducible HDSearch scatter-gather benchmark.

```
 query -> root -> leaf-0 ... leaf-15 -> merge -> response
```

p99 determined by slowest leaf — the incast effect.

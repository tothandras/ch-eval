# Methodology (draft, Phase 1)

This page is the contract between the measurements and `REPORT.md`. It is
finalized in Phase 7, but the rules below are already binding.

## Principles

1. **Measured or labelled.** Every number in the report comes from a file under
   `results/`. Anything not measured is labelled "assessed, not measured" and
   cites a primary source.
2. **One variable at a time.** Comparisons hold the following fixed and change
   only the dimension under test:
   - workload, scenario and seed;
   - row count;
   - query parameters;
   - client location and client binary (`clickhouse-benchmark` 26.3.33.24);
   - shape and settings.

   The two dimensions under test are the target, or the scenario on one
   target.
3. **Parity before performance.** For each workload and scenario, value digests
   must match across targets over an identical seeded window before any
   latency or cost comparison is reported. A mismatch blocks the comparison and
   is investigated.
4. **Local is not performance.** `local-single`, `local-cluster` and `kind`
   results are used only for parity, correctness and smoke checks.
5. **Reproduce surprises.** A result that contradicts expectations or the
   existing repo's recommendation is reproduced, and the owner is asked,
   before it is reported.

## Perf runs

- Fixed seed (`--seed 42`) and pinned `--time-end`.
- 10M rows for OpenMeter perf and 100M for COGS. OTel sizes are fixed per
  workload in Phase 4/5.
- Warmup: one untimed pass per query (see Phase 2 task), then `--iterations 10`
  timed iterations through `clickhouse-benchmark`, at `--repeat ≥3` on
  separate service warm-ups.
- Variance is reported as the spread of per-repeat p50s.
- CPU: `OSCPUVirtualTimeMicroseconds` from `query_log`, read across replicas.
  On a multi-node target the per-iteration CPU is the sum over every node's
  `query_log` rows sharing the iteration's `initial_query_id`, read through
  `clusterAllReplicas(<cluster>, system.query_log)`. The `log_comment` that
  tags an iteration travels with the query to remote shards. Results record
  the source as `os_cpu_cluster` (or `os_cpu` on a single node).
- Seeding: every insert sets `async_insert` explicitly (0 unless
  `--async-insert`), because 26.2+ servers default it to 1 and a server
  default must not decide the insert path on one target only. On a sharded
  target the seed goes through the Distributed table with
  `distributed_foreground_insert=1`, so rows are on their shards before the
  timed queries start. On a replicated target the harness runs
  `SYSTEM SYNC REPLICA ON CLUSTER` before measuring.
- Schema: each run rolls the scenario's migrations down and up again, so
  the down path is exercised on every fresh run.
  `bench migrate verify` checks every replica: each table exists, the engine
  family matches the dialect, and the `CREATE` statement is identical.
- Cold reads: `enable_filesystem_cache=0` where the target has a filesystem
  cache, plus explicit cache drops where the target allows them (feature-matrix
  ST-08). On self-managed EBS, the OS page cache is dropped on the node. Cloud
  cannot do this, which is a recorded asymmetry.
- Settings: every run records changed session, MergeTree and server settings.
  Runs with different settings are not compared.

## Parity gate

- `bench compare --scenario S --targets a,b[,c]` diffs the value digests
  of every query. Before diffing it requires identical seeded inputs:
  generator, rows, RNG seed, namespaces and `time_end`. A mismatch, or a
  differing digest, exits non-zero, and no performance statement is made for
  that scenario until it is resolved.
- When any side is a local tier, or the sides are on different tiers, the
  output is labelled "smoke, not performance".
- **Sharding caveat.** Sharding is by `cityHash64(namespace)`. With the
  default single namespace every row lands on one shard, so parity on a
  sharded target also needs a multi-namespace run (`--namespaces 8`).

## Cross-target verdict rules (to be finalized)

- **Latency and CPU.** Per workload and scenario, compute the median across
  queries of the per-query p50 ratio between targets. A difference is
  *material* only when it exceeds the larger of 10% and the observed
  run-to-run spread.
- **Cost.** Compare at matched utilization using billed-shape costs. Report
  the utilization and ops-hours/month at which the cheaper target flips.
- **Capability.** Any feature-matrix row tagged OpenMeter-relevant whose
  probed status differs between targets is a design callout, independent of
  performance.
- **Recommendation and flip conditions.** Each workload gets one of `cloud`,
  `altinity` or "no material difference", together with the computed
  conditions that flip it.

## Replica consistency

See `openspec/changes/cross-target-evaluation/design.md` D7. It is reported as
a stale-read rate, lag percentiles, monotonicity violations per 10k reads, and
the latency overhead of each consistency policy.

## Known asymmetries (recorded, not corrected)

- Cloud's hardware generation and noisy neighbours are unobservable.
- Cloud versions (26.4/26.6) are ahead of Altinity Stable (26.3). The proposed
  `upstream-cloudver` control isolates this.
- Cloud bills compressed egress. The harness's egress estimate is an
  uncompressed upper bound, as in the source harness.

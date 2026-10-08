# AI Cost Management on ClickHouse Cloud vs sharded Altinity `prod-us`

> **Internal. Remove before this repository is made public.** Brief: [`BRIEF.md`](BRIEF.md).
> Evidence: [`results/`](results/) and the scripts in [`scripts/`](scripts/). How to reproduce: [Evidence index](#evidence-index).

**How this was evaluated.**
- Nothing ran against the team's clusters.
- **Verified** means one of two things:
  - I ran it locally against real ClickHouse builds and the result file is cited;
  - or a primary source states it, and the source is cited.
- **Builds used locally:**
  - `altinity/clickhouse-server:26.3.33.10001.altinitystable` (the target `prod-us` build);
  - `26.3.16.10001.altinitystable` (`prod-us` today);
  - upstream `26.3.33.24`, `26.4.5.143` (the closest to Cloud dev's reported 26.4), `26.7.4.58` and `26.8.2.7`.
- **The 3 × 2 cluster:**
  - A local Docker Compose copy of the proposed `prod-us` layout (D-05, D-07, D-09, D-10), on Altinity 26.3.33.10001.
  - Two `Replicated` databases, `Distributed` tables, an interserver secret and the `SQL_` prefix.
  - Replicated access storage and replicated SQL UDFs.
  - Keeper.
- **Data:** synthetic and deterministic.
  - 2.0M events over 99 days and 20 tenants. One tenant holds about 40%.
  - 5% Kafka redeliveries and 0.5% producer retries with a changed `time`.
  - About 4.7 tag keys per event.
- **Scope of local numbers:** latency and CPU figures from a 4-vCPU sandbox are plan evidence (rows read, which operator runs where, pass or fail). **They are not platform performance results.**
- **ClickHouse Cloud** could not be run. Cloud behaviour is cited or inferred and labelled as such.

## 1. Summary

1. **D-13 is not correct on `prod-us` as specified (verified).**
   - Under D-09, one Kafka batch for a tenant spreads over all three shards. A "max stored offset is visible" check can pass while some shards' rows are still missing on the replica being read.
   - With replication lag injected, the check passed with only **108 of 300** and **194 of 300** rows visible.
   - A reader that lands on a second replica with the default `prefer_localhost_replica = 1` also read stale data (193 of 300).
   - Fix: a **count-based completeness check in the same statement**. It never passed stale in the same test, it does not depend on replica routing, and it works unchanged on Cloud.
2. **The current `prod-us` build (26.3.16) breaks tenant isolation (verified).**
   - As the reader, `mergeTreeProjection(…, by_subject)` returned **all tenants' rows**, and `mergeTreeIndex(…)` returned index data for 10 tenants.
   - 26.3.33.10001 blocks `mergeTreeIndex` and applies the policy to `mergeTreeProjection`.
   - D-16 is therefore a hard gate. The verification gate needs these two table functions added to item 5.
3. **Drop both projections (verified on six builds).**
   - Neither `by_subject` nor `by_model` is used by any statement in section 9. `FINAL` disables the projection-index path, and only an unfiltered equality read without `FINAL` ever picked `by_subject`.
   - They cost about **20–34% more insert CPU**, **25% more merge CPU** and **12% of table bytes** locally.
   - They are also the `mergeTreeProjection` attack surface from item 2, and they block lightweight deletes.
4. **A producer retry with a changed `time` is counted twice under `FINAL` (verified: 10,000 of 10,000 such retries were double-counted).**
   - The sorting key starts with `time`, so identity `(tenant_id, source, id)` alone does not collapse rows.
   - This contradicts "each event exactly once, whatever the retry history".
   - The cheapest fix is at ingestion: reject or normalize a known identity with a different `time`. The `time` bound on HTTP ingestion already makes a 72-hour identity window enough.
5. **`FINAL` is not partition-local by default on any build from 26.3.16 to 26.8.2 (verified via the pipeline).** The automatic decision does not fire for this key. Setting `do_not_merge_across_partitions_select_final = 1` in the reader profile is safe, because every collapsible copy shares `time` and therefore its partition.
6. **Shard by `cityHash64(tenant_id)` unless pilot data shows one tenant above about 25% of volume. This replaces D-09 (verified mechanics, inferred trade-off).**
   - Every tenant read then goes to one shard instead of three (`optimize_skip_unused_shards = 1`). Locally that read 11–63% fewer rows.
   - Results are digest-identical to D-09 and to a single node on all 19 statements.
   - Field-value duplicates disappear: D-09 stores them about 3×.
   - The cost is skew. The 40% tenant put **59%** of rows on one shard.
7. **The notifier's load is set by the number of statements, not the platform.** Two simple exact changes, applied before any pre-aggregation:
   - **Coalesce:** evaluate each tenant once per hold window rather than once per partition (up to 12×).
   - **Combine:** fold a tenant's rules into one `sumIf` statement. Locally that read **10× fewer rows and used about 6× less CPU** for 10 rules (verified).
8. **The FOCUS monthly export (D2) exceeds the 512 MiB limit at about 240,000 groups per tenant-month (verified: `MEMORY_LIMIT_EXCEEDED` on a single node with CSV).**
   - The reader profile is `readonly = 1`, so the exporter cannot turn on spilling.
   - Chunk D2 per day: its charge periods are already daily. That brings the peak to **39 MiB** with identical output.
9. **The reader profile's `MAX` constraints do not bound defaults (verified).** With nothing set, the reader ran with `max_threads = 4` (the core count), `max_memory_usage = 0` (unlimited) and `max_execution_time = 0`. On a shared cluster the profile must set values, not only maxima. Explicit values are forwarded to remote shards (verified).
10. **D-12's premise does not hold (verified).** With a replica down, a `Replicated`-database DDL returns `TIMEOUT_EXCEEDED` (159) after `distributed_ddl_task_timeout`, yet it is **applied on every live replica and queued for the missing one**. "Fails instead of leaving a partial change" is not what happens; the runner must treat 159 as "applied, pending".

## 2. Use-case comparison

How to read this table:
- "Cloud" means a single logical shard with SharedMergeTree. Its behaviour is inferred from documentation, plus a single-node local run as a proxy for one-shard plans.
- "`prod-us`" means the proposed 3 × 2 layout, verified on the local copy.
- Confidence is about the verdict, not about absolute performance.

| Use case | ClickHouse Cloud | Altinity `prod-us` (D-09) | Verdict | Confidence |
| --- | --- | --- | --- | --- |
| 9.1 Ingestion | One part per partition per batch. Insert dedup on SharedMergeTree is replicated-style (inferred). Merges on shared storage. No fan-out. | A block splits by shard × partition, so up to 3× more parts per batch. Foreground insert waits for every shard with no timeout (`insert_distributed_timeout = 0`, verified). Identical redelivery is deduplicated per shard sub-block (verified). The materialized view fires per shard and stores about 3× the field values (verified). | Cloud is simpler. `prod-us` needs `insert_distributed_timeout` and acceptance of 3× part creation. Correctness is equal (verified for redelivery). | High |
| 9.2 Console aggregates | Whole query on one replica. Reads the tenant's range once. | Inner aggregation per shard, middle and outer on the initiator (verified). Reads 1.4–1.8× the rows of one node for month and quarter aggregates, and up to 3.6× for a one-day range, because the tenant spans 3 shards (verified). Tail latency is the slowest of 3 shards on a shared cluster (inferred). No pruning under D-09. | Comparable. Tenant sharding removes most of the gap. Projections help neither. | Medium |
| 9.3 Forecast | A5 like A1. A6 is table-free. | Same, with fan-out. A6 is table-free and identical. | Difference does not matter | High |
| 9.4 Budget balances | Exact with `FINAL` | Exact with `FINAL`; digests equal (verified) | Equal correctness. Fan-out cost as in 9.2. | High |
| 9.5 Notifier | One replica runs the whole statement, so a same-statement check is consistent. SharedMergeTree metadata visibility is asynchronous; `select_sequential_consistency` exists (cited). | D-13 as specified gives stale passes (verified). The count-based same-statement check is correct (verified). Load is 3 remote sub-queries per evaluation. | Use the count-based check on both platforms. Load fixes are platform-independent. | High (correctness), Medium (load) |
| 9.6 Event list | No read-in-order under `FINAL`. An unbounded request scans the tenant's history. | Same per shard. `LIMIT` is pushed down (`distributed_push_down_limit = 1`); the initiator merges 3 sorted streams (verified). | Bound the time window on both. No platform difference. | High |
| 9.7 Field-value pickers | Streaming `DISTINCT` (verified on one node) | Streaming `DISTINCT` per shard plus a merge on the initiator; results identical (verified). Stores about 2.96× the lookup rows (788,734 vs 266,905, verified). | Correct on both. Tenant sharding removes the duplication. | High |
| 9.8 Exports | D2 month fails at 512 MiB with CSV (single node, verified). D1 day is fine. | D2 month passed at about 500 MiB on the initiator with TSV. CSV failed at about 536 MiB (verified). | Chunk D2 by day on both. The edge-proxy idle limit is unknown for `prod-us`. | High (memory), Low (proxy) |
| 9.9 Migrations, mutations, retention | Shared database, SharedMergeTree. One observed mutation that never completes (brief). The `system.mutations` poll stays. | The `Replicated` database propagates DDL and mutations to all shards from one node (verified). `Distributed` rejects mutations (verified). A replica outage gives 159 with the change still applied (verified). | `prod-us` works but needs runner changes (159 handling, cluster-wide mutation polling). Use `ttl_only_drop_parts = 1` on both. | High |
| 9.10 Tenant isolation | Shared access. Fixes present on 26.4+ (brief). | With the secret, remote shards run as the reader, with `SQL_tenant_id` and limits forwarded (verified). Without it they run as `default`. 26.3.16 leaks through table functions (verified). | `prod-us` is acceptable only on ≥ 26.3.33 with D-10. Extend the gate. | High |
| 9.11 Move | Source side: read with `FINAL` or not, then export | Copies collapse under `FINAL` with D-09 or tenant sharding (verified for identical rows). Changed-`time` retries stay duplicated on both sides. | Feasible. Prefer an S3 hand-off. Reconcile per tenant-month. | Medium |

## 3. Findings (ranked by severity)

### F1 — D-13's max-offset check passes on incomplete data (Critical)

**Use case:** 9.5. **Platforms:** `prod-us`. **Label:** verified (`scripts/d13_trial.py`, local 26.3.33.10001 3 × 2).

**Evidence.** Fetches were stopped on one replica per shard, and a 300-row batch was inserted through `Distributed` with foreground insert.

| Configuration | Result |
| --- | --- |
| Writer default (`random`), reader `in_order` + max-offset `throwIf` | **Passed with 108 of 300 and 194 of 300 rows visible** (2 of 14 trials) |
| Reader on a second replica with `prefer_localhost_replica = 1` (default) | Read its lagging local replica: **193 of 300** |
| Writer and reader both `in_order` + `prefer_localhost_replica = 0` | Consistent while the first replicas were healthy |
| `countIf(kafka_partition = p AND kafka_offset BETWEEN lo AND hi) < expected` in the same statement | Never passed stale (0 of 8) |

**Why the max-offset check fails.** A batch's rows for one tenant land on all three shards. The max offset proves only that one shard's part is visible on the replica being read.

**Impact.** Alerts are evaluated on usage that is still missing. Thresholds are crossed late or not at all until the tenant's next usage — the exact failure the rejected alternative was meant to prevent.

**Recommendation.** Inside the threshold statement, `throwIf` on the count of the batch's rows for that tenant and partition: the notifier knows the count, because it consumed the batch. Keep `FINAL`; redelivered copies collapse, so the count stays exact.
- **Routing:** the check and the evaluation read the same replica per shard because they run in one statement, so `in_order` is no longer needed for correctness.
- **Dropped records:** records the worker drops make the check fail until the budget runs out, which matches today's behaviour.
- **Cloud:** the same statement works on Cloud. The whole query runs on one replica.

**Effort:** small (query builder). **Touches:** D-13, B5.

### F2 — `prod-us` 26.3.16 lets the reader bypass row policies (Critical)

**Use case:** 9.10. **Platforms:** `prod-us` today. **Label:** verified (`scripts/tf_bypass.sh`, `results/tablefunction-bypass.log`).

**Evidence.** As `aicm_reader` with tenant A's setting:

| Build | `mergeTreeProjection(ai_cost_management, usage_events, by_subject)` | `mergeTreeIndex(...)` |
| --- | --- | --- |
| 26.3.16.10001 | **200,000 rows (every tenant)** | **index data for 10 tenants** |
| 26.3.33.10001 | 80,124 rows (tenant A only) | refused: "Cannot read from `mergeTreeIndex` because a row policy is applied" |

**Impact.** A tenant-isolation break for any holder of the reader role on the current build.

**Recommendation.**
- Make D-16 a hard gate.
- Add `mergeTreeIndex`, `mergeTreeProjection`, `mergeTreeAnalyzeIndexes` and the other `mergeTree*` table functions to gate item 5.
- Dropping the projections (F3) removes the `mergeTreeProjection` surface entirely.

**Effort:** small. **Touches:** D-16, gate 9.10.

### F3 — Both projections are unused and cost insert, merge, storage and security (High)

**Use cases:** 9.1, 9.2, 9.10. **Platforms:** both. **Label:** verified (`results/version-behaviors.log`; the `projections` column of `results/queries/*.tsv`).

**Evidence.**

| Check | Result |
| --- | --- |
| Projection use by the 19 statements | None, on 26.3.16, 26.3.33, 26.4.5, 26.7.4 and 26.8.2 |
| Projection use under `FINAL` | Never: `subject = …` with `FINAL` read 523k–571k rows with no projection |
| Projection use without `FINAL` | `by_subject` served `subject = …` (155k rows). No section-9 statement has that shape (B5 is the only non-`FINAL` read and filters neither column). |
| Storage | 55 MiB of 447 MiB (12%) |
| Insert CPU (single thread) | +20–34% |
| `OPTIMIZE FINAL` merge CPU | +25% |
| Lightweight `DELETE`/`UPDATE` | Not supported on tables with projections (cited: [projections](https://clickhouse.com/docs/concepts/features/projections/projections)) |

**Recommendation.**
- Ship two migrations, `ALTER TABLE usage_events DROP PROJECTION by_subject` and `… by_model`, each with a cluster-only companion where needed.
- After that, the settings `deduplicate_merge_projection_mode` and `allow_part_offset_column_in_projections` can go too.

**Effort:** small. **Touches:** the TODO in the original migration, and Cloud's default that disables `_part_offset`.

### F4 — A producer retry with a changed `time` is double-counted (High)

**Use cases:** 9.1, 9.2, 9.4, 9.5. **Platforms:** both. **Label:** verified (local loads: 10,000 retries with `time` + 1 s; `FINAL` count 2,010,000 = 2,000,000 + 10,000).

**Why.** `ORDER BY (tenant_id, toStartOfMinute(time), time, source, id)` collapses only copies with the same `time`.

**Recommendation.** Enforce immutability of `time` per identity at ingestion. The HTTP endpoint already bounds `time` to 72 h, so a 72-hour identity window is sufficient. Two options:
- a worker-side deduplication store, as OpenMeter does;
- or a read-time `LIMIT 1 BY tenant_id, source, id` for authoritative totals. That one is expensive and not recommended.

Alternatively, record a deliberate decision that a changed `time` is a new event, and narrow the exactness requirement to match. **Effort:** medium. **Touches:** the correctness requirement.

### F5 — The D2 monthly FOCUS export can exceed 512 MiB; the exporter cannot enable spilling (High)

**Use case:** 9.8. **Platforms:** both. **Label:** verified (`scripts/probe_query.py`; single node and cluster).

**Evidence.** About 243k groups per tenant-month, from roughly 244k events. In this synthetic data the grouping keys are nearly unique, which is a worst case.

| Run | Peak memory | Result |
| --- | ---: | --- |
| Single node, `Null` output | 403 MiB | passed |
| Single node, CSV, parallel formatting on | 487–500 MiB | **failed** `MEMORY_LIMIT_EXCEEDED` |
| Cluster initiator, TSV | 500 MiB | passed, barely |
| Cluster initiator, CSV | would use 536 MiB | **failed** `MEMORY_LIMIT_EXCEEDED` |
| One day of D2 | 39 MiB (single node), 29 MiB (cluster initiator) | passed |

The reader cannot set `max_bytes_before_external_group_by`, `max_bytes_before_external_sort` or `output_format_parallel_formatting`: all fail with `READONLY`.

**Recommendation.** Chunk FOCUS exports per day. The output is identical: charge periods are daily and the sort key starts with `charge_period_start`. Alternatively, put spill settings in the reader profile. Measure groups per tenant-month in production (M5). **Effort:** small.

### F6 — The reader profile bounds only explicit settings, not defaults (High on `prod-us`)

**Use cases:** 9.2–9.8. **Platforms:** both, critical on the shared cluster. **Label:** verified.

**Evidence.**
- With no override, `getSetting` returned `max_threads = 4`, `max_memory_usage = 0` and `max_execution_time = 0` despite `MAX 2`, `MAX 536870912` and `MAX 300`.
- Explicit values were forwarded to every shard. Remote `query_log` showed `max_threads = 2`, `max_memory_usage = 536870912` and `SQL_tenant_id`.

**Recommendation.** Set values in the profile, for example `max_threads = 2 MAX 2` and `max_memory_usage = 536870912 MAX 536870912`, so the API and notifier get bounded resources by default. If the API needs more threads than the exporter, use two profiles (API vs exporter).

In the same profile change, add these settings, each verified as blocked today by `readonly = 1`:
- `do_not_merge_across_partitions_select_final = 1`;
- `optimize_skip_unused_shards = 1`, needed for F8;
- `prefer_localhost_replica = 0`, only if D-13 keeps `in_order`;
- `log_comment` (`CHANGEABLE_IN_READONLY`), for per-query attribution on the shared cluster.

**Effort:** small (append-only profile migration).

### F7 — D-12: a replica outage returns 159 but the change is applied (Medium-High)

**Use case:** 9.9. **Platforms:** `prod-us`. **Label:** verified.

**Evidence.** With one replica stopped:
- `ADD COLUMN … SETTINGS alter_sync = 2` returned `TIMEOUT_EXCEEDED` (159) after `distributed_ddl_task_timeout`. The column existed on the 5 live replicas, and the 6th applied it on restart.
- `MATERIALIZE COLUMN` behaved the same way.

**Recommendation.** The runner should treat 159 from a `Replicated` database as "applied on active hosts, pending on N". It should then:
- poll `clusterAllReplicas(…, system.mutations)` and the database's DDL queue until the change completes;
- never re-send a non-idempotent statement (`RENAME`, `MATERIALIZE`, type changes);
- record a distinct state, so a retry does not trip the checksum check.

**Effort:** medium. **Touches:** D-11, D-12, the runner.

### F8 — Sharding key: tenant-only sharding fits this workload better than D-09 (Medium)

**Use cases:** all reads, 9.5, 9.7, 9.11. **Platforms:** `prod-us`. **Label:** verified mechanics and parity, inferred recommendation.

**Evidence.** A parallel `cityHash64(tenant_id)` table on the same cluster, read by the same queries with `optimize_skip_unused_shards = 1`:

| Aspect | Tenant-only sharding vs D-09 |
| --- | --- |
| Shards per tenant read | **1** instead of 3 |
| Rows read (19 statements) | 11–63% fewer |
| Result digests | Equal to D-09 and the single node on all 19 |
| `FINAL` | Exact (2,010,000 on both) |
| Field-value duplication | Removed (D-09 stores about 2.96×) |
| Notifier per-tenant evaluation | Single-shard |
| Skew | The 40% tenant put **59%** of rows on one shard |
| Rebalancing | None automatic: "ClickHouse doesn't support automatic shard rebalancing" (cited: [scaling-clusters](https://clickhouse.com/docs/guides/sre/scaling-clusters)) |

Pruning requires the reader profile to carry `optimize_skip_unused_shards = 1` (F6).

**Recommendation.** Choose `cityHash64(tenant_id)` if pilot data (M4) shows no tenant above about 25% of volume and no tenant expected to exceed one shard's headroom. Otherwise keep D-09. Decide before the move, because resharding later is a copy.

**Effort:** small (a cluster-only migration change before go-live).

### F9 — `parallel_distributed_insert_select = 2` silently ignores the destination sharding key (Medium)

**Use cases:** 9.11 and any future resharding. **Platforms:** `prod-us`. **Label:** verified.

**Evidence.** `INSERT INTO ai_cost_management.usage_events_t SELECT … FROM ai_cost_management.usage_events` between two `Distributed` tables on the same cluster ran as shard-local `INSERT INTO ai_cost_management_local.usage_events_t` on each shard. Every tenant stayed on its original shard. Setting `parallel_distributed_insert_select = 0` routed rows by the new key.

**Recommendation.** Any internal copy or reshard between `Distributed` tables sets `parallel_distributed_insert_select = 0`. D-17 with a `remoteSecure` or S3 source is not affected. **Effort:** trivial.

### F10 — `FINAL` merges across partitions by default (Medium)

**Use cases:** 9.2–9.5, 9.8. **Platforms:** both. **Label:** verified (EXPLAIN PIPELINE on 26.3.16 to 26.8.2).

**Evidence.**
- The default plan has 2 `ReplacingSorted` streams that span partitions (18 and 27 parts).
- With `do_not_merge_across_partitions_select_final = 1` the plan has 3, one per monthly partition.
- `enable_automatic_decision_for_merging_across_partitions_for_final = 1` did not trigger for `PRIMARY KEY (tenant_id, toStartOfMinute(time))`.
- 0 identities had copies in different partitions.
- `enable_vertical_final` is already 1 by default.

**Recommendation.** Set it in the reader profile (F6) and measure on US dev (M3). It is safe because every collapsible copy shares `time`.

**Effort:** trivial.

### F11 — The notifier's per-rule, per-partition full-window scans (Medium)

**Use case:** 9.5. **Platforms:** both. **Label:** verified for the combined statement, inferred for the cost model.

**Cost model.** Per hold window (30 s), the cost is P × T × R full-month `FINAL` scans of each tenant's rows:

| Symbol | Meaning |
| --- | --- |
| P | evaluations per tenant per cycle, up to 12 |
| T | tenants with rules |
| R | rules per tenant |

**Evidence.** Locally, 10 separate rule statements read 909,680 rows and used 0.392 s CPU. One combined `sumIf` statement read 90,968 rows and used 0.067 s CPU, with the same exact sums.

**Recommendation.** Evaluate each tenant at most once per hold window across partitions, and fold its rules into one statement with per-rule `sumIf` and a grouped key where `each` applies. Together that is up to 12 × R fewer scans and stays exact.

Incremental materialized views see duplicates before `FINAL`, so they would need their own exactness proof; defer them until measured volumes require it. **Effort:** medium (notifier).

### F12 — Unbounded event-list requests scan the whole tenant history (Medium)

**Use case:** 9.6. **Platforms:** both. **Label:** verified.

**Evidence.**
- `ORDER BY time DESC LIMIT 21` with `FINAL` does a full sort.
- The bounded month read 91k rows. The unbounded request read **999k**, all of the tenant's rows.
- An all-descending order did not enable reading in order.
- On the cluster, each shard sorts its part and the initiator merges three streams.

**Recommendation.** Require or default a time window (for example 7 days), and page by time windows.

The `id` bloom filter (2.5 MiB) is used by no statement. Keep it only if support lookups by id exist. **Effort:** small.

### F13 — Field-value lookup: about 3× duplicates under D-09, no TTL, and +31% insert cost (Low-Medium)

**Use cases:** 9.1, 9.7. **Label:** verified.

**Evidence.**

| Measure | Result |
| --- | --- |
| Lookup rows | 788,734 on the cluster vs 266,905 on one node |
| Picker results | Identical; `DISTINCT` streams per shard and on the initiator |
| Materialized view cost (single node) | +31% latency and +32% CPU per 10,000-row batch |

**Recommendation.**
- With tenant sharding (F8) the duplication disappears.
- Without it, accept it.
- There is no time column for a TTL. If growth matters, add `last_seen` and a periodic rebuild later.
- Tenant erasure must include this table (F14).

### F14 — Tenant erasure on a cluster (Low-Medium)

**Use case:** 9.9. **Label:** verified.

**Evidence.**
- `ALTER TABLE ai_cost_management_local.usage_events DELETE WHERE tenant_id = … SETTINGS mutations_sync = 2`, issued once on one node, erased the tenant on all 6 replicas through the `Replicated` database.
- The same statement against the `Distributed` table fails with `NOT_IMPLEMENTED` (48).
- `usage_event_field_values` still held 38,971 rows for the tenant.

**Recommendation.** Document the local-database form. Erase both tables, then verify through `clusterAllReplicas`.

### F15 — Foreground insert has no timeout (Low-Medium)

**Use case:** 9.1. **Label:** verified default (`insert_distributed_timeout = 0`), inferred effect.

**Recommendation.** Set `insert_distributed_timeout`, for example 60 s, on the writer. A stalled shard then fails the batch, which the worker already retries, instead of hanging until network timeouts. The worker-lag monitors (900 s) cover this.

### F16 — The map bucket settings are inert at this tag density (Low)

**Use case:** Q6. **Label:** verified.

**Evidence.**
- `map_buckets_min_avg_size = 32` and `map_serialization_version_for_zero_level_parts = 'basic'` by default.
- At about 4.7 keys per event, `with_buckets` writes a single bucket regardless of strategy.
- `basic`, sqrt/32 and constant/8 had identical bytes and reads.

**Recommendation.** Drop the three settings, or leave them. Revisit only if `avg(length(tags))` is 32 or more (M6).

### F17 — Grants and access objects on the cluster (Low)

**Label:** verified.

**Evidence.**
- Replicated access storage plus `user_defined_zookeeper_path` made one `CREATE` reach all 6 nodes: roles, policies, users, the definer user and both UDFs.
- `CREATE TABLE … AS … ENGINE = Distributed` keeps `MATERIALIZED` columns and drops indexes and projections, which is correct.
- Grants on the `Distributed` tables are needed in addition to the local ones (cluster-only migrations).

**Recommendation.**
- **D-14:** the data team configures replicated access storage and replicated UDFs on `prod-us`; otherwise every object needs creating per node.
- **D-11:** add both checks to the preflight.

## 4. Proposed decisions

| Decision | Position | Reason |
| --- | --- | --- |
| D-06 | Keep, time-boxed | No alternative while the provider cannot grant `CREATE ROLE`. Rotate the password after stage 2. |
| D-07 | **Keep** | Verified: DDL in the local `Replicated` database reaches all shards without `ON CLUSTER`, and same-named `Distributed` tables let application SQL stay identical. |
| D-08 | Keep, with additions | The materialized view fires per shard on rows arriving through `Distributed`, so skipping the backfill is right. Each change to a local table also needs the `Distributed` table's columns and grants updated, as a cluster-only companion. |
| D-09 | **Change** to `cityHash64(tenant_id)`, conditional on M4 | F8. Keep `(tenant_id, id)` only if one tenant would dominate a shard. |
| D-10 | **Keep (required)** | Verified: without the secret, remote shards run as `default`. Row policies of the initial user still applied on 26.3.33, but grants and profile limits are not enforced remotely. With it, remote shards ran as `aicm_reader` with all settings forwarded. |
| D-11 | Keep, extend | Also check: `custom_settings_prefixes` contains `SQL_`; replicated access storage and UDF path; the reader's effective `prefer_localhost_replica`, `load_balancing` and `optimize_skip_unused_shards`; `parallel_distributed_insert_select` for copy jobs. |
| D-12 | **Change** | F7: `alter_sync = 2` does not make an outage atomic. Handle 159 as pending and poll mutations cluster-wide. |
| D-13 | **Replace** | F1: count-based completeness check in the same statement, on both platforms. Drop B5. `in_order` becomes optional. |
| D-14 | Keep | Plus replicated access storage (F17). |
| D-15 | Keep | The edge proxy's 330 s idle tolerance must be measured (M9) before exports move. |
| D-16 | **Keep (hard gate)** | F2 |
| D-17 | Keep the watermark plan; **prefer the S3 hand-off** | Both sharding options co-locate copies, so overlaps collapse (verified for identical rows). S3 decouples the clusters, is resumable per month, and avoids long cross-internet queries and credentials in query text on a shared cluster (assessed). Reconcile per tenant × month: raw `count()`, `FINAL` `count()` and `FINAL` `sum(reported_cost_usd)` below the watermark on both sides. Changed-`time` duplicates (F4) appear identically on both sides and do not explain differences. |

## 5. Measurement requests

Run as an admin, read-only. On Cloud use `clusterAllReplicas(default, …)`; on `prod-us` use the cluster name.

**M1 — Versions and features (all three platforms).**
```sql
SELECT version(), getSetting('async_insert'), getSetting('allow_experimental_full_text_index');
SELECT name, value FROM system.merge_tree_settings WHERE name IN ('allow_part_offset_column_in_projections','map_buckets_min_avg_size','map_serialization_version_for_zero_level_parts');
SELECT engine_full FROM system.tables WHERE database = currentDatabase() AND name = 'usage_events';
```
Changes a recommendation if: Cloud is below 26.2 (text index not GA, #96794), or `prod-us` is below 26.3.33 (F2).

**M2 — Projection use in production (US dev and stage 1).**
```sql
SELECT arrayJoin(projections) AS p, count() FROM clusterAllReplicas(default, system.query_log)
WHERE event_date >= today() - 14 AND has(tables, concat(currentDatabase(), '.usage_events')) AND type = 'QueryFinish' GROUP BY p;
```
Changes a recommendation if: either projection appears in a meaningful share of queries, which would block F3.

**M3 — `FINAL` cost by statement shape (US dev under the planned 3× load test).**
```sql
SELECT normalized_query_hash, any(substring(query, 1, 80)) q, count() n, quantile(0.95)(query_duration_ms) p95_ms, avg(read_rows) rows, avg(memory_usage) mem
FROM clusterAllReplicas(default, system.query_log) WHERE event_time > now() - INTERVAL 1 HOUR AND type = 'QueryFinish' AND user = '<reader user>'
GROUP BY normalized_query_hash ORDER BY sum(query_duration_ms) DESC LIMIT 20;
```
Repeat with `SETTINGS do_not_merge_across_partitions_select_final = 1` on the top statements. Changes a recommendation if: p95 stays above 1 s, which would warrant pre-aggregation for that shape.

**M4 — Tenant distribution (stage 1).**
```sql
SELECT tenant_id, count() c, round(c / sum(c) OVER () * 100, 1) pct FROM usage_events GROUP BY tenant_id ORDER BY c DESC LIMIT 10;
```
Changes a recommendation if: the top tenant is above about 25%, or is projected to be, which means keep D-09 instead of F8.

**M5 — FOCUS groups per tenant-month.**
```sql
SELECT tenant_id, toStartOfMonth(time) m, uniqExact(toStartOfDay(time), provider, model, region, subject, application_id, teams, mapFilter((k, v) -> has(['env','cost_center'], k), tags)) g
FROM usage_events FINAL GROUP BY tenant_id, m ORDER BY g DESC LIMIT 5;
```
Changes a recommendation if: `g` is above about 150,000, which makes the per-day chunking in F5 mandatory.

**M6 — Tag density.**
```sql
SELECT avg(length(tags)), quantile(0.99)(length(tags)), uniqExact(arrayJoin(mapKeys(tags))) FROM usage_events;
```
Changes a recommendation if: the average is 32 or more, in which case the bucket settings (F16) matter.

**M7 — Effective reader and writer settings on `prod-us`, run as each user.**
```sql
SELECT getSetting('prefer_localhost_replica'), getSetting('load_balancing'), getSetting('max_threads'), getSetting('max_memory_usage'),
       getSetting('parallel_distributed_insert_select'), getSetting('insert_distributed_timeout'), getSetting('optimize_skip_unused_shards');
SELECT name, value FROM system.server_settings WHERE name IN ('custom_settings_prefixes','user_defined_zookeeper_path');
SELECT * FROM system.user_directories;
```
Changes a recommendation if: `custom_settings_prefixes` lacks `SQL_`, which blocks the reader; or access storage is not replicated, which affects D-14 and F17.

**M8 — Interserver secret in effect (`prod-us`).** As the reader, run any `Distributed` count with `query_id = 'secret-check'`, then as admin:
```sql
SELECT hostName(), user, initial_user FROM clusterAllReplicas('<cluster>', system.query_log) WHERE initial_query_id = 'secret-check' AND NOT is_initial_query;
```
Changes a recommendation if: `user` is not the reader, meaning D-10 is not in effect, which is a blocker.

**M9 — Edge-proxy idle tolerance (`prod-us`, through the real path).**
```sql
SELECT sleepEachRow(3) FROM numbers(100) SETTINGS max_block_size = 1, max_execution_time = 300;
```
Run it as the reader over HTTPS 8443 with `wait_end_of_query = 1`; the connection then sits idle for about 300 s. Changes a recommendation if: the connection drops, so exports must use smaller chunks or a different transport.

**M10 — Replication lag profile (`prod-us`).**
```sql
SELECT hostName(), database, table, absolute_delay, queue_size, inserts_in_queue FROM clusterAllReplicas('<cluster>', system.replicas) WHERE absolute_delay > 0 OR queue_size > 0;
SELECT hostName(), type, count(), max(now() - create_time) FROM clusterAllReplicas('<cluster>', system.replication_queue) GROUP BY 1, 2;
```
Sample every minute for a day. Changes a recommendation if: p99 lag exceeds the notifier's 30 s budget, which makes F1 urgent and calls for a larger visibility budget.

**M11 — SharedMergeTree read-after-write on US dev.** Insert one marker row with a known `id`, then immediately and repeatedly run:
```sql
SELECT hostName(), count() FROM clusterAllReplicas(default, usage_events) WHERE tenant_id = <T> AND id = '<marker>' GROUP BY 1;
```
Once with `select_sequential_consistency = 0` and once with `= 1`. Changes a recommendation if: lag is persistently above 0 at the default, which confirms that the same-statement check is also needed on Cloud.

**M12 — Noisy neighbours (`prod-us`).**
```sql
SELECT user, count(), sum(ProfileEvents['OSCPUVirtualTimeMicroseconds'])/1e6 cpu_s, max(memory_usage) FROM clusterAllReplicas('<cluster>', system.query_log)
WHERE event_time > now() - INTERVAL 1 DAY AND type = 'QueryFinish' GROUP BY user ORDER BY cpu_s DESC;
SELECT * FROM system.workloads; SELECT * FROM system.resources; SELECT * FROM system.quotas;
```
Changes a recommendation if: other users dominate CPU at AICM peak times; then the data team should define workload classes.

**M13 — Mutation and Keeper health (`prod-us`, before and after each migration).**
```sql
SELECT hostName(), database, table, command, is_done, latest_fail_reason FROM clusterAllReplicas('<cluster>', system.mutations) WHERE NOT is_done;
SELECT hostName(), database, table FROM clusterAllReplicas('<cluster>', system.replicas) WHERE is_readonly;
```
Changes a recommendation if: any rows appear, which blocks the next migration.

## 6. Open questions

| Question | What would resolve it |
| --- | --- |
| Production volume, tenant skew and growth | The load test and M4 |
| Do HTTP-ingestion producers ever retry with a changed `time` (F4)? | Client documentation, or comparing the counts of duplicate identities with distinct `time` in stage 1 |
| Does `prod-us` already use replicated access storage, `custom_settings_prefixes = SQL_`, an interserver secret and a default profile that sets `prefer_localhost_replica` / `load_balancing`? | M7, M8 |
| Does the Altinity edge proxy tolerate a 330 s idle HTTP connection? | M9 |
| SharedMergeTree metadata-visibility lag on Cloud, and the cost of `select_sequential_consistency` there | M11 |
| How Cloud executes the failed mutation from 2026-09-09 (error 517 loop) | ClickHouse Cloud support ticket, plus `system.mutations` history on US dev |
| `prod-us` backup method and retention (`clickhouse-backup` vs `BACKUP … TO S3`), and monitoring without Datadog | The data team's runbook. At minimum: replica lag, read-only replicas, parts per partition, failed mutations, Keeper latency, `system.errors` |
| Workload scheduling on `prod-us` | `CREATE WORKLOAD`/`CREATE RESOURCE`, available since 24.11 (PR #69187); the system tables are present on 26.3.33. Whether the data team will define AICM workload classes is open. |

## Appendix A — Answers to section 10

1. **Feature acceptance.** Yes.
   - All 19 schema and access statements apply unchanged on 26.3.16 and 26.3.33 (Altinity), 26.3.33 (upstream), 26.4.5, 26.7.4 and 26.8.2 (single node). The `ReplicatedReplacingMergeTree` versions also apply on the 26.3.33 cluster (verified, `results/q1-features/`).
   - No experimental flags were needed. `enable_full_text_index` and `allow_part_offset_column_in_projections` default to 1 in open source, and the text index is GA in 26.2 (#96794).
   - `FINAL` plans, projection selection, map bucketing and event-list scans behave identically from 26.3.16 to 26.8.2 (verified).
   - The only relevant default that differs is `read_in_order_use_virtual_row`: 0, then 1 in 26.8.
2. **Sharding key.** See F8. Choose `cityHash64(tenant_id)` unless M4 shows dominance.
3. **`FINAL` partition-locality.** It is not partition-local without the setting (F10). `enable_vertical_final` is already on. Cost is roughly proportional to the rows in the tenant's range; on a cluster add 1.4–1.8× rows read under D-09, and up to 3.6× for narrow ranges.
4. **Projections.** Unused by every statement; drop them (F3).
5. **Profile constraints.** The constraints do not bound defaults (F6). Explicit values are forwarded to the remote shards.
6. **Map buckets.** Inert below 32 keys per row (F16).
7. **Notifier sustainability.** Not sustainable as statement count grows. Coalesce per tenant and combine rules (F11). Exact, no new tables.
8. **D-13.** Incorrect (F1). On Cloud, the same-statement check is equivalent. `select_sequential_consistency = 1` is the documented SharedMergeTree alternative (cited: [SMT consistency](https://clickhouse.com/docs/products/cloud/features/infrastructure/shared-merge-tree#consistency)).
9. **Noisy neighbours.** The risky sources:
   - exports (F5);
   - notifier fan-out (F11);
   - unbounded event lists (F12);
   - default-unbounded reader resources (F6);
   - the move (D-17);
   - mutations (F7, F14).

   What the data team should set:
   - per-profile values, not only maxima;
   - `max_concurrent_queries_for_user` (currently 0);
   - quotas;
   - a workload class for AICM.
10. **Backups, upgrades and monitoring.**
    - **Cloud:** managed backups (daily with 24 h retention by default; configurable on Scale and Enterprise) and release channels.
    - **`prod-us`:** backups need clickhouse-backup or `BACKUP … TO S3` with a tested restore. Upgrades are operator rolling updates. Monitoring must come from system tables or Prometheus, given there is no Datadog.
    - All assessed, citations in `docs/research/` of this repository.
11. **Measurements.** See section 5. Blocking results:
    - M1 showing `prod-us` below 26.3.33;
    - M7 showing `custom_settings_prefixes` without `SQL_`;
    - M8 showing remote queries not running as the reader;
    - M9 dropping the connection;
    - M13 showing stuck mutations or read-only replicas.

## Evidence index

| Claim | How to reproduce | Output |
| --- | --- | --- |
| Feature acceptance and defaults on six builds | `scripts/q1_features.sh <image>…` | `results/q1-features/*.log` |
| 3 × 2 cluster mirroring D-07, D-09, D-10 | `cluster/docker-compose.yml`, `cluster/sql/*.sql`, `scripts/run_sql.py` | — |
| Synthetic data (deterministic) | `scripts/gen.py`, `sql/gen/usage_events_gen.sql` | — |
| Per-query digests, rows read and projection use (single, D-09, tenant-sharded) | `scripts/run_queries.py` | `results/queries/*.tsv` |
| Version behaviours (`FINAL` plan, projections, event list, maps) | `scripts/version_behaviors.sh <image>` | `results/version-behaviors.log` |
| Table-function row-policy bypass (26.3.16 vs 26.3.33) | `scripts/tf_bypass.sh <image>` | `results/tablefunction-bypass.log` |
| D-13 trials (lag injection) | `scripts/d13_trial.py` | `results/d13-trials.log` |
| Export memory | `scripts/probe_query.py` | `results/export-memory.log` |
| Tenant-sharded variant queries | `QDIR=sql/queries-tenant-sharded scripts/run_queries.py …` | `results/queries/cluster3x2-tenantshard.tsv` |

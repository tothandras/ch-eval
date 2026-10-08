## Context

This repo forks `tothandras/clickhouse-benchmarks` at `542b1e5`. Only the Go
module path is renamed. The fork gives us:

- a perf runner that delegates timing to `clickhouse-benchmark` and reads CPU
  from `query_log`;
- a value digest per query, the parity gate;
- a COGS runner that attributes CPU to insert, merge, query and idle from
  `query_log`/`part_log` and prices it with Cloud profiles;
- an OpenSpec workflow.

Its two OpenMeter scenarios are `baseline-openmeter` (`data String`) and
`proposal` (`data JSON CODEC(ZSTD(3))` + `bloom_filter` on `id`).

Facts that shape this design. Sources are in `docs/research/`.

**Versions**
- **Pinned builds:** Altinity Stable `26.3.33.10001.altinitystable` and
  upstream `26.3.33.24` share the same 26.3.33 base, which gives a clean
  build-vs-build control.
- **Cloud is ahead:** Cloud runs 26.6 on the Fast/Regular channels and 26.4 on
  Slow.

**OpenMeter today**
- One `om_events` table, created at process start with
  `CREATE TABLE IF NOT EXISTS`.
- `ENGINE = MergeTree` is hard-coded.
- No ClickHouse migrations.
- Postgres uses golang-migrate naming (`YYYYMMDDHHMMSS_name.{up,down}.sql`) and
  a `schema_om` version table.

**Server defaults (26.2+)**
- `async_insert=1`, `wait_for_async_insert=1`, and deduplication on for every
  insert.
- Plain `MergeTree` deduplicates nothing (`non_replicated_deduplication_window=0`).

**OTel schemas**
- The exporter (contrib v0.162.0) and ClickStack both recommend
  `create_schema: false`.
- On ≥26.2 servers both use `text` indexes.

**Sandbox constraints**
- The build sandbox can run Docker and kind (patched for cgroup v1).
- It cannot reach clickhouse.com or builds.clickhouse.com. Docs are cited by
  URL and verified against the docs source in GitHub.
- Local smoke run, 2026-10-08:
  - Both 26.3 images run the unchanged harness.
  - `baseline-openmeter` and `proposal` digests match across upstream and
    Altinity: 30/30 and 34/34 at 200k rows.

## Goals / Non-Goals

**Goals:**
- Measured, reproducible answers per workload (OpenMeter events, OTel traces,
  OTel metrics, OTel logs) for `cloud` vs `altinity`, with `upstream-oss`
  isolating build effects.
- One migration source of truth that applies on every target, plus measured
  online schema changes.
- A cited, probed feature matrix. Design-relevant differences are called out
  at the top of the report.
- Self-managed COGS that is directly comparable to the existing Cloud COGS,
  reconciled against real bills.
- `REPORT.md` generated from results, with a recommendation per workload and
  the conditions under which it flips.
- Owner additions on 2026-10-08:
  - **Replication and sharding on self-managed (D13).** What it costs and
    which correctness risks it carries, measured, including the
    sharding-related risks in the owner's SharedMergeTree analysis.
  - **Compute/data separation vs colocation (D14).** Cloud's separated compute
    and object storage against self-managed local disks and S3-backed
    MergeTree.
  - **On-demand scaling and rightsizing (D15).** For compute and for storage,
    on both platforms.

**Non-Goals:**
- **Drawing performance or cost conclusions from local tiers.** Compose and
  kind are for correctness, parity and smoke tests only, and the report
  generator enforces this.
- **Sharding Cloud.** Cloud scales by replicas over shared storage and is not
  sharded here. Sharding is evaluated as a self-managed design option (D13),
  and the comparison is Cloud's unsharded scale-out against self-managed
  sharded and unsharded topologies.
- **More than 2 shards on paid tiers.** 2 shards is enough to expose routing,
  completeness and rebalancing behaviour. Larger shard counts are assessed,
  not measured.
- **Re-litigating the `om_events` design.** The `proposal` recommendation is an
  input. If a finding contradicts it, work stops and the owner is asked.
- **Altinity Antalya** beyond a feature-matrix row.

## Decisions

### D1. Fork, don't rewrite (ADR-0001)
- Keep `runner/`, `runner/digest.go`, `cogs/attribution.go`, `seed/`,
  `accounting/` collectors and `pricing/` semantics.
- All extensions are additive or minimal, documented edits. Anything else
  needs a new ADR.
- Alternative rejected: a fresh harness. It would lose the parity and COGS
  methodology the existing results were reconciled with.

### D2. Target is a manifest plus live metadata (ADR-0002)

**Manifest.** `targets/<name>/target.json` declares:
- `platform`: `cloud`, `altinity` or `upstream-oss`;
- `image` or service;
- `tier`;
- `dialect`: `single`, `replicated` or `cloud`;
- `cluster` name for `clusterAllReplicas` and `ON CLUSTER`;
- `storage`: `local`, `ebs`, `s3` or `shared`;
- `shape`: replicas, vCPU, GiB;
- `location`.

**Connection.** Connection strings never live in the repo. They come from
`CLICKHOUSE_DSN_<TARGET>` or `CLICKHOUSE_DSN`.

**Live metadata.** On every run, `bench/target` captures into `metadata.json`:
- `version()`, `system.build_options`;
- `system.clusters`, `system.macros`, disks and storage policies;
- changed rows from `system.settings`, `system.merge_tree_settings` and
  `system.server_settings`;
- `CGroupMaxCPU` and the memory limit;
- client host and harness commit.

**Manifest vs reality.** A mismatch between the manifest and the live target
(for example, the version) fails the run unless `--allow-drift` is given.

### D3. Results layout and run IDs

**Path.** Results go to `results/<workload>/<scenario>/<target>/<run-id>/`,
containing `run.json` (the unchanged `runner.Run` schema plus a `target`
block), `report.md`, `metadata.json` and, for COGS, `cogs.json`.

**Run ID.** `<UTC yyyymmddThhmmssZ>-<6-char ULID suffix>`. The same run ID tags
every cloud resource the run creates.

**Tier gate.** `tier` is recorded. `bench report` uses only `aws` and `cloud`
tiers for performance and cost.

### D4. Migrations (ADR-0003)

**Why custom.** OpenMeter has no ClickHouse migration tool to align with. We
align instead with its Postgres convention and add the two things golang-migrate
lacks for this job: dialect templating, and per-migration resource capture for
online-change measurements.

**Files.** `YYYYMMDDHHMMSS_<name>.up.sql` / `.down.sql`, optionally with a
`.tmpl` suffix. Go `text/template` renders them with a small, closed helper set:

| Helper | Renders as |
| --- | --- |
| `{{ onCluster }}` | `ON CLUSTER '<cluster>'` on `replicated`; on `cloud`, empty or `ON CLUSTER default` depending on probe ST-03 |
| `{{ engine "MergeTree" }}` | `MergeTree` / `ReplicatedMergeTree('/clickhouse/tables/{shard}/{database}/{table}', '{replica}')` / `SharedMergeTree` |
| `{{ engine "ReplacingMergeTree" "ver" }}` | the same mapping with engine arguments |
| `{{ db }}` | the target database |
| `{{ if dialect "cloud" }}…{{ end }}` | settings that only apply on some dialects |

`bench migrate render --target X` writes plain SQL that golang-migrate can
apply, so OpenMeter could adopt the files without our runner.

**Runner** (`bench migrate up|down|status|verify`):
- Each statement is one `Exec`, in order.
- State lives in `schema_migrations_ch` (version, name, checksum of the
  rendered SQL, dirty, applied_at, duration_ms, target). Its engine and
  `ON CLUSTER` follow the dialect, so the state is replicated with the data it
  describes.
- State reads use `SYSTEM SYNC REPLICA` on `replicated` and
  `select_sequential_consistency=1` on `cloud`.
- `alter_sync=2` and `mutations_sync=2` are set (probe-verified on Cloud).
- Every statement is tagged with a `log_comment`. Duration, CPU, read and
  written bytes come from `query_log`/`part_log`, and mutations are tracked
  through `system.mutations` until done.

**Idempotency.**
- Up migrations must use `IF [NOT] EXISTS` forms. A lint fails on bare
  `CREATE` or `DROP`.
- Re-applying an applied version is a no-op (checksum must match; a mismatch is
  an error).
- A test applies up → down → up → up on every dialect.

**Online changes** (`migrations/<workload>/`), one realistic one per workload,
each measured per target:

| Workload | Change |
| --- | --- |
| openmeter | `data String` → `JSON` on populated `baseline_openmeter_events`; add the `id` bloom index + `MATERIALIZE INDEX` |
| otel-traces | add a materialized hot-attribute column + `MATERIALIZE COLUMN` |
| otel-metrics | add a rollup MV + backfill |
| otel-logs | add a `text` index + `MATERIALIZE INDEX` |

### D5. Workload layout

**Directory.** `scenarios/<workload>/<variant>/` holds:
- `init/`, the migration set;
- `queries/*.sql`;
- `queries/mix.json`;
- `params.go`, a provider registered in `bench/workloads`.

**Discovery.** It becomes two levels deep. `--workload` and `--scenario`
select.

**OpenMeter.** Scenarios keep their table names, so the query files stay
byte-identical.

**OTel.** Each OTel scenario gets its own database
(`<workload>_<variant>`, for example `otel_traces_clickstack`), so the
published DDL stays verbatim apart from database and engine placeholders.

**Parameters.** Derived deterministically from the generator config, for
example a TraceId known to exist or a high-cardinality label value present in
the data. Identical data therefore yields identical digests on every target.

### D6. OTel generators and variants (ADR-0004)

**Determinism.** Generators are pure functions of `(seed, index)`, like
`bench/seed`.

**Traces generator.**
- A fixed service graph: about 30 services, each with 5–40 operations.
- Each trace is a root span plus children that follow the graph, with
  log-normal per-operation durations and per-operation error rates.
- Span attributes depend on the span kind: `http.*`, `db.*`, `rpc.*`,
  `messaging.*`.
- Resource attributes come per service instance: `k8s.pod.name`,
  `service.version`, `host.name`.
- Events and links are occasional.
- Cardinality is tunable through these knobs:

  | Knob | Typical values |
  | --- | --- |
  | services | ~30 |
  | ops/service | 5–40 |
  | pods/service | — |
  | high-card attributes | `user.id`, `url.full` |
  | depth/fanout | — |

**Metrics generator.**
- Metric families: monotonic cumulative sums with occasional resets, gauges,
  explicit-bucket histograms, exponential histograms and summaries.
- Series = metric × attribute combinations, with a series-count knob.
- The `high-cardinality` profile pushes a label (for example `pod`, `user.id`)
  to 10⁵–10⁶ series.
- A fixed scrape interval.

**Logs generator.**
- Per-service message templates with variable parts.
- A severity distribution.
- A configurable fraction correlated to generated traces through the same
  TraceId/SpanId functions.
- Planted rare tokens for needle-in-haystack searches.

**Ingest paths.**
- **Direct native insert** through a per-variant column mapper, for scale.
- **OTLP over gRPC** to `otel/opentelemetry-collector-contrib:0.162.0`, with the
  ClickHouse exporter at `create_schema: false`, `async_insert` and
  `sending_queue.batch` on. This runs at least once per target.

**Parity across paths.** Both paths must yield identical digests on a small
dataset. That verifies the mapper matches the exporter.

**Variants.**
- `exporter-default`: the v0.162.0 templates, `json: false`.
- `clickstack`: hyperdx main seeds.
- `tuned`: chosen and justified in Phase 4 from A/B evidence. Candidates are
  materialized hot attributes, a TraceId lookup table or projection, and a
  series-id dimension table for metrics.

**OTel queries.** Logs queries are full-text, severity/service filters,
attribute filters, trace correlation, top-N messages, and 1h/24h/7d scans.

### D7. Replica-consistency experiment (from the owner's SharedMergeTree research)

**Writer.**
- Inserts OpenMeter-shaped events for one namespace at a steady rate.
- Every batch carries a monotonically increasing `batch_seq`, and the writer
  records the ack time and the cumulative acked total.

**Probes.**
- **Per-replica visibility.** Poll
  `clusterAllReplicas(<cluster>, db.t)` grouped by `hostName()` and record the
  ack → visible lag per replica. This works on ReplicatedMergeTree (local
  copies) and on SharedMergeTree (per-replica metadata view).
- **Client-observed reads.** N readers read the cumulative total through the
  load balancer (k8s Service or Cloud endpoint). Each read records:
  - *stale*: the observed total is below the acked total at query start;
  - *monotonicity violation*: the total decreased relative to that reader's
    previous read;
  - *deficit*: the size of the shortfall in rows and in seconds.

**Policies under test.**

| Engine | Policies |
| --- | --- |
| ReplicatedMergeTree | default; `insert_quorum=<replicas>` + `insert_quorum_parallel=0` + `select_sequential_consistency=1`; `SYSTEM SYNC REPLICA … LIGHTWEIGHT` before the read |
| SharedMergeTree | default; `select_sequential_consistency=1` |

Each policy runs × sync/async insert, under light and heavy ingest, and
during a replica restart.

**Reported.**
- The stale-read rate, lag p50/p99/max, violations per 10k reads, and the
  insert/read latency overhead of each policy.
- The experiment answers the research's hypotheses directly. If a result
  contradicts the proposal's assumptions, the owner is asked before the report
  states it.

### D8. Topologies and matched shapes (ADR-0005)

**Local tiers**
- `local-single`: one node, used by CI.
- `local-cluster`: compose with 2 replicas × 1 shard + 3 Keeper + an S3
  emulator.
- `kind`: Altinity operator CHI with 1 shard × 2 replicas + CHK with 3 Keeper.

**Perf/COGS tier**
- One AWS region, chosen by the owner.
- **Cloud:** N replicas × 16 GiB / 4 vCPU.
- **Self-managed:** EKS + Altinity operator. Same-image pods with Guaranteed
  QoS at 4 vCPU / 16 GiB, one pod per dedicated node (m7i.2xlarge-class).
  Keeper runs on 3 small dedicated nodes.
- **Client:** a VM in the same AZ.
- **Storage modes (self-managed):**
  - `ebs`: gp3, with baseline IOPS and throughput recorded;
  - `s3`: an s3 disk + filesystem cache on local disk, no zero-copy.
- `altinity` and `upstream-oss` use identical manifests; only the image
  differs.

**`upstream-cloudver` control (proposed).** Upstream at Cloud's version on the
self-managed topology. A result that differs between cloud and altinity but
not between `upstream-cloudver` and cloud is attributed to the version gap.

### D9. Self-managed COGS

**Profile schema.** The pricing schema gains `kind: self-managed` with these
line items:
- compute: `[{role, count, usd_per_hour}]` for ch, keeper, eks, plus an optional NAT;
- EBS: GB-month, IOPS and throughput above baseline;
- S3: storage GB-month and per-1k PUT/GET/LIST;
- egress per GB;
- ops: `usd_per_hour` × `hours_per_month`, with a sensitivity sweep.

**Billed-shape semantics.** The window cost is the sum of node-hours. The CPU
share splits it exactly as `pricing.BilledShape` does today, so attribution is
untouched.

**S3 requests** come from `ProfileEvents`/`system.events` deltas.

**Cost curves.** $/month vs utilization per target. Cloud's curve includes
idling; self-managed is flat at the node floor. Crossover points are computed,
not eyeballed.

**Reconciliation.**
- Cloud: the usage-statement CSV (existing).
- AWS: Cost Explorer, grouped by the `run-id` cost-allocation tag. Daily
  granularity is flagged, as it is for Cloud.

### D10. Feature matrix as data
- `docs/feature-matrix.yaml` holds the rows: id, capability, citations,
  per-target hypothesis, probe path.
- Probes are SQL files with an expected-outcome header (`expect: ok`,
  `error:<code>` or `value:<v>`), or Go tests for client and replication
  behaviour.
- `bench probe --target X` records outcome, error code and `version()`.
- `bench feature-matrix` renders the markdown. A row with no probe result
  stays marked as a hypothesis.

### D11. Report generation
- `bench report` walks `results/`, picks the latest valid run per
  (workload, scenario, target, tier ∈ {aws, cloud}), and writes `REPORT.md`
  with these sections:
  - design-decision callouts;
  - executive summary;
  - per-workload verdict;
  - feature-matrix highlights;
  - cost comparison with crossovers;
  - caveats;
  - open questions.
- Verdict rules live in code and in `docs/methodology.md`:
  - **Latency or CPU:** a difference counts only when the median across
    queries exceeds run-to-run noise, measured by `--repeat` variance.
  - **Cost:** compared at matched utilization.
- Each verdict lists its flip conditions, computed from the curves (for
  example, "flips to Cloud below X% utilization or above Y ops-hours/month").

### D12. CI
- **Every PR:**
  - `gofmt`, `go vet`, `go test ./...`, `openspec validate --strict`;
  - migration tests on `local-single` and `local-cluster` for both
    `altinity` and `upstream-oss`;
  - a smoke bench at ~100k rows per workload with the cross-target parity
    compare.
- **Nightly:** the kind + operator job.
- GitHub runners use cgroup v2, so the sandbox's kind patches are guarded by
  detection.

### D13. Replication and sharding on self-managed (owner addition)

**Why.** On Cloud, scaling out means adding replicas over one shared copy of
the data. On self-managed, scaling out means more replicas (each a full copy)
or more shards (a partitioned copy behind a `Distributed` table). Each choice
carries operational and correctness consequences that OpenMeter would own. The
owner's analysis names specific sharding risks: shard coverage, the
Distributed insert queue, partial results, and resharding. They are measured
here, not assumed.

**Topologies (self-managed, same image per build)**

| ID | Shards × replicas | Role |
| --- | --- | --- |
| `1x2` | 1 × 2 | Baseline, HA, unsharded |
| `1x4` | 1 × 4 | Scale out by replicas, optionally with parallel replicas |
| `2x2` | 2 × 2 | Sharded: `Distributed` over `ReplicatedMergeTree`, `internal_replication=true` |

Local tiers run `1x2` and `2x2` for correctness (compose `local-cluster` and
kind). Paid tiers run all three at matched total vCPU, against Cloud at the
same replica count.

**Migration templating.** A `shardedTable` helper renders, for sharded
topologies:
- a local table `<name>_local`, created `ON CLUSTER` with
  `ReplicatedMergeTree`;
- a `Distributed` table `<name>` over it, with a declared sharding key.

Queries keep addressing `<name>`, so query files stay byte-identical. On
unsharded dialects the helper renders the plain table. For OpenMeter the
sharding-key candidates are:
- `cityHash64(namespace)`: tenant co-location, allowing pruning with
  `optimize_skip_unused_shards`;
- `cityHash64(id)` or `rand()`: even spread, with fan-out for every query.

A skewed multi-tenant dataset is needed: an optional Zipf namespace
distribution in the generator. It is opt-in and appended after existing
draws, so default streams stay byte-identical.

**Questions, each answered by a measurement or a labelled assessment**

Replication

| ID | Question | Measurement |
| --- | --- | --- |
| R1 | How stale can reads be, and how often? Does `Distributed` fall back to stale replicas by default? | Per policy, see D7. Sources say it does: `fallback_to_stale_replicas_for_distributed_queries=1` and `max_replica_delay_for_distributed_queries=300` s. Upstream `load_balancing=random`; the operator profile sets `nearest_hostname`. Both profiles are measured. |
| R2 | What do quorum inserts cost? | Insert latency/throughput with `insert_quorum` = replica count vs the default |
| R3 | What happens when Keeper loses quorum? | Measured on kind and paid: time until tables go read-only, the error the OpenMeter-style writer sees, and recovery time. Single Keeper node loss is expected to have no impact; verified. |
| R4 | How long does replica loss and rebuild take? | Kill a replica pod (and, separately, its volume). Measure re-sync time to 100% of 100M rows and the latency impact on surviving replicas. |
| R5 | How much is storage amplified? | Bytes per replica measured: EBS has N copies; S3 disks without zero-copy also have N copies; Cloud has one shared copy |
| R6 | How does DDL propagate when a replica is down? | `ON CLUSTER` behaviour and DDL queue timeout vs the `Replicated` database engine |

Sharding

| ID | Question | Measurement |
| --- | --- | --- |
| S1 | Is the sharded result identical? | Digest parity for `2x2` vs `1x2` for every query is a gate |
| S2 | What does sharding buy for queries? | Per-tenant queries with shard pruning vs fan-out; `2x2` vs `1x4` (parallel replicas) vs Cloud at the same total vCPU (paid) |
| S3 | What happens when a shard is unavailable? | Default behaviour vs `skip_unavailable_shards=1`, which may return partial billable totals without an error; probe and experiment |
| S4 | How does insert routing affect visibility? | Time-to-queryable for: Distributed with background queue (default), Distributed foreground (`distributed_foreground_insert=1`), and client-side direct-to-shard inserts; plus queue behaviour while a shard is down |
| S5 | What does resharding cost? | Add a shard through the CHI. Measure operator schema creation, the absence of automatic rebalancing, and the time and cost of a rebalance by copy at 100M rows (measured on a smaller dataset locally, at 100M on paid if budget allows) |
| S6 | When must you shard? | Derived from the D15 rightsizing sweep: the largest single-shard shape that meets the SLO, and the data or ingest volume at which it stops |

**Reported as** a replication and sharding section in `REPORT.md` with a
self-managed topology recommendation for OpenMeter. Any correctness risk
measured here (partial results, stale fallback, decreasing totals) is a
design callout.

### D14. Compute/data separation (Cloud) vs colocation (self-managed) (owner addition)

**Configurations compared**

| Label | Target | Storage layout |
| --- | --- | --- |
| `separated` | `cloud` | SharedMergeTree on object storage, per-replica filesystem cache, plus the distributed cache if the service has it |
| `colocated` | self-managed `ebs` | ReplicatedMergeTree on gp3. Data local to compute; OS page cache. |
| `self-separated` | self-managed `s3` | ReplicatedMergeTree on an s3 disk with a local filesystem cache. Data remote, but metadata local per replica, so replicas are not stateless. Run twice: `cache_on_write_operations` off (upstream default) and on (matches Cloud's `enable_filesystem_cache_on_write_operations=1` default). |

Locality probes run before C1–C4 and decide how they are read:
- whether `system.filesystem_cache` and `FilesystemCacheSize` are readable on
  Cloud;
- which distributed-cache settings exist and whether they are enabled.

The sources disagree on whether Cloud replicas keep a local SSD cache (CS-02),
so this cannot be assumed.

**Questions**

| ID | Question | Measurement |
| --- | --- | --- |
| C1 | What do cold reads cost? | Per query class: latency and CPU cold vs warm. Cloud uses `SYSTEM DROP FILESYSTEM CACHE` + `enable_filesystem_cache=0`; EBS uses an OS page-cache drop on the node; s3 uses a filesystem-cache drop. Bytes read from object storage vs cache come from ProfileEvents. |
| C2 | How does latency depend on working set vs cache size? | Query-window sweep (1h/24h/7d/30d) at 100M rows. On self-managed `s3`, also cap the filesystem cache at 25/50/100% of the table's compressed size. Plot cache hit ratio vs p95. Cloud's cache size is not configurable, so it is recorded from `system.filesystem_cache*` where readable. |
| C3 | What changes on the write path? | Insert latency, time-to-queryable, merge throughput and merge CPU. Object-store request counts per 1M events (a cost line on self-managed `s3`; included in the price on Cloud). |
| C4 | How long do caches take to warm after a topology change? | First-query latency curve after replica add, restart, upgrade and resume-from-idle, until steady state. Shared with D15. On Cloud, also a sub-run with the Cloud-only prewarm levers (`cache_populated_by_fetch`, `prewarm_mark_cache`), reported separately from defaults. |
| C5 | How does storage cost scale with replicas? | Measured bytes × price: one shared copy (Cloud) vs N copies (EBS) vs N copies (S3 without zero-copy). |
| C6 | What is the failure blast radius? | Local disk loss: measured (R4). Object-store throttling or outage: assessed, not measured. |
| C7 | Does compute-compute separation isolate workloads? | Optional Cloud experiment if budget allows: query p95 under heavy ingest, same warehouse vs a separate read-only warehouse. Otherwise assessed with citations. |

**Reported as** a "compute/data separation" section. It states which
OpenMeter and OTel query classes are sensitive to separation (cold-read
penalty × how often they run cold), and what the `s3` self-managed mode does
and does not reproduce of Cloud's model.

### D15. On-demand scaling and rightsizing (owner addition)

**Compute scaling (measured unless labelled)**

| ID | Question | Measurement |
| --- | --- | --- |
| E1a | What happens during a Cloud scaling transition? | Cloud vertical autoscaling decides over 3 h and 30 h lookback windows (`docs/research/scaling-rightsizing.md`), so a minutes-long load step cannot trigger it. Instead, schedule a size change at a UTC hour with the beta `scalingSchedule` API and measure: transition time, make-before-break overlap (the replica count temporarily rises), query errors and latency during the change, cache warm-up on the new replicas, and billed compute units (usage statement). |
| E1b | How does Cloud vertical autoscaling react to sustained load? (optional, separate budget approval) | Hold a sustained heavy load for ≥ 4 h. Poll `CGroupMaxCPU`/memory per replica over time and record when autoscaling acts and what it bills. |
| E2 | How fast is Cloud horizontal scale-out? | `numReplicas` 2→3 via the API: time until the new replica serves queries, and its cold-cache latency curve (C4) |
| E3 | How fast does Cloud resume from idle? | Time to first result and latency of the first N queries after idle; the idle floor billed |
| E4 | How do scheduled scaling and scale-down behave on Cloud? | API probe of `scalingSchedule` and `replicaScaling`; assessed if the tier does not allow it |
| E5 | What does self-managed vertical scaling cost? | Change the CHI pod resources: rolling-restart duration, failed queries and inserts during the roll, and per-replica downtime |
| E6 | How long does adding a self-managed replica take? | `replicasCount` 2→3: operator schema creation, time to fetch 100M rows, and impact on existing replicas |
| E7 | What does adding a self-managed shard involve? | Same as S5 |
| E8 | Can self-managed autoscale? | Assessed: Kubernetes HPA does not fit stateful ClickHouse; only scheduled or manual scaling applies |

**Storage scaling**

| ID | Question | Measurement |
| --- | --- | --- |
| E9 | How does Cloud storage scale? | Automatic, billed per compressed TB-month (assessed and cited). Measured bytes per event drive the cost. |
| E10 | How does self-managed EBS storage scale? | Online PVC expansion through the operator (`storageManagement.provisioner: Operator`) and EBS CSI: time and disruption, measured. EBS allows 4 modifications per rolling 24 h and no shrinking; assessed and cited. |
| E11 | How much disk headroom do merges and mutations need? | Peak disk use during merges and during the String→JSON migration, as a fraction of steady-state data size (measured). This sets the EBS provisioning factor. |
| E12 | Is self-managed S3-backed storage elastic? | Measured bytes and requests; cited |

**Rightsizing**

| ID | Question | Measurement |
| --- | --- | --- |
| E13 | What is the smallest shape that meets the SLO? | Per workload, at a declared target load (ingest events/s + query mix QPS + p95 SLO per class, defined in `cells/`), sweep the replica size: Cloud 8/16/32 GiB per replica, self-managed 2/4/8 vCPU pods at the same 1:4 ratio. Report the smallest compliant shape per target. Cost comparisons in `REPORT.md` use each target's smallest compliant shape, not just identical shapes. |
| E14 | What does storage really cost? | Measured bytes per event per replica × retention (OpenMeter 13 months, OTel 30 days by default; parameters) × headroom factor (E11) → $/month per target and storage mode |
| E15 | Where do the costs cross? | A load-profile cost model: a parameterized diurnal/weekly curve. Cloud follows it using the autoscaling and idle behaviour measured in E1–E3; self-managed is provisioned for peak (plus optional scheduled scaling). Output: monthly cost per target and the crossover points (peak/average ratio, utilization, ops-hours) used by the verdict flip conditions. |

### D16. AI Cost Management (internal workload; owner addition)

**Scope.** The brief is self-contained: no access to the team's clusters, and an explicit output
format. It is answered with:
- **Local verification** against the exact builds in play:
  - Altinity 26.3.16 and 26.3.33;
  - upstream 26.3.33, 26.4.5, 26.7.4 and 26.8.2.
- **A local 3 × 2 cluster** that copies the proposed `prod-us` design:
  - two `Replicated` databases;
  - `Distributed` tables;
  - an interserver secret and the `SQL_` prefix;
  - replicated access storage and replicated UDFs.
- **Deterministic synthetic data** containing redeliveries, changed-time retries and late events.
- **Read-only measurement statements** for the team's own Cloud and `prod-us` environments.

**Where things live.** Everything sits under `internal/ai-cost-management/`. The OpenMeter and OTel
workloads stay public-safe.

**Overlap with D13.** The AICM experiments are a concrete instance of D13's questions:
- R1: stale fallback and replica routing;
- S1: parity;
- S2: pruning;
- S3 and S4: insert routing;
- R6: DDL with a replica down.

The public D13 implementation reuses the AICM scripts' techniques (`SYSTEM STOP FETCHES` lag
injection, the count-based visibility check, and `parallel_distributed_insert_select`).

## Risks / Trade-offs

**Fairness and comparability**
- **Cloud hardware is unknown and noisy.** The CPU model and neighbours are
  invisible. → Mitigation: repeat runs (`--repeat ≥3`) on separate days,
  report variance, and compare the same instance family on self-managed.
- **Version skew.** Cloud 26.4/26.6 vs 26.3. → Mitigation: the
  `upstream-cloudver` control, with skew-explained results labelled.
- **Cache effects and idle scaling on Cloud.** → Mitigation: idle scaling is
  disabled for perf runs and enabled only in the idle-floor COGS cell;
  warm/cold are measured explicitly.

**Tooling and environment**
- **Kind needs sandbox patches** (cgroup v1, oom_score_adj). → Mitigation:
  patches live in `targets/kind/` behind detection, and CI uses stock kind.
- **No MinIO image.** `minio/minio` is no longer pullable. → Mitigation: pick
  another S3 emulator (for example SeaweedFS or LocalStack); this is a Phase 2
  task.
- **The docs site is unreachable from the sandbox.** → Mitigation: cite
  published URLs, verify against the docs source at a pinned commit, and let
  the owner spot-check.

**Cost**
- **Paid spend.** → Mitigation: per-run estimates before every paid run, a hard
  budget, run-id tags, teardown that is verified by API listing, and a
  `targets/*/down --verify` step that fails loudly.

**Scope and spend from the owner additions (D13–D15)**
- **More shapes, more spend.** Topology sweeps (`1x2`/`1x4`/`2x2`), shape
  sweeps (3 sizes × 2 platforms) and scaling experiments multiply paid hours.
  → Mitigation:
  - Sweeps run on the OpenMeter workload first, at 10M rows. OTel workloads
    reuse the shapes it identifies.
  - Each sweep is estimated and approved separately.
- **Cloud autoscaling cannot be controlled or observed internally.**
  → Mitigation: treat it as a black box. Measure the externally visible signals
  (per-replica cgroup limits over time, latency, errors, billed units), and
  label the internals as assessed.
- **Sharding needs non-default data.** → Mitigation: the opt-in Zipf
  namespace distribution is appended after existing generator draws, so
  default streams and existing digests stay unchanged.

**Scope**
- **OTel logs were added mid-flight.** → Mitigation: logs reuse the traces
  infrastructure and generators, and are planned as Phase 5b.

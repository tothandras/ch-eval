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

**Non-Goals:**
- **Drawing performance or cost conclusions from local tiers.** Compose and
  kind are for correctness, parity and smoke tests only, and the report
  generator enforces this.
- **Sharded topologies.** All perf topologies are 1 shard × N replicas. The
  user research's sharding risks are discussed, not measured, and labelled
  "assessed, not measured".
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

**Scope**
- **OTel logs were added mid-flight.** → Mitigation: logs reuse the traces
  infrastructure and generators, and are planned as Phase 5b.

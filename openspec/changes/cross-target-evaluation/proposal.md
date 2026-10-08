## Why

The forked harness answers *which `om_events` table design is cheapest* on a
single target. OpenMeter now also needs to choose *where* ClickHouse runs:
ClickHouse Cloud or a self-managed Altinity Stable cluster. It also needs to
know whether that answer holds for the observability data it plans to store:
OTel traces, metrics and logs.

Today nothing measures how the two platforms compare on capability,
performance, cost and operational burden. Several platform differences are not
only performance questions; they change OpenMeter's design:

- **Engine and DDL.** SharedMergeTree on Cloud vs ReplicatedMergeTree + Keeper
  on self-managed. OpenMeter hard-codes `ENGINE = MergeTree` and has no
  ClickHouse migration mechanism.
- **Read consistency across replicas.** This matters for billing snapshots and
  entitlement enforcement (see `docs/research/sharedmergetree-usage-collection.md`).
- **Ingest defaults.** Both changed in 26.2: `async_insert=1` and
  deduplicate-all-inserts.

This change makes the **target platform** a first-class benchmark dimension,
next to the existing table-design dimension. Every claim in the final report
is either measured or explicitly labelled "assessed, not measured".

## What Changes

**Targets and harness**
- **Target dimension.** `targets/<name>/` holds a manifest, a single `up`/`down`
  entry point per topology tier, and a tested teardown. Targets are `cloud`,
  `altinity` and `upstream-oss`. `upstream-cloudver` is a proposed control:
  upstream at Cloud's version on self-managed topology, which separates
  version effects from architecture effects. Tiers are `local-single`
  (compose), `local-cluster` (compose with 2 replicas + 3 Keeper + S3
  emulator), `kind` (Altinity operator), `aws` (EKS + Altinity operator, EBS
  gp3 or S3-backed MergeTree) and `cloud` (ClickHouse Cloud API).
- **Harness.** The harness gains `--target`, captures run metadata (versions,
  build, topology, changed server/session/MergeTree settings, shapes, seed,
  client location) and writes results to
  `results/<workload>/<scenario>/<target>/<run-id>/`. `bench compare` gains a
  cross-target mode that gates on value-digest parity per workload and
  scenario. Runner, digest and COGS attribution stay unchanged (ADR-0001).

**Migrations**
- **One templated source** under `migrations/` and `scenarios/*/*/init/`, using
  golang-migrate-compatible file naming as OpenMeter does for Postgres.
- **Dialects:** `single`, `replicated` (ON CLUSTER + ReplicatedMergeTree with
  Keeper paths) and `cloud` (SharedMergeTree; the probe decides how ON
  CLUSTER is handled).
- **Behaviour:** a migrations table, idempotency, ordered apply, up/down/re-apply
  tests per target, and per-migration duration and resource capture for
  online changes (ADR-0003).

**Workloads**
- `scenarios/<workload>/<variant>/`:
  - **`openmeter`:** `baseline-openmeter` and `proposal`, schema and queries
    unchanged.
  - **`otel-traces`, `otel-metrics`, `otel-logs`:** each with an
    `exporter-default`, a `clickstack` and a `tuned` variant. OTel logs were
    added to scope by the repo owner on 2026-10-08.
  - **`replica-consistency`:** a new experiment that measures stale reads,
    visibility lag and cumulative-total monotonicity across replicas under each
    consistency policy.
- **Generators:** deterministic, seeded generators for OTel traces, metrics and
  logs under `generators/`, with tunable cardinality. Two output paths: direct
  native insert at scale, and OTLP into an OTel Collector (contrib 0.162.0)
  with the ClickHouse exporter (`create_schema: false`). The existing
  OpenMeter seeder is reused unchanged.

**Owner additions (2026-10-08): topology, locality, elasticity**
- **Replication and sharding on self-managed (design D13).**
  - Topologies `1x2`, `1x4` and `2x2`.
  - Questions: stale reads and stale-replica fallback, quorum cost, Keeper
    quorum loss, replica rebuild, storage amplification and DDL propagation.
  - For sharding: digest parity of sharded vs unsharded, shard pruning by
    namespace, partial results when a shard is unavailable, Distributed insert
    queue visibility, resharding without automatic rebalancing, and when
    sharding becomes necessary.
  - A `shardedTable` migration helper keeps query files unchanged.
- **Compute/data separation vs colocation (design D14).**
  - Cloud's separated storage vs self-managed EBS (colocated) and S3-backed
    MergeTree (self-separated).
  - Measured: cold vs warm reads, working set vs cache size, write path and
    merges, warm-up after topology changes, storage copies × price, and
    optionally warehouse isolation.
- **On-demand scaling and rightsizing (design D15).**
  - Cloud: vertical autoscaling step response, horizontal scale-out,
    idle/resume, scheduled scaling.
  - Self-managed: vertical (rolling restart), add replica, add shard.
  - Storage: Cloud automatic; EBS online expansion, merge and mutation
    headroom; S3-backed.
  - Rightsizing: the smallest SLO-compliant shape per workload and target, and
    a load-profile cost model that computes crossover points.

**Owner addition (2026-10-08): AI Cost Management (internal workload)**
- **Workload:** AI Cost Management usage analytics. Compares ClickHouse Cloud (one logical shard) with a shared Altinity 3 × 2 `prod-us` cluster.
- **Brief and deliverable:** under `internal/ai-cost-management/`. The directory is internal and is removed before any publication.
- **Local verification, against real builds:**
  - feature acceptance on 26.3.16 through 26.8.2;
  - a 3 × 2 cluster copy of the proposed design;
  - tenant isolation across shards;
  - the notifier's read-after-write check under injected replication lag;
  - sharding-key alternatives;
  - mutations;
  - export memory.
- **Paid-tier follow-ups:** the deliverable's measurement requests, run by the owning team or in Phase 3+ if the owner wants this repo to run them.

**Feature matrix and cost**
- **Feature probes.** `bench probe` runs SQL and Go probes per target.
  `docs/feature-matrix.md` is generated from `docs/feature-matrix.yaml` plus
  probe results, and every row carries a primary-source citation.
- **COGS extension.** Self-managed pricing profiles cover compute, EBS, S3
  storage and requests, Keeper nodes, EKS control plane, egress and a
  parameterized ops-hours line. COGS reports $/1M events, $/1k queries by
  class, $/GB-month and the idle floor at matched utilization, plus where the
  cost curves cross. Reconciliation runs against the Cloud usage statement and
  AWS Cost Explorer (run-id tags).

**Other aspects and reporting**
- **Other aspects, measured where possible:**
  - Concurrency: max QPS at a p95 SLO with ingest running.
  - Ingest: events/s, merge CPU share, parts over time, time-to-queryable.
  - Cold vs warm reads.
  - Reliability: replica loss, Keeper loss, rolling upgrade.
  - Backup/restore and cross-target portability.
  - Security and multi-tenancy probes.
  - `clickhouse-go` v2 compatibility.
  Operational burden is a qualitative write-up with evidence.
- **`REPORT.md`** is generated by `bench report` from `results/`, never
  hand-typed. Results from local tiers are refused for performance and cost
  conclusions.
- **CI:** lint, unit tests, `openspec validate`, migration tests and a
  small-row-count smoke run against `altinity` and `upstream-oss` on every PR.

## Capabilities

### New Capabilities
- `target-platforms`: target manifests, per-tier `up`/`down`, run-metadata capture, cost guardrails (tagging, verified teardown).
- `schema-migrations`: templated, dialect-aware, idempotent, recorded migrations with per-migration cost capture.
- `otel-workloads`: OTel traces/metrics/logs scenario variants, seeded generators, and the collector ingest path.
- `replica-consistency`: the cross-replica visibility and monotonicity experiment.
- `cross-target-comparison`: results layout per target, cross-target compare, and the digest parity gate.
- `feature-probes`: probe format, probe runner, and the generated feature matrix.
- `evaluation-report`: generated `REPORT.md` with per-workload verdicts and flip conditions.
- `replication-sharding`: self-managed replication and sharding topologies, migration templating for sharded tables, and the R1–R6 / S1–S6 experiments.
- `storage-locality`: compute/data separation (Cloud) vs colocation (self-managed EBS) vs self-managed S3-backed storage, experiments C1–C7.
- `elastic-scaling`: on-demand compute and storage scaling experiments (E1–E12) and rightsizing with load-profile cost modelling (E13–E15).
- `ai-cost-management-workload`: the internal AI Cost Management evaluation (brief, local 3 × 2 verification, report in the brief's output format).

### Modified Capabilities
- `scenario-format`: scenarios nest under a workload (`scenarios/<workload>/<variant>/`). `init.sql` becomes an `init/` migration set. Per-workload parameter providers replace the fixed v1 parameter set for non-OpenMeter workloads.
- `benchmark-harness`: `--target` flag; results path `results/<workload>/<scenario>/<target>/<run-id>/`; run metadata captured on every run.
- `cogs-pricing`: a profile `kind` of `cloud` or `self-managed` with self-managed line items.
- `cogs-resource-accounting`: the cluster name for `clusterAllReplicas` comes from the target instead of a hard-coded `default`; S3 request counters are collected.

## Impact

**Code**
- **Additive:** `bench/` gains `target/`, `migrate/`, `probe/`, `report/` and `consistency/` packages.
- **Small, targeted edits:** `runner.Discover` (workload nesting), `runner.Write` (path), `accounting.DetectTarget` (cluster name), `pricing` (profile kind).
- **Unchanged:** `runner/digest.go`, `cogs/attribution.go`, `seed/`.

**Dependencies**
- New: the OTel Collector contrib image, the Altinity operator (0.27.4), and an S3 emulator for local S3-backed tests.
- New: Terraform (AWS + `ClickHouse/clickhouse` provider 3.37.0) for paid tiers.

**Paid resources**
- ClickHouse Cloud and AWS (EKS/EC2/EBS/S3) are created only after explicit owner approval and a stated budget.
- Every resource is tagged with the run ID and has a tested, verified teardown.

**Repo hygiene**
- Synthetic data only. Secrets only via environment or a git-ignored `.envrc`.

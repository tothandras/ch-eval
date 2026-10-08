# PROGRESS

Durable log for this evaluation. **Every session starts by reading this file
and `openspec/changes/cross-target-evaluation/` (proposal, design, tasks).**
Keep it accurate enough that a fresh agent can continue.

## Current state

- **Phase:** 2 (local targets, migrations, generators, parity) is complete,
  apart from the kind 2×2 tier (see tasks.md 2.13). Work stops here, before
  Phase 3: no paid resource is created without the owner's approval and
  budget. The Phase 1 questions below are still open. Phase 2 used the
  documented defaults.
- **Owner decision pending:** whether to report the quorum + async-insert
  LOGICAL_ERROR upstream
  (`results/experiments/repro/quorum-async-logical-error-20261008.md`).
- **AI Cost Management (internal workload):** brief received 2026-10-08,
  locally verified, deliverable written:
  `internal/ai-cost-management/REPORT.md`. Waiting for owner review.
- **Publication blocker:** `internal/` holds internal design material. Remove
  or scrub it before the repository is made public. The repository is private
  as of 2026-10-08.
- **Branch:** `claude/lucid-edison-5q68bh`.
- **Plan of record:** OpenSpec change `cross-target-evaluation`
  (`openspec validate --strict` passes).

## Phase log

### Phase 1: read, plan, draft feature matrix (2026-10-08)

**Done**
- **Harness forked.** Forked `tothandras/clickhouse-benchmarks@542b1e5`
  verbatim; only the module path was renamed (ADR-0001). `go build`,
  `go vet` and `go test ./...` pass on Go 1.26.2.
- **Research** (summaries plus verbatim sources in `docs/research/`):
  - OpenMeter ClickHouse DDL: `openmeterio/openmeter@5a0e27e6` has a single
    `om_events`, `CREATE IF NOT EXISTS` at startup, no migrations and no
    cluster support. Postgres uses golang-migrate naming with a `schema_om`
    version table.
  - OTel DDL: exporter v0.162.0 and ClickStack `hyperdx@eeabaebe` for traces,
    logs and metrics. 27 verbatim files in `docs/research/otel-ddl/`.
  - Versions: Altinity Stable `26.3.33.10001.altinitystable`, upstream
    `26.3.33.24` (same base), Cloud 26.6 (Fast/Regular) or 26.4 (Slow),
    operator 0.27.4, collector 0.162.0, kind 0.33.0.
  - The owner's SharedMergeTree / usage-collection analysis is archived
    verbatim as hypotheses
    (`docs/research/sharedmergetree-usage-collection.md`). It is turned into
    the measured `replica-consistency` experiment (design D7).
- **Local feasibility** (smoke only; no performance meaning):
  - Docker works once `dockerd` is started. All pinned images pull except
    `minio/minio`.
  - kind 0.33 works with two sandbox patches (see Environment notes).
  - The unchanged harness ran `baseline-openmeter` and `proposal` at 200k rows
    on single-node upstream 26.3.33.24 and Altinity 26.3.33.10001 with no
    errors. Value digests matched across targets (30/30 and 34/34) and across
    scenarios (30/30).
  - Observed ad hoc: `async_insert` defaults to `true` on both 26.3 images.
- **Draft feature matrix:** `docs/feature-matrix.md`, about 50 rows, each
  with a citation and a planned probe, all marked as hypotheses.
- **Plan and spec docs:** OpenSpec proposal, design (D1–D12), delta specs
  (7 new and 4 modified capabilities), tasks; ADRs 0001–0005;
  `docs/methodology.md` draft.

**Findings so far** (cited, not yet probed; design-callout candidates)

1. **Async and dedup defaults changed under OpenMeter.** Since 26.2,
   `async_insert=1`, `wait_for_async_insert=1` and deduplicate-all-inserts are
   defaults. That covers every target in this evaluation. OpenMeter's sink
   never sets `async_insert`, so its inserts become async on all targets. Its
   retries regenerate `store_row_id`, so block-hash dedup never catches them.
   Plain `MergeTree` deduplicates nothing.
2. **Engine choice is a correctness question on self-managed.** OpenMeter
   hard-codes `ENGINE = MergeTree`. Cloud converts it to SharedMergeTree; on a
   self-managed replicated cluster it creates independent, unreplicated tables
   per replica. Any self-managed HA deployment therefore needs DDL changes.
3. **Executable UDFs are GA on Cloud**, which refutes the starting hypothesis
   that they are self-managed only.
4. **Cloud uses the Shared database engine**, not `Replicated`.
5. **Lightweight `UPDATE` is still beta** in 26.x. Lightweight mutations fail
   on tables with projections.

**Decisions**

| Decision | Recorded in |
| --- | --- |
| Fork, don't rewrite | ADR-0001 |
| Target manifest plus live metadata; results under `results/<workload>/<scenario>/<target>/<run-id>/` | ADR-0002 |
| Migrations: golang-migrate naming plus a custom templated runner | ADR-0003 |
| OTel: verbatim exporter and ClickStack schemas plus one evidence-justified tuned variant; logs in scope | ADR-0004 |
| Local tiers are correctness-only; paid tier uses matched shapes (proposal) | ADR-0005 |

**Blockers or risks**
- clickhouse.com, docs.altinity.com and builds.clickhouse.com are unreachable
  from the sandbox. Docs are cited by URL and verified against the docs source
  on GitHub. `clickhouse-benchmark` is extracted from the server image instead.
- `minio/minio` cannot be pulled. Phase 2 needs another S3 emulator, for
  example SeaweedFS or LocalStack.

### Phase 1 addendum: owner additions (2026-10-08)

The owner added three areas after the first Phase 1 summary. Each is now a
design decision with measured experiments, plus feature-matrix sections and
research notes:

| Area | Design | Feature-matrix section | Research note |
| --- | --- | --- | --- |
| Replication and sharding on self-managed | D13 | §9, RS-01…14 | `docs/research/replication-sharding.md` |
| Cloud compute/data separation vs colocation | D14 | §10, CS-01…09 | `docs/research/compute-storage-separation.md` |
| On-demand scaling and rightsizing for compute and storage | D15 | §11, SC-01…07 | `docs/research/scaling-rightsizing.md` |

New delta specs: `replication-sharding`, `storage-locality`,
`elastic-scaling`. ADR-0005 is amended with the topologies `1x2`/`1x4`/`2x2`,
the two s3 cache-on-write variants, and the rightsizing sweep.

**New source findings** (cited, not probed; design-callout candidates)

6. **Stale reads are on by default on self-managed.** Distributed reads fall
   back to stale replicas
   (`fallback_to_stale_replicas_for_distributed_queries=1`, with
   `max_replica_delay_for_distributed_queries=300` s). Upstream
   `load_balancing=random` lets successive reads differ. The operator's
   profile switches to `nearest_hostname` but leaves the stale fallback on.
   Read-after-write guarantees on ReplicatedMergeTree need `insert_quorum`
   **and** `insert_quorum_parallel=0` (whose default is the opposite) plus
   `select_sequential_consistency=1`.
7. **Sharding risks.**
   - The default Distributed insert is a local disk queue sent in the
     background, and can be lost on a rough restart.
   - `skip_unavailable_shards=1` returns partial totals without an error. The
     default (0) is safe.
   - There is no automatic rebalancing, and clickhouse-copier is unsupported.
   - Upstream ("largest server first") and Altinity ("start sharded") disagree
     on when to shard.
8. **Sources conflict on Cloud's local cache.** The docs describe a
   per-replica SSD filesystem cache; a Cloud blog says nodes no longer cache on
   local disks. Probe CS-02 settles it. Cloud fills the cache on writes by
   default and self-managed does not, so the `s3` mode runs both ways. Cloud
   does not bill object-storage requests or intra-region transfer; self-managed
   S3 pays per request.
9. **Scaling.**
   - Cloud vertical autoscaling uses 3 h and 30 h lookback windows, so short
     peaks are followed only by scheduled scaling (beta API) or idling (≥ 5 min
     timeout, wake in about tens of seconds).
   - Make-before-break temporarily adds replicas.
   - Horizontal autoscaling is private preview.
   - Self-managed: vertical changes roll host by host. A new replica copies the
     whole dataset. EBS allows 4 modifications per rolling 24 h (the "6-hour
     cooldown" is out of date) and cannot shrink. Merges reserve 2× the merged
     size on disk.

**Experiment design changes this forced**
- **E1 is split:** a scheduled-scaling transition (E1a) and an optional ≥ 4 h
  sustained-load autoscaling run (E1b).
- **C4 adds a Cloud prewarm sub-run.**
- **Locality probes run before C1–C4.**

### AI Cost Management workload (2026-10-08, owner addition)

**What was done.** All work was local and correctness-only. Details:
`internal/ai-cost-management/README.md`.
- **Brief:** stored, with its SQL verbatim. One out-of-scope item is omitted
  from the copy.
- **Builds tested:** Altinity 26.3.16 and 26.3.33; upstream 26.3.33, 26.4.5,
  26.7.4 and 26.8.2.
- **Local 3 × 2 cluster:** two `Replicated` databases, `Distributed` tables,
  an interserver secret, the `SQL_` prefix, and replicated access storage and
  UDFs.
- **Data:** 2M deterministic synthetic events, plus redeliveries and
  changed-time retries.
- **Experiments:** listed in tasks §8.
- **Deliverable:** `internal/ai-cost-management/REPORT.md` in the brief's
  format, with each claim labelled verified, inferred or assumed and an
  evidence index.

**Headline results (verified locally)**

| # | Result |
| --- | --- |
| 1 | D-13's max-offset check passes stale under lag: 108/300 and 194/300 rows visible. A count-based same-statement check never did. |
| 2 | Altinity 26.3.16 lets the reader bypass row policies via `mergeTreeProjection` (all tenants) and `mergeTreeIndex`. Fixed in 26.3.33. |
| 3 | The projections are unused on every build and cost about +20–34% insert CPU and +25% merge CPU. |
| 4 | Changed-time retries are double-counted under `FINAL`. |
| 5 | `FINAL` is not partition-local by default. |
| 6 | The D2 monthly export can exceed 512 MiB, and the reader cannot enable spilling. |
| 7 | Profile `MAX` constraints do not bound defaults. |
| 8 | `parallel_distributed_insert_select=2` ignores the destination sharding key. |
| 9 | A replica outage gives DDL error 159 with the change still applied. |
| 10 | Tenant sharding gives single-shard reads with identical results, but skews data: 59% on one shard. |

**Local cluster note.** The cluster keeps running between turns while the
container lives. Recreate it from the README.

### Phase 2: local targets, migrations, generators, parity (2026-10-08)

Everything here ran on local tiers: correctness, parity and smoke only, never
performance (ADR-0005).

**Built.**
- **Migration tool** `migrations/migrate`:
  - dialect-aware templates (single, replicated, sharded, cloud) and lint;
  - state in `schema_migrations_ch`, with dirty and pending (error 159) states;
  - mutation waits and cost capture.
  - `bench migrate status|up|down|verify|force|render`;
    `TestMigrationLifecycle` passes on compose for both builds.
- **Targets.**
  - compose for `altinity` and `upstream-oss`: single, 1×2 (+3 Keeper +
    SeaweedFS S3), 1×2 on the s3 policy, 2×2, and s3 cache variants.
  - kind with operator 0.27.4 (see below).
  - A cloud stub that refuses to start without approval.
  - No connection string or secret is stored in the repository.
- **Harness.**
  - `bench --target --workload`; a fresh down+up migration per run;
    explicit `async_insert` on every insert.
  - Cluster-aware CPU probe; `metadata.json` with a drift gate.
  - Results go to `results/<workload>/<scenario>/<platform>/<tier>/<run-id>/`.
  - `bench compare` checks cross-target and cross-scenario value parity and
    verifies identical seeded inputs. `--runs` compares explicit runs, for
    ingest-path parity. COGS is adapted to the new layout.
- **OTel.**
  - Deterministic generators in `generators/otel` for traces, logs and five
    metric types.
  - A direct mapper that reproduces the contrib exporter v0.162.0.
  - OTLP/HTTP JSON through a pinned collector (`targets/collector`).
  - Six scenarios, ported verbatim by `migrations/tools/port_otel_ddl.py`.
  - Full-table fingerprints in `parity/`.
- **Probes and matrix.** `bench probe` runs 49 probes (44 SQL + 5 Go).
  `docs/feature-matrix.yaml` (85 rows) feeds the generated
  `docs/feature-matrix.md`.
- **Experiments and tooling.**
  - `bench consistency` (D7).
  - `bench experiment S3|S4|R6|R3` (D13 local correctness).
  - Zipf tenant skew (`--namespace-skew`).
  - CI in `.github/workflows/ci.yml`.

**Evidence (local, all from clean commits in an evidence worktree):**
- **Parity.** Value digests and full-table fingerprints match across targets
  for OpenMeter (both scenarios) and all six OTel scenarios:
  - altinity vs upstream-oss on single and 1×2 (`results/parity/local-*.md`);
  - altinity single vs 2×2 vs kind (`results/parity/local-sharded-kind-*.md`);
  - OpenMeter also on 1×2-s3.
- **OTel ingest paths.** The collector path and the direct mapper store
  identical data in every table of all six scenarios. The two schema variants
  agree on every shared query.
- **Probes.** 50 probes (45 SQL + 5 Go) on altinity single, 1×2, 2×2 and kind,
  and on upstream single and 1×2. Every expectation was confirmed, none
  contradicted (`results/probes/`, `docs/feature-matrix.md`). Notable
  confirmations:
  - async insert and dedup defaults;
  - plain MergeTree on a cluster is unreplicated;
  - the operator profile sets `load_balancing=nearest_hostname` and
    `prefer_localhost_replica=0`;
  - `skip_unavailable_shards=1` returns partial data silently.
- **Replica consistency** (`results/consistency/`). This is local shape only;
  the numbers are not evidence.

  | Policy | Behaviour |
  | --- | --- |
  | Default | about 4–6% of load-balanced reads stale; some totals go backwards; lag p99 about 0.1 s, several seconds across a replica restart |
  | Quorum with sync inserts | no stale reads, at about 4× the insert p50 |
  | `SYNC REPLICA … LIGHTWEIGHT` before each read | stale reads about 0 |

  **Quorum with async inserts (the 26.2+ default) is rejected.**
- **Experiments** (`results/experiments/`):

  | ID | Outcome |
  | --- | --- |
  | S3 | default query fails (279); `skip_unavailable_shards=1` silently returned 50,088 of 100,000 rows |
  | S4 | background Distributed inserts acknowledge before rows are queryable and queue while a shard is down; foreground inserts fail |
  | R6 | `ON CLUSTER` and Replicated-database DDL both time out (159) with a replica down and apply on its return |
  | R3 | quorum loss makes tables read-only (242) while SELECT works; inserts recover in 6–9 s |
  | E6 (kind) | raising `replicasCount` from 2 to 3: the operator registered the host, created the table on it with no DDL from us, and the replica fetched 1M rows, all about 83 s after the patch |
  | E5 (kind) | a pod memory-limit change rolled the 3 hosts one at a time; 140 s to `Completed` |
- **kind (operator 0.27.4).** Both builds passed the migration lifecycle, all
  50 probes, a consistency smoke (the same shape as compose: default about 5%
  stale, quorum and `SYNC REPLICA` 0) and value parity against their single
  tier for every workload and scenario
  (`results/parity/local-sharded-kind-*.md`, `local-kind-upstream-*.md`).
- **s3 storage variants.** `local-cluster-s3-cow` (write-through caching,
  Cloud's documented default) and `local-cluster-s3-smallcache` (256 MiB)
  passed parity and all 50 probes. The cow tier is being re-verified after the
  write-through fix below. Each run records the effective cache
  settings and the cache occupancy after seeding (D14 mechanics, not
  performance). After 500k OpenMeter rows on 1×2, the read-only tiers held
  1,200 segments and 50.7 MB, filled by merge and fetch reads. Write-through
  held 1,922 segments and 210.7 MB.
- **CI.** GitHub Actions is green, including the Docker smoke job with the
  parity gate and collector-path parity.

**Findings along the way** (no performance claims):
- 26.2+ defaults `async_insert=1`, so the harness must set it explicitly.
  Otherwise the server default would decide the insert path on one target only.
- Sharding by `cityHash64(namespace)` puts a single-tenant dataset on one
  shard. Parity on a sharded target therefore also needs multi-namespace data.
- The Altinity operator rejects cluster names with `_`
  (`^[a-zA-Z0-9-]{0,15}$`), so kind tiers use `ch-eval`.
- The operator bundle pulls with `imagePullPolicy: Always`. Offline kind nodes
  need a local `IfNotPresent` copy, and that is the only deviation.
- The forked harness counted untracked `results/` as dirty. This is fixed.
- Quorum inserts are rejected with async inserts. Through SQL `VALUES` the error
  is 2. When the data travels as a block (any `FORMAT` clause, and
  every clickhouse-go batch insert), both 26.3.33 builds raise
  **LOGICAL_ERROR 49** ("Should be checked earlier"), so the validation is
  bypassed on that path. A minimal reproduction is in
  `experiments/repro/quorum-async-logical-error.sh`, with output in
  `results/experiments/repro/`. It has not been reported upstream; the owner
  decides.
- `kind load docker-image` fails on single-platform pulls (`--all-platforms`
  digest error), so `targets/kind/up` imports with `docker save | ctr import`.
  remote_servers can lag a CHI that has reached `Completed`, so `up` also waits
  until `system.clusters` lists every host.
- **Write-through caching needs two knobs.** The cache disk's
  `cache_on_write_operations=1` does nothing unless the setting
  `enable_filesystem_cache_on_write_operations` is also 1. That setting
  defaults to 0 upstream and is documented as 1 on Cloud. The first s3-cow
  tier set only the disk knob, and its cache after seeding matched the
  read-only tiers byte for byte. The `cache_after_seed` snapshot exposed this,
  and the tier now sets both. This is a fairness trap for D14: a self-managed
  "cache on write" configuration can silently be off.
- Local compose tiers filled the sandbox disk through `information`-level
  Docker logs. They now log at `warning`, rotate logs, and keep only the system
  logs the harness reads.

## Questions for the owner (Phase 1 stop point)

1. **Budget:** what is the cap for paid resources, and which AWS account or
   profile and ClickHouse Cloud organization should be used? The revised
   estimate is below. Every paid run gets its own estimate and approval
   before anything is created.

   | Block | Basis | Estimate |
   | --- | --- | ---: |
   | Self-managed `1x2` | ≈ 97 h × $1.77/h | ≈ $175 |
   | Self-managed 4-node topology sweep | ≈ 10 h × $2.68/h | ≈ $27 |
   | Cloud, 2 × 16 GiB plus sweeps and scaling | ≈ 55 h | ≈ $90–110 |
   | Storage, S3, egress, backups | — | ≈ $20–40 |
   | Idle and debug overhead | +40% | — |
   | **Base total** | | **≈ $450–600** |
   | Optional: `upstream-cloudver` | — | +≈ $45 |
   | Optional: E1b (≥ 4 h autoscaling) | — | +≈ $10 |
   | Optional: C7 warehouse isolation | — | +≈ $8 |

   How the hourly figures were built (us-east-1):
   - **Self-managed `1x2`, about $1.77/h:**
     - 2 × m7i.2xlarge ClickHouse nodes at $0.4032/h each;
     - 3 × m7i.large Keeper nodes;
     - EKS at $0.10/h;
     - an m7i.2xlarge client VM;
     - gp3 at $0.08/GB-month, with prices from the AWS Price List on
       2026-10-07.
   - **4-node topologies:** about $2.68/h.
   - **Cloud, 2 × 16 GiB:** about $1.19/h at $0.2985 per compute-unit hour.
     That rate is **unverified for today** and will be re-checked.
   - **eu-central-1** is about 20% more.
   - **Suggested cap:** about $700.
   - **Bigger reductions** are available by cutting the OTel paid runs to one
     build, or by skipping the `s3` cache-on-write variant.
2. **Region and shape:** proposed is one AWS region (eu-central-1 to match the
   earlier COGS run, or us-east-1), with 2 replicas × 16 GiB / 4 vCPU on
   Cloud's Scale tier and the same on EKS with the Altinity operator. Is 2 or 3
   replicas preferred?
3. **`upstream-cloudver` control:** add upstream at Cloud's version (26.6) on
   the self-managed topology to separate version skew from architecture? It
   adds roughly 25% to self-managed spend.
4. **Self-managed on EKS with the operator** for paid runs, as the brief says,
   rather than cheaper EC2 + Docker. Confirm.
5. **Scenario comment:** `scenarios/proposal/init.sql` mentions an internal
   data-copy row count. It is already public in the source repo. Should it stay
   byte-identical or be scrubbed?
6. **Replica-consistency scope:** is design D7 (stale reads, lag,
   monotonicity, policy overhead, RMT vs SMT) the right test of the
   SharedMergeTree analysis? Are there OpenMeter query shapes to add, such as
   the charge snapshot query with `stored_at < StoredAtLT`?
7. **Sharding on paid tiers:** is the paid `2x2` topology needed at 10M rows
   for query behaviour only, or also at 100M for resharding cost (S5)? I
   propose 10M paid plus resharding measured on a smaller dataset; 100M only
   if the budget allows.
8. **Target load and SLOs for rightsizing (E13):** what ingest rate and query
   load should the rightsizing sweep target? I propose the existing COGS mixed
   cell (5k events/s + 4 qps) as "typical", and 25k events/s + 16 qps as
   "peak". Proposed p95 SLOs: meter_agg ≤ 500 ms, payload_heavy ≤ 1 s,
   lookup ≤ 200 ms. These are placeholders; real OpenMeter SLOs would be
   better.
9. **Retention for storage rightsizing (E14):** OpenMeter events 13 months
   and OTel 30 days by default? Override if OpenMeter has real retention
   policies.
10. **AI Cost Management:** review `internal/ai-cost-management/REPORT.md`.
    Should this repository run any of its measurement requests on paid tiers?
    It cannot reach the team's clusters itself. Should `internal/` stay in
    this repository at all?

## Next

- **Phase 2 leftover:** kind 2×2 (tasks.md 2.13) did not fit the sandbox's
  memory. Compose 2×2 already covers sharded DDL and parity; run kind 2×2 on a
  larger host if the operator's sharded path itself needs evidence.
- **Phase 3 (paid) needs the owner's go-ahead:** budget, region, shape, and
  the open questions above. A spend estimate is presented before anything is
  created.

## Environment notes (for resuming in a fresh container)

**Go.** The toolchain is 1.24 by default and `go.mod` needs 1.26.2:
`export GOTOOLCHAIN=go1.26.2`. It auto-downloads through the Go proxy.

**Docker.** The daemon is not running at container start:
```bash
nohup dockerd > /tmp/dockerd.log 2>&1 &
```
The sandbox forbids raising `nofile` ulimits, so do not pass
`--ulimit nofile=...`.

**clickhouse-benchmark / client.**
```bash
cid=$(docker create clickhouse/clickhouse-server:26.3.33.24)
docker cp $cid:/usr/bin/clickhouse /home/user/tools/ch/clickhouse
docker rm $cid
ln -s clickhouse /home/user/tools/ch/clickhouse-benchmark
ln -s clickhouse /home/user/tools/ch/clickhouse-client
```
Then add `/home/user/tools/ch` to `PATH`.

**kind and kubectl.**
- Install: `GOBIN=/home/user/tools/bin go install sigs.k8s.io/kind@latest`.
- kubectl comes from `dl.k8s.io`.
- The sandbox is cgroup v1 without CAP_SYS_RESOURCE, so the kind config needs
  both patches:
  ```yaml
  containerdConfigPatches:
  - |-
    [plugins."io.containerd.grpc.v1.cri"]
      restrict_oom_score_adj = true
  nodes:
  - role: control-plane
    kubeadmConfigPatches:
    - |
      kind: KubeletConfiguration
      failCgroupV1: false
  ```
- Containers cannot reach the host proxy, so images are pulled on the host and
  imported into the node with `docker save | docker exec -i <node> ctr -n k8s.io images import -`.
  `kind load docker-image` fails on single-platform pulls. `targets/kind/up`
  does all of this.

**OpenSpec.** Install with
`npm install --no-save @fission-ai/openspec@1.4.0` (delete any stray
`package-lock.json`), then use `OPENSPEC_TELEMETRY=0 ./node_modules/.bin/openspec ...`.

**Research clones** (not committed) live in `/home/user/research/`:
`openmeter`, `otelcol-contrib`, `hyperdx`, `clickhouse-docs`.

**Disk.** The sandbox disk is small. ClickHouse at `information` log level
filled it twice through Docker's json logs, which left files truncated to zero
bytes. `targets/compose/gen.py` now logs at `warning` with rotation. Prune
unused images and volumes before large runs, and check `df -h /`.

**Evidence runs.** Run evidence from a clean worktree so `--require-clean`
and the harness commit stay valid while the main tree is edited:
```bash
git worktree add --detach /home/user/ch-eval-run HEAD
ln -s /home/user/ch-eval/targets/.state /home/user/ch-eval-run/targets/.state
```
Then copy `results/` back. Do not use `pkill -f` or `pgrep -f` with patterns
that also appear in the calling shell's command line, because they match the
shell itself.

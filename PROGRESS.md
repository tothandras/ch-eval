# PROGRESS

Durable log for this evaluation. **Every session starts by reading this file
and `openspec/changes/cross-target-evaluation/` (proposal, design, tasks).**
Keep it accurate enough that a fresh agent can continue.

## Current state

- **Phase:** 1 complete. **Waiting for owner review.** Phase 2 does not start
  until the questions below are answered.
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

## Questions for the owner (Phase 1 stop point)

1. **Budget:** what is the cap for paid resources? The rough plan estimate is
   about $300–500 total and will be re-estimated before each paid run. Which
   AWS account or profile, and which ClickHouse Cloud organization, should be
   used?
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

## Next (Phase 2, after approval)

See `openspec/changes/cross-target-evaluation/tasks.md` §2:
- compose and kind targets;
- target metadata and results layout;
- migration tool and tests;
- probes and the generated feature matrix;
- OTel generators and the collector path;
- cross-target compare;
- CI;
- the replica-consistency smoke run.

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
- Containers cannot reach the host proxy, so load images into kind with
  `kind load docker-image`.

**OpenSpec.** Install with
`npm install --no-save @fission-ai/openspec@1.4.0` (delete any stray
`package-lock.json`), then use `OPENSPEC_TELEMETRY=0 ./node_modules/.bin/openspec ...`.

**Research clones** (not committed) live in `/home/user/research/`:
`openmeter`, `otelcol-contrib`, `hyperdx`, `clickhouse-docs`.

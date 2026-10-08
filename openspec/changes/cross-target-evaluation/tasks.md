## 1. Phase 1: plan and draft feature matrix (STOP: owner review)

- [x] 1.1 Read the source harness (README, bench/README, scenarios, cells, pricing, openspec) and fork it with only the module path renamed; build and tests green on Go 1.26.2
- [x] 1.2 Research OpenMeter's ClickHouse DDL management (`docs/research/openmeter-clickhouse-ddl.md`)
- [x] 1.3 Research exporter-default and ClickStack OTel DDL (traces, metrics, logs) and copy it verbatim (`docs/research/otel-ddl/`)
- [x] 1.4 Pin versions: Altinity Stable 26.3.33.10001, upstream 26.3.33.24, Cloud 26.4/26.6, operator 0.27.4, collector 0.162.0
- [x] 1.5 Local feasibility: Docker images run; kind works with the sandbox patches; the unchanged harness runs on both 26.3 builds with digests 30/30 and 34/34
- [x] 1.6 Draft `docs/feature-matrix.md` with cited hypotheses and planned probes
- [x] 1.7 Write this proposal, design, delta specs, ADRs 0001–0005, PROGRESS.md
- [x] 1.8 Owner additions folded in: replication and sharding (D13), compute/data separation (D14), on-demand scaling and rightsizing (D15); feature-matrix sections 9–11 drafted with citations
- [ ] 1.9 Owner review: budget, region, shape, `upstream-cloudver` control, open questions in PROGRESS.md

## 2. Phase 2: local targets, migrations, generators, parity green

- [x] 2.1 `targets/compose`: `local-single` and `local-cluster` (2 replicas + 3 Keeper + S3 emulator) for `altinity` and `upstream-oss`; one `up`/`down` each
- [x] 2.2 `targets/kind`: Altinity operator 0.27.4 + CHK (3) + CHI (1×2) for both images; sandbox patches behind cgroup-v1 detection; images loaded into the node (operator bundle pinned by sha256; images imported with `ctr` because `kind load` fails on single-platform pulls; bundle pull policy set to IfNotPresent locally; CHI cluster names cannot contain `_`)
- [x] 2.3 Target manifests plus `bench/target`: DSN resolution, live metadata capture, drift check (fails the run unless `--allow-drift`)
- [x] 2.4 Results layout `results/<workload>/<scenario>/<target>/<run-id>/`; `--target`; workload-nested discovery (OpenMeter queries byte-identical)
- [x] 2.5 `bench migrate` (render/up/down/status/verify/force), `schema_migrations_ch`, lint, cost capture (`--measure`); port the OpenMeter inits to templated migrations
- [x] 2.6 Migration tests (up→down→up→up) on single, replicated (compose cluster and kind) for both builds
- [x] 2.7 Probe runner plus `docs/feature-matrix.yaml` plus the generator for `feature-matrix.md`; implement the locally runnable probes — 50 probes (45 SQL + 5 Go); all confirmed on altinity (single, 1×2, 2×2, kind, s3 variants) and upstream-oss (single, 1×2, kind); local experiment observations shown per row
- [x] 2.8 OTel generators (traces, metrics, logs), deterministic with unit tests; direct-insert mappers for all variants (exporter-default and clickstack, ported by `migrations/tools/port_otel_ddl.py`)
- [x] 2.9 Collector ingest path (compose service, `create_schema: false`) plus the path-parity test — `targets/collector`, `--ingest-path collector`, `bench compare --runs`; full-table fingerprints identical for all six OTel scenarios
- [x] 2.10 Cross-target `bench compare` with parity gate; local parity green for every workload and scenario on altinity vs upstream-oss (`results/parity/`: single, 1×2, 2×2, kind, s3 variants)
- [x] 2.11 CI workflow: lint, unit tests, `openspec validate`, migration tests, smoke run on altinity and upstream-oss (`.github/workflows/ci.yml`; gofmt cleanup landed as its own no-behaviour commit; green on GitHub Actions, smoke job included)
- [x] 2.12 `replica-consistency` experiment implemented and smoke-tested on local-cluster and kind (no conclusions) — `bench consistency`; altinity and upstream 1×2 (compose), altinity and upstream kind; restart variant on compose
- [ ] 2.13 Sharded topologies: compose `local-cluster` 2x2 profile and kind CHI 2x2; `shardedTable` migration helper; digest parity 2x2 vs 1x2 for openmeter — compose 2×2 done with parity for every workload (1 and 8 namespaces for openmeter); kind 2×2 not run (sandbox memory), the compose 2×2 covers the sharded DDL and parity questions
- [x] 2.14 Opt-in Zipf namespace distribution in the generator (appended draw; default stream byte-identical, proven by test) — `--namespace-skew`
- [x] 2.15 Local correctness experiments: S3 (unavailable shard: default vs `skip_unavailable_shards`), S4 (Distributed insert modes, time-to-queryable), R3 (Keeper quorum loss on kind), R6 (ON CLUSTER with a replica down), E5/E6 (operator vertical change, add replica) at small scale — S3/S4 on altinity 2×2; R3/R6 on both builds (compose, which also runs 3 Keepers); E5/E6 on altinity kind
- [x] 2.16 Self-managed `s3` storage mode locally (S3 emulator + filesystem cache); cache-size cap knob for C2; locality metadata capture — `local-cluster-s3`, `-s3-cow` (cache on write) and `-s3-smallcache` (256 MiB) tiers, parity green; disks, object-storage type, cache occupancy and the effective cache settings in `metadata.json`, plus `seed.cache_after_seed`; write-through needs `enable_filesystem_cache_on_write_operations` as well as the disk knob (found by the snapshot, fixed in the cow tier)

## 3. Phase 3: OpenMeter cross-target perf (paid; STOP before creating resources)

- [ ] 3.1 `targets/aws` Terraform (VPC, EKS, node groups, gp3, S3 bucket, client VM; run-id tags) with tested `down --verify`
- [ ] 3.2 `targets/cloud` (API or Terraform provider 3.37.0; run-id in name/tags) with tested `down --verify`
- [ ] 3.3 Spend estimate → owner approval
- [ ] 3.4 10M perf runs (`--repeat 3`, warm and cold) for both scenarios × cloud, altinity(ebs), altinity(s3), upstream-oss(ebs) [+ upstream-cloudver]; parity gate
- [ ] 3.5 Online-change migrations measured per target
- [ ] 3.6 replica-consistency on cloud and altinity (and upstream-oss)
- [ ] 3.7 Topology sweep (paid, separate approval): altinity `1x2` / `1x4` (parallel replicas) / `2x2` vs cloud at matched total vCPU; S2 shard pruning; R4 replica rebuild at 100M; R5 storage per replica
- [ ] 3.8 Locality (D14): cold/warm and window sweep on cloud vs `ebs` vs `s3`; cache-size cap sweep on `s3`; write path and S3 requests per 1M events; optional warehouse isolation (C7)

## 4. Phase 4: OTel traces

- [ ] 4.1 Port exporter-default and clickstack trace DDL to init migrations (verbatim plus placeholders)
- [ ] 4.2 Query set (trace by id, latency percentiles by service/span over time, error rate, attribute filter, top-N slow, dependency edges, 1h/24h/7d scans)
- [ ] 4.3 A/B the tuned candidates locally for correctness, then on paid tiers; justify the tuned variant in an ADR
- [ ] 4.4 Collector ingest run per target; perf runs; parity

## 5. Phase 5: OTel metrics and logs

- [ ] 5.1 Metrics variants, queries (rate/increase, gauge windows, histogram quantiles, label group-by, high-cardinality filters, MV rollups vs query time), high-cardinality profile
- [ ] 5.2 Logs variants and queries (full-text, severity/service, attributes, trace correlation, top-N, time scans)
- [ ] 5.3 Collector ingest run per target; perf runs; parity

## 6. Phase 6: COGS and other aspects

- [ ] 6.1 Self-managed pricing profiles; `kind` in the pricing package; S3 request accounting
- [ ] 6.2 COGS cells at 100M preload (idle, ingest, query, mixed) on cloud and altinity at matched shapes; reconcile against the Cloud statement and AWS Cost Explorer
- [ ] 6.3 Cost curves and crossovers
- [ ] 6.4 Concurrency: max QPS at the p95 SLO per class with ingest running
- [ ] 6.5 Ingest: sustained events/s, merge CPU share, parts over time, time-to-queryable
- [ ] 6.6 Reliability: replica loss, Keeper loss, rolling upgrade (self-managed); Cloud equivalents where possible
- [ ] 6.7 Backup/restore at 100M plus cross-target portability (cloud → self-managed and back)
- [ ] 6.8 Security and multi-tenancy probes; client-compatibility probes
- [ ] 6.9 Operational-burden write-up with evidence
- [ ] 6.10 Scaling (D15): Cloud scheduled-scaling transition (E1a), optional ≥ 4 h sustained-load autoscaling (E1b, separate approval), scale-out (E2), idle/resume (E3), scheduled scaling probe (E4); self-managed vertical roll (E5), add replica (E6), add shard (E7); EBS online expansion (E10); merge and mutation headroom (E11)
- [ ] 6.11 Rightsizing sweep (E13) per workload: Cloud 8/16/32 GiB vs self-managed 2/4/8 vCPU pods at the declared load and SLO; storage cost per retention (E14)
- [ ] 6.12 Load-profile cost model and crossovers (E15), feeding the verdict flip conditions

## 7. Phase 7: report

- [ ] 7.1 `bench report` with verdict rules plus flip conditions; `docs/methodology.md` final
- [ ] 7.2 Regenerate `REPORT.md` and the feature matrix; verify no cloud resources remain; archive this change

## 8. AI Cost Management (internal workload, owner addition 2026-10-08)

- [x] 8.1 Store the brief (`internal/ai-cost-management/BRIEF.md`) and its SQL verbatim; mark `internal/` as a publication blocker
- [x] 8.2 Q1 feature acceptance on Altinity 26.3.16/26.3.33, upstream 26.3.33/26.4.5/26.7.4/26.8.2
- [x] 8.3 Local 3 × 2 cluster copying D-05/D-07/D-09/D-10; deterministic synthetic data with redeliveries, changed-time retries and late events
- [x] 8.4 Experiments:
  - parity (single node vs D-09 vs tenant sharding);
  - projections;
  - `FINAL` plan;
  - isolation with and without the secret;
  - table-function bypass on 26.3.16;
  - D-13 under lag;
  - mutations and erasure;
  - D-12 with a replica down;
  - materialized-view cost;
  - map buckets;
  - export memory;
  - `parallel_distributed_insert_select`
- [x] 8.5 `internal/ai-cost-management/REPORT.md` in the brief's six-part format, plus appendix answers to section 10
- [ ] 8.6 Owner review; decide whether this repo should run any of the measurement requests on paid tiers


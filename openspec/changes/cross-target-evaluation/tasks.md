## 1. Phase 1: plan and draft feature matrix (STOP: owner review)

- [x] 1.1 Read the source harness (README, bench/README, scenarios, cells, pricing, openspec) and fork it with only the module path renamed; build and tests green on Go 1.26.2
- [x] 1.2 Research OpenMeter's ClickHouse DDL management (`docs/research/openmeter-clickhouse-ddl.md`)
- [x] 1.3 Research exporter-default and ClickStack OTel DDL (traces, metrics, logs) and copy it verbatim (`docs/research/otel-ddl/`)
- [x] 1.4 Pin versions: Altinity Stable 26.3.33.10001, upstream 26.3.33.24, Cloud 26.4/26.6, operator 0.27.4, collector 0.162.0
- [x] 1.5 Local feasibility: Docker images run; kind works with the sandbox patches; the unchanged harness runs on both 26.3 builds with digests 30/30 and 34/34
- [x] 1.6 Draft `docs/feature-matrix.md` with cited hypotheses and planned probes
- [x] 1.7 Write this proposal, design, delta specs, ADRs 0001–0005, PROGRESS.md
- [ ] 1.8 Owner review: budget, region, shape, `upstream-cloudver` control, open questions in PROGRESS.md

## 2. Phase 2: local targets, migrations, generators, parity green

- [ ] 2.1 `targets/compose`: `local-single` and `local-cluster` (2 replicas + 3 Keeper + S3 emulator) for `altinity` and `upstream-oss`; one `up`/`down` each
- [ ] 2.2 `targets/kind`: Altinity operator 0.27.4 + CHK (3) + CHI (1×2) for both images; sandbox patches behind cgroup-v1 detection; images loaded with `kind load`
- [ ] 2.3 Target manifests plus `bench/target`: DSN resolution, live metadata capture, drift check
- [ ] 2.4 Results layout `results/<workload>/<scenario>/<target>/<run-id>/`; `--target`; workload-nested discovery (OpenMeter queries byte-identical)
- [ ] 2.5 `bench migrate` (render/up/down/status/verify), `schema_migrations_ch`, lint, cost capture; port the OpenMeter inits to templated migrations
- [ ] 2.6 Migration tests (up→down→up→up) on single, replicated (compose cluster and kind) for both builds
- [ ] 2.7 Probe runner plus `docs/feature-matrix.yaml` plus the generator for `feature-matrix.md`; implement the locally runnable probes
- [ ] 2.8 OTel generators (traces, metrics, logs), deterministic with unit tests; direct-insert mappers for all variants
- [ ] 2.9 Collector ingest path (compose service, `create_schema: false`) plus the path-parity test
- [ ] 2.10 Cross-target `bench compare` with parity gate; local parity green for every workload and scenario on altinity vs upstream-oss
- [ ] 2.11 CI workflow: lint, unit tests, `openspec validate`, migration tests, smoke run on altinity and upstream-oss
- [ ] 2.12 `replica-consistency` experiment implemented and smoke-tested on local-cluster and kind (no conclusions)

## 3. Phase 3: OpenMeter cross-target perf (paid; STOP before creating resources)

- [ ] 3.1 `targets/aws` Terraform (VPC, EKS, node groups, gp3, S3 bucket, client VM; run-id tags) with tested `down --verify`
- [ ] 3.2 `targets/cloud` (API or Terraform provider 3.37.0; run-id in name/tags) with tested `down --verify`
- [ ] 3.3 Spend estimate → owner approval
- [ ] 3.4 10M perf runs (`--repeat 3`, warm and cold) for both scenarios × cloud, altinity(ebs), altinity(s3), upstream-oss(ebs) [+ upstream-cloudver]; parity gate
- [ ] 3.5 Online-change migrations measured per target
- [ ] 3.6 replica-consistency on cloud and altinity (and upstream-oss)

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

## 7. Phase 7: report

- [ ] 7.1 `bench report` with verdict rules plus flip conditions; `docs/methodology.md` final
- [ ] 7.2 Regenerate `REPORT.md` and the feature matrix; verify no cloud resources remain; archive this change

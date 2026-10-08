# ch-eval: ClickHouse Cloud vs self-managed Altinity, measured

**Question.** For OpenMeter usage events, OTel traces, OTel metrics and OTel
logs, how do ClickHouse Cloud and self-managed Altinity ClickHouse (OSS)
compare on capability, performance, cost and operational burden?

**Status.** Phase 2 (local targets, migrations, harness integration) is in
progress. No performance conclusions yet; local runs are smoke and parity
only. See [`PROGRESS.md`](PROGRESS.md) for the live log and
[`openspec/changes/cross-target-evaluation/`](openspec/changes/cross-target-evaluation/)
for the plan.

## Targets

| Target | What | Why |
| --- | --- | --- |
| `cloud` | ClickHouse Cloud (SharedMergeTree, object storage) | Managed baseline |
| `altinity` | Altinity Stable `26.3.33.10001.altinitystable`, Altinity operator + ClickHouse Keeper | The candidate |
| `upstream-oss` | Upstream `26.3.33.24` (same 26.3.33 base), same topology | Separates build effects from version and architecture effects |

## Layout

| Path | Contents |
| --- | --- |
| `bench/` | Go harness forked from [`tothandras/clickhouse-benchmarks`](https://github.com/tothandras/clickhouse-benchmarks) (perf runner, value digest, COGS), extended with targets, migrations, probes, compare and report ([ADR-0001](docs/adr/0001-fork-existing-harness.md)) |
| `targets/` | Per-target `up`/`down`: compose, kind with the Altinity operator, AWS (Terraform), ClickHouse Cloud (API) |
| `migrations/` | The dialect-aware migration tool and per-workload online changes ([ADR-0003](docs/adr/0003-migration-tool.md)) |
| `scenarios/<workload>/<variant>/` | `init/` migrations, `queries/`, `mix.json` |
| `generators/` | Seeded OTel traces, metrics and logs generators. OpenMeter events reuse `bench/seed`. |
| `results/<workload>/<scenario>/<target>/<run-id>/` | Raw results plus run metadata |
| `docs/` | [`feature-matrix.md`](docs/feature-matrix.md), [`methodology.md`](docs/methodology.md), [`adr/`](docs/adr/), [`research/`](docs/research/) |
| `REPORT.md` | Generated from `results/` by `bench report`, never hand-typed |

## Quick start (local targets)

```bash
targets/altinity/up local-single
targets/upstream-oss/up local-single
go build -o bin/bench ./bench/cmd/bench

# fresh migrate (down+up) + seed + timed queries + digests, per target
for t in altinity/local-single upstream-oss/local-single; do
  ./bin/bench --target $t --workload openmeter --rows 1000000 --time-end 2026-06-01T00:00:00Z
done

# cross-target compare with the value-parity gate (exits non-zero on mismatch)
./bin/bench compare --workload openmeter --scenario proposal \
  --targets altinity/local-single,upstream-oss/local-single

./bin/bench migrate status --target altinity/local-single --scenario proposal
./bin/bench migrate verify --target altinity/local-single --scenario proposal

targets/altinity/down local-single && targets/upstream-oss/down local-single
```

DSNs come from each manifest's `dsn_env` variable (see `.envrc.example`).
Local tiers fall back to their passwordless default.

`clickhouse-benchmark` must be on `PATH`. Any ClickHouse server image ships it
as `/usr/bin/clickhouse`.

## Rules this repo follows

- Synthetic data only.
- Credentials live only in the environment or the git-ignored `.envrc`.
- Paid resources are created only after owner approval. Each one is tagged
  with the run ID and torn down with a verified `down`.

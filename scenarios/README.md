# scenarios

Scenarios are grouped by workload: `scenarios/<workload>/<variant>/`. Each
variant is one table design for that workload.

Current variants of the `openmeter` workload:

- `openmeter/baseline-openmeter/` — Verbatim reproduction of OpenMeter's `om_events` (`data String`, queried via `JSON_VALUE`). The reference.
- `openmeter/proposal/` — **The recommended default** of the forked harness. Combines `data JSON` + `CODEC(ZSTD(3))` + a `bloom_filter` skip index on `id`, queried via native subcolumns. Includes the full meter-query set plus `lookup_by_id.sql` (so the `id` bloom is exercised) and the extra grouped Kong api variants.

OTel workloads (`otel-traces`, `otel-logs`, `otel-metrics`), each with two
variants:

- `exporter-default/` — the collector-contrib ClickHouse exporter's own DDL
  (v0.162.0, `json: false`, text indexes).
- `clickstack/` — ClickStack's schema (hyperdx seeds), including its
  key/value rollup views.

They are ported verbatim apart from placeholders by
`migrations/tools/port_otel_ddl.py` ([ADR-0004](../docs/adr/0004-otel-scenarios.md)).
Each variant has its own database, `<workload>_<variant>` (for example
`otel_traces_clickstack`). A `tuned` variant is added in Phase 4/5 from A/B
evidence.

Generators (`generators/otel`) are deterministic in `--seed` and
`--time-end`. `--rows` counts spans, log records, or metric points across the
five metric tables. `--ingest-path collector` loads the same data through the
pinned OpenTelemetry Collector instead of the direct mapper.

```
scenarios/<workload>/<variant>/
  scenario.json     # workload, database, logical tables (first = seeded table), generator
  init/             # migrations: YYYYMMDDHHMMSS_name.up.sql[.tmpl] + .down.sql[.tmpl]
  queries/          # one .sql file per benchmark query (one SELECT each), plus mix.json
  parity/           # optional untimed checks digested with the queries (full-table fingerprints)
```

`scenario.json`:

```json
{"workload": "openmeter", "database": "openmeter", "tables": ["proposal_events"], "generator": "openmeter-events"}
```

**Table naming.** Tables keep their names from the forked harness
(`baseline_openmeter_events`, `proposal_events`), so the query files stay
byte-identical. Each workload has its own database (`openmeter`), and the
harness connects to it, so queries use unqualified table names.

## init/ contract (migrations)

- **Format and tooling.** `init/` is a migration set applied by
  `migrations/migrate` ([ADR-0003](../docs/adr/0003-migration-tool.md)).
  Files are named golang-migrate style, and `.tmpl` files are Go templates
  rendered per target dialect:
  - `single`: plain MergeTree;
  - `replicated`: `Replicated*` engines with `ON CLUSTER`, plus a
    `Distributed` table and `<table>_local` on sharded topologies;
  - `cloud`: SharedMergeTree, no `ON CLUSTER`.
- **Template helpers.** `db`, `table`, `onCluster`, `engine`,
  `globalEngine`, `distributed`, `dropDistributed`, `settings`, `cluster`,
  `dialect` and `sharded`. `bench migrate render --target T --scenario S`
  prints what a target would run.
- **Idempotent.** Every statement uses `IF [NOT] EXISTS` (lint-enforced).
- **Down migrations drop what the up migration created.** A fresh `bench`
  run rolls the set down and up again before seeding.
- **Append-only once applied.** An applied migration whose rendered
  checksum changes is an error. Ship a new migration instead.

## Seed contract

Scenarios that need data populated can either:

1. **Use a registered generator** named in `scenario.json`. `openmeter-events`
   is the shared Go seeder under `bench/seed/`. This is the default
   for scenarios that need a structured JSON payload, deterministic RNG, or
   controllable group-by cardinality (most baseline OpenMeter-shaped
   scenarios). The seeder writes events matching the OpenMeter event shape;
   the harness invokes it in-process.

   The seeder emits a **weighted mix of heterogeneous event types** to model
   real OpenMeter usage, where the `data` field is user-controlled and differs
   per event type. The default catalog is: `api_request` (50%, the baseline
   type carrying `{value, group1, group2}` that the canonical meter queries
   read), `kong_api_request` (25%, HTTP request/response fields), `llm_request`
   (15%, `tokens`/`model`/`provider`), `workload` (7%, `duration_seconds`/
   `region`), and `agent_run` (3%, `agent_name`). Numeric payload fields are
   emitted as JSON strings (e.g. `"tokens":"1"`) to match real producers, so
   queries must extract them with `toDecimal128OrNull(...)` over the path's
   stringified form (any JSON-stored type, for exact billing). Selection
   is seedable, so reruns produce byte-identical data across table variants.
   Per-type queries (`llm_tokens_by_model`, `kong_status_by_route`,
   `workload_seconds_by_region`, `agent_runs_by_name`) aggregate each type's
   own fields filtered by its `type`.

2. **Skip seeding** (`bench --skip-seed`). The harness then applies only
   pending migrations and runs the queries against whatever data is already
   in the table, which is useful for a pre-populated reference dataset.

## queries/ contract

- One `.sql` file per query. The filename without extension is the query
  identifier reported in results.
- Exactly one SELECT statement per file. Multiple statements produce an error
  and that query is skipped.
- Parameters use ClickHouse's native `{name:Type}` placeholder syntax. The
  harness binds them from a fixed default set documented below.

### Default parameter set (v1)

For v1 the harness binds a small, fixed set of parameters from compiled-in
defaults so scenarios that conform to the baseline OpenMeter event shape work
out of the box without per-scenario configuration:

| Parameter            | Type     | Default                            |
| -------------------- | -------- | ---------------------------------- |
| `{namespace:String}` | String   | `"default"`                        |
| `{type:String}`      | String   | `"api_request"`                    |
| `{from:DateTime}`    | DateTime | seeder time-span start             |
| `{to:DateTime}`      | DateTime | seeder time-span end               |
| `{subjects:Array(String)}` | Array(String) | first 10 subjects from the seeder pool |

A future scenario that needs different parameters will introduce a
per-scenario `params.json` manifest. We deliberately delay that until a second
variant forces the need, to keep v1 small.

### `queries/mix.json` (cogs replay mixes)

Scenarios that participate in `bench cogs` runs ship a `queries/mix.json`
declaring named query mixes. The contract is strict by design — adding a query
file forces a conscious classification decision:

- Every `*.sql` in `queries/` must appear in **exactly one** class or in the
  mix's `exclude` list; loading fails otherwise, naming the offender.
- Every name referenced by a class must have a matching `.sql` file.
- `exclude` is for diagnostic variants that are not production traffic (e.g.
  the `*_no_prewhere` PREWHERE-comparison queries).
- Class `weight`s set the per-arrival selection probability; queries within a
  class are selected uniformly.
- While weights are placeholders (not measured production frequencies), the
  mix carries a `notes` field that every cogs report renders as a caveat.

```jsonc
{
  "production": {
    "notes": "weights are placeholders pending measured frequencies",
    "classes": {
      "key_only":  { "weight": 10, "queries": ["count_total", "..."] },
      "meter_agg": { "weight": 70, "queries": ["sum_hour", "..."] }
    },
    "exclude": ["sum_hour_group1_no_prewhere"]
  }
}
```

`mix.json` lives inside `queries/` but is invisible to the perf path, whose
discovery only picks up `*.sql` files.

## Discovery rules

- The harness walks `scenarios/<workload>/` (`--workload`, default
  `openmeter`). Every subdirectory containing `init/` and `scenario.json` is a
  scenario.
- Files starting with `_` or `.`, and the `README.md` itself, are ignored.
- A scenario can be selected explicitly via `--scenario <name>` (repeatable)
  on the harness CLI.

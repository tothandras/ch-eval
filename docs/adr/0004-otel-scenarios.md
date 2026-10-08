# ADR-0004: OTel workloads use verbatim upstream schemas plus one evidence-justified tuned variant

- Status: Proposed (Phase 1). The tuned variants are decided in Phases 4–5.
- Date: 2026-10-08

## Context

- **Recommended practice:** both the exporter and ClickHouse docs recommend
  `create_schema: false` and owning the DDL.
- **Two credible upstream schemas exist:**
  - the exporter templates (contrib v0.162.0);
  - ClickStack's seeds (hyperdx `eeabaebe`).
- **Scope change:** OTel logs were added to scope by the repo owner on
  2026-10-08.

## Decision

**Workloads.** `otel-traces`, `otel-metrics` and `otel-logs`, each with
three variants:

| Variant | Source | Allowed deviation |
| --- | --- | --- |
| `exporter-default` | contrib v0.162.0 templates, `json: false` | only database, engine and TTL placeholders differ from the source |
| `clickstack` | hyperdx main | same |
| `tuned` | starts from the better of the two | each lever earns its place through an A/B result recorded in `results/` and an ADR amendment |

**Tuned-variant candidates:**
- materialized hot attributes;
- a TraceId lookup table or projection;
- JSON attributes, which ClickStack itself says are beta, so this needs
  evidence;
- a series-id dimension table for metrics;
- rollup MVs for metrics.

**Shared defaults:**
- All variants use the ≥26.2 `text`-index forms, because every target is at
  least 26.2.
- Generators are deterministic. A direct-insert path is used for scale; an
  OTLP→collector path is used at least once per target, with path parity
  checked by digests.
- Variant databases are named `<workload>_<variant>` so table names stay as
  upstream publishes them.

## Consequences

- Comparing against unmodified upstream schemas keeps the evaluation honest
  about out-of-the-box behaviour on each platform.
- The tuned variant cannot be cherry-picked: its levers are documented with
  their measured effect.

## Implementation notes (Phase 2)

- **Porting.** `migrations/tools/port_otel_ddl.py` builds each variant's
  `init/` migration from the verbatim blocks in `docs/research/otel-ddl/`.
  Every substitution must match exactly once; the script is the record of
  what differs from upstream. The substitutions are:
  - database and table names, and the cluster clause;
  - the engine;
  - the TTL;
  - the `SETTINGS` clause, which appends the storage policy on `s3` targets.

  Distributed tables are added only on sharded topologies, with these sharding
  keys:

  | Tables | Sharding key |
  | --- | --- |
  | traces | `cityHash64(TraceId)`, which keeps a trace on one shard |
  | logs, metrics, rollups | `rand()` |

- **TTL.**
  - `exporter-default` keeps the exporter default, which is no TTL (`ttl` 0).
  - `clickstack` renders its 720h TTL macros as `toIntervalDay(3650)`. Pinned
    benchmark windows lie in the past, and a 30-day TTL would drop them.
- **ClickStack seeds included:** `00002` logs, `00003` metrics, `00005`
  traces, and the `00006`/`00007` key/value rollups with their materialized
  views.
- **ClickStack seeds not included:**
  - `00004_hyperdx_sessions`: browser sessions, not an OTel signal here.
  - `00008_otel_metrics_timeseries`: the experimental `TimeSeries` engine,
    which the OTLP exporter does not write. It is covered by a feature probe
    instead.
  - The `<26.2` compat files.
- **Collector path.** The collector is contrib 0.162.0 with
  `create_schema: false`, `async_insert`, `sending_queue.batch` and
  `block_on_overflow` (`targets/collector/`). It is fed OTLP/HTTP with the
  JSON encoding rather than OTLP/gRPC. The receiver, decoder and exporter are
  the same, and no protobuf or gRPC dependency is needed.
- **Path parity.** `parity/*.sql` queries hash every column the exporter
  inserts, per row, summed in an order-independent way. Materialized-view
  targets are merged per key first, so the result does not depend on insert
  block boundaries. These queries are digested but not timed.
  On 2026-10-08, locally on `altinity/local-single`, the collector path and
  the direct mapper produced identical fingerprints for every table of all six
  scenarios (50k rows each). The direct mapper therefore reproduces the
  exporter's v0.162.0 column mapping for the generated data.
- **Generators** live in `generators/otel`. They use string and int attribute
  values only: `pcommon.Value.AsString` renders doubles in a way that would
  need separate verification.

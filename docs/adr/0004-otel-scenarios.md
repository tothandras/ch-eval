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

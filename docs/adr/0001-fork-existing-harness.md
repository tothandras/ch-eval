# ADR-0001: Fork the existing harness; extend, don't rewrite

- Status: Accepted
- Date: 2026-10-08

## Context

[`tothandras/clickhouse-benchmarks@542b1e5`](https://github.com/tothandras/clickhouse-benchmarks/tree/542b1e5)
already provides the pieces this evaluation needs:

- the perf runner (`clickhouse-benchmark` timing plus `query_log` CPU);
- the value digest that proves parity across table designs;
- the COGS runner, which attributes CPU to insert, merge, query and idle from
  system tables and prices it;
- seeded OpenMeter event generation;
- the OpenSpec workflow.

Its COGS model reconciled within 1–3% of a ClickHouse Cloud bill. The
evaluation adds one dimension (the target platform) and new workloads.

## Decision

- **Fork verbatim** into this repository. The only change is the Go module
  path (`github.com/openmeterio/ch-playground` → `github.com/tothandras/ch-eval`).
- **Keep these packages unchanged:**
  - `bench/runner` (including `digest.go`);
  - `bench/cogs/attribution.go`;
  - `bench/seed`;
  - the `bench/accounting` collectors' SQL;
  - `bench/pricing` billed-shape semantics.
- **Extend additively** through new packages: `bench/target`, `bench/migrate`,
  `bench/probe`, `bench/report`, `bench/consistency`, `generators/*`.
- **Allow only small, documented edits** at these seams: discovery depth,
  result path, the cluster name in `accounting.Target`, and the pricing
  profile `kind`.
- **Keep the forked OpenSpec specs** as the baseline. Changes go through delta
  specs.
- **Do not copy the source's committed result files.** They remain available
  in the source repository.

## Consequences

- Earlier single-target results and new cross-target results come from the same
  measurement code, so they stay comparable.
- Any future change to digest normalization, CPU source selection or the
  attribution split needs its own ADR, because it would invalidate
  comparisons with prior runs.
- The forked scenario files are byte-identical to the source, including
  comments. One comment in `scenarios/proposal/init.sql` mentions the row count
  of an internal data copy. It is already public in the source repository, and
  scrubbing it is an open question for the owner.

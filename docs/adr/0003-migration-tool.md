# ADR-0003: ClickHouse migrations with golang-migrate layout and a custom dialect-aware runner

- Status: Proposed (Phase 1). Final shape after the Phase 2 probes ST-03 (ON CLUSTER on Cloud) and the settings probes.
- Date: 2026-10-08

## Context

The evaluation brief says to "check how OpenMeter manages ClickHouse DDL today
and align with it". See [research](../research/openmeter-clickhouse-ddl.md).

**OpenMeter's ClickHouse side**
- It runs one `CREATE TABLE IF NOT EXISTS om_events … ENGINE = MergeTree` at
  every process start.
- There is no migration tool, no version table and no `ALTER` path.
- Past schema changes (`store_row_id`, the `stored_at` index) never reached
  existing tables automatically.
- There is no cluster support. A hard-coded `MergeTree` is converted to
  SharedMergeTree on Cloud, but on a self-managed replicated cluster it creates
  independent, unreplicated tables on each replica.

**OpenMeter's Postgres side**
- Atlas generates golang-migrate files: `YYYYMMDDHHMMSS_name.up.sql` /
  `.down.sql`, embedded with `go:embed`, version table `schema_om`, and a
  `migration-job` mode.

**What this evaluation needs that OpenMeter does not have**
1. One source of truth rendered for three dialects:

   | Dialect | DDL shape |
   | --- | --- |
   | `single` | plain MergeTree |
   | `replicated` | `ON CLUSTER` + `ReplicatedMergeTree` with Keeper paths |
   | `cloud` | SharedMergeTree, Shared database |

2. A replicated version table, so migration state is not per-replica.
3. Per-migration duration and resource capture, to measure online changes such
   as String→JSON and `MATERIALIZE INDEX` per target.

## Decision

**Align with OpenMeter's conventions rather than its absence of a tool:**
- golang-migrate file naming and up/down pairing;
- embedded files;
- a dedicated version table, `schema_migrations_ch`, analogous to `schema_om`.

**Add a thin runner** (`bench migrate`) with:
- **Template helpers.** Go `text/template` with a closed helper set:
  `onCluster`, `engine`, `db`, `dialect`. Arbitrary per-target forks are
  rejected by lint.
- **`render` subcommand.** Emits plain SQL per dialect. That output is
  golang-migrate compatible, so OpenMeter could adopt the files with its
  existing tool.
- **Replicated state.** The version table uses the dialect's engine and
  `ON CLUSTER`. State reads are synchronized: `SYSTEM SYNC REPLICA` on
  `replicated`, `select_sequential_consistency` on `cloud`.
- **Checksums.** Each applied version stores the checksum of its rendered SQL.
  A mismatch is an error. Dirty-flag semantics follow golang-migrate.
- **Synchronous ALTERs.** `alter_sync=2` and `mutations_sync=2`, plus waiting
  on `system.mutations` until done.
- **Cost capture.** Statements are tagged with `log_comment` so `query_log`
  and `part_log` yield duration, CPU and bytes.
- **Idempotency lint.** Up statements must use `IF [NOT] EXISTS`.

**Rejected alternatives**
- **golang-migrate's ClickHouse driver alone.** It has no dialect templating
  and no resource capture. Its default version-table engine is not replicated
  across replicas; that is to be confirmed by a probe in Phase 2.
- **Atlas for ClickHouse.** It adds a heavy dependency and its ClickHouse
  support is not used by OpenMeter today. Rendered SQL would still need
  templating.
- **Keeping `CREATE … IF NOT EXISTS` at startup**, as OpenMeter does. It cannot
  evolve existing tables, and racing creators are exactly what the OTel
  exporter docs warn against with `create_schema`.

## Consequences

- The evaluation's own DDL becomes a reusable starting point for OpenMeter.
- **Not a production requirement:** making ClickHouse schema changes reach
  existing deployments is out of scope here. It is noted as a design callout
  candidate.
- **Open until Phase 2 probes:** the `cloud` dialect's `ON CLUSTER` rendering,
  and the settings Cloud rejects.

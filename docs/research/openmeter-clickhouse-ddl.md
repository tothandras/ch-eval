# How OpenMeter manages ClickHouse DDL today

Source: `openmeterio/openmeter` at
[`5a0e27e6`](https://github.com/openmeterio/openmeter/tree/5a0e27e6f50b301e9bdb93772a6a970fd4a1f58f)
(2026-10-07). Read-only source inspection, input to
[ADR-0003](../adr/0003-migration-tool.md).

## Findings

1. **One table, created at process start.** `om_events` is the only ClickHouse
   object. The streaming connector runs `CREATE TABLE IF NOT EXISTS` from
   `openmeter/streaming/clickhouse/connector.go:78-82` and `:291-303` unless
   `SkipCreateTables` is set. Only a test sets it. The DDL is built with
   `huandu/go-sqlbuilder` in `openmeter/streaming/clickhouse/event_query.go:20-50`.
   Every binary (server, jobs, sink-worker, balance-worker, billing-worker,
   notification-service) runs it on startup.
2. **No ClickHouse migrations.** There is no version table, no `ALTER`, and no
   migration tool for ClickHouse. Schema changes are edits to the `CREATE`
   statement, which `IF NOT EXISTS` turns into a no-op on existing tables:
   - `store_row_id` was added in `4f48f8be` (2025-06-06, marked `[BREAKING CHANGE]`). Existing deployments must `ALTER` by hand.
   - The `stored_at` minmax index was added in `8243d0a5` (2026-03-18). Existing tables never get it, because nothing runs `ADD INDEX` or `MATERIALIZE INDEX`.
   - Per-meter materialized views were removed in `00da3356` (2025-03-12).
3. **No cluster mode.** The engine is hard-coded `MergeTree`. There is no
   `ON CLUSTER`, no `Replicated*` or `Distributed` engine, no Keeper path and no
   engine option in config (`app/config/aggregation.go:13-82`).
4. **Postgres uses Ent → Atlas → golang-migrate.**
   - `atlas.hcl` sets `migrations_format = "golang-migrate"`.
   - Files are named `YYYYMMDDHHMMSS_<name>.up.sql` / `.down.sql` under `tools/migrate/migrations`, embedded with `//go:embed`.
   - The version table is `schema_om` (`tools/migrate/migrate.go:23`).
   - Modes: `migration` (default), `migration-job`, `false`.
   - CI runs `atlas migrate diff/lint/validate`.
5. **Client and insert path.**
   - Client: `clickhouse-go/v2 v2.48.0`, native protocol, no compression set.
   - Inserts are a multi-row `INSERT ... VALUES` with optional `SETTINGS` from `InsertQuerySettings`.
   - `AsyncInsert` and `AsyncInsertWait` both default to false. With `AsyncInsert` on, the call is `clickhouse.WithAsync(wait)`.
   - The sink assigns a fresh ULID `store_row_id` on every attempt and retries a failed batch whole. It never uses `insert_deduplication_token`.
   - Deduplication happens before the insert, in an optional memory or Redis deduplicator.
6. **Deployment guidance.**
   - The Helm chart can install a single-node `ClickHouseInstallation` through the Altinity operator (`clickhouse/clickhouse-server:24.10`, chart dependency `altinity-clickhouse-operator 0.23.3`), labelled "Not recommended for production environments".
   - Compose uses `clickhouse/clickhouse-server:25.12.3-alpine`.
   - There is no guidance on Cloud or on a clustered topology.

## Current DDL (verbatim, `event_query_test.go:24`)

```sql
CREATE TABLE IF NOT EXISTS openmeter.om_events (namespace String, id String, type LowCardinality(String), subject String, source String, time DateTime, data String, ingested_at DateTime, stored_at DateTime, INDEX om_events_stored_at stored_at TYPE minmax GRANULARITY 4, store_row_id String) ENGINE = MergeTree PARTITION BY toYYYYMM(time) ORDER BY (namespace, type, subject, toStartOfHour(time))
```

## Implications for this evaluation (hypotheses to measure)

- **Every server under test changes OpenMeter's write semantics.** Since 26.2,
  `async_insert` and insert deduplication are on by default. That covers both
  26.3 builds and Cloud 26.4/26.6. OpenMeter never sets `async_insert`, so its
  synchronous inserts silently become async inserts with
  `wait_for_async_insert=1`.
- **Deduplication cannot catch OpenMeter's retries.** Engine-level dedup uses
  block hashes, so it only catches byte-identical retries. OpenMeter's retries
  carry new `store_row_id`s and are never byte-identical.
- **Any non-single-node target needs DDL templating** (engine and `ON CLUSTER`),
  because OpenMeter's hard-coded `MergeTree` is what Cloud auto-converts to
  SharedMergeTree. On a self-managed cluster it would create independent,
  unreplicated per-replica tables.

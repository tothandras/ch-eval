# OTel schemas: exporter defaults vs ClickStack

Verbatim DDL lives in [`otel-ddl/`](otel-ddl/). Each file's header cites its
source file and commit. Two sources:

- **Exporter defaults:** `open-telemetry/opentelemetry-collector-contrib`
  `exporter/clickhouseexporter/internal/sqltemplates/*`. The DDL is byte-identical
  between release v0.162.0 (2026-09-29) and main at `809052cc`.
- **ClickStack:** `hyperdxio/hyperdx` `docker/otel-collector/schema/seed/*.sql`
  at `eeabaebe` (2026-10-07), applied by ClickStack's own goose-based migrator.
  ClickStack's collector runs with `create_schema: false`.

## Owning the DDL is the documented recommendation

- Exporter README: "for production workloads, it is recommended that you manage
  your own schema by setting `create_schema` to `false` in the config. This
  prevents each exporter process from racing to create the database and
  tables, and makes it easier to upgrade the exporter in the future."
- ClickHouse docs (integrating OpenTelemetry): "We recommend users disable auto
  schema creation and create their tables manually."
- ClickHouse docs (schema design): "We recommend users always create their own
  schema for logs and traces". The page also calls the default map bloom
  indexes "generally ineffective".

Every OTel scenario here therefore runs the collector with
`create_schema: false` and applies its DDL through `migrations/`.

## Differences that matter for the benchmark

| Signal | Exporter default | ClickStack |
| --- | --- | --- |
| Traces | `MergeTree`, `PARTITION BY toDate(Timestamp)`, `ORDER BY (ServiceName, SpanName, toDateTime(Timestamp))`; bloom filters on TraceId and on map keys/values; separate `otel_traces_trace_id_ts` table plus MV for TraceId → time-range lookup | Same sort and partition key; ALIAS `*AttributeItems` (`k=v`) arrays with `text(tokenizer='array')` indexes; `text` index on `lower(SpanName)`; extra materialized columns; **no** trace_id_ts table |
| Logs | `ORDER BY (toStartOfFiveMinutes(Timestamp), ServiceName, Timestamp)`; on ≥26.2 `text` indexes on TraceId, map keys/values and `lower(Body)`, otherwise bloom + `tokenbf_v1` | Same keys; `*AttributeItems` with text(array) indexes; `enable_block_number_column`/`enable_block_offset_column` |
| Metrics (5 tables) | `ORDER BY (ServiceName, MetricName, toStartOfHour(TimeUnix), cityHash64(Attributes), TimeUnix)`, bloom on map keys/values | Nearly identical: exemplars flattened into arrays, minor codec changes |
| JSON | `json: true` exporter option switches logs and traces attribute maps to the `JSON` type (metrics stay Map) | Recommends Map over JSON; JSON support is beta |

All targets run ≥ 26.2 (both 26.3 builds and Cloud 26.4/26.6), so the
`text`-index variants of both schemas apply everywhere. The `<26.2` compat files
are kept for reference only.

## ClickStack performance-tuning levers (candidates for the "tuned" variants)

Source: https://clickhouse.com/docs/use-cases/observability/clickstack/performance_tuning

1. Materialize hot attributes into columns, then backfill with `MATERIALIZE COLUMN`.
2. Skip indexes: bloom for high-cardinality strings; `text` indexes on 26.2+; minmax for numerics.
3. Map direct read: `k=v` ALIAS arrays with text(array) indexes.
4. Keep the primary key unless access patterns clearly differ.
5. Materialized views for aggregation-heavy charts.
6. Projections, for example re-sorting by `TraceId` (about 2× write I/O).

Also recommended: always set `ttl_only_drop_parts=1`
(https://clickhouse.com/docs/use-cases/observability/clickstack/ttl).

## Exporter ingest-path defaults

| Option | Default |
| --- | --- |
| `async_insert` | `true`, added to the DSN; `wait_for_async_insert` is left at the server default |
| `compress` | lz4 |
| `timeout` | 5s |
| `sending_queue` | on, 10 consumers, queue size 1000 |
| `sending_queue.batch` | off. The README recommends it over the batch processor. ClickStack and the ClickHouse docs use the batch processor (10000 rows / 5s). |

## Discrepancies noticed in the sources

- The README names the exponential-histogram table `otel_metrics_exp_histogram`, but `config.go` creates `otel_metrics_exponential_histogram`.
- The traces JSON-mode INSERT template has 3 placeholders for 2 added columns. JSON-mode traces ingest must be checked before use.
- The traces DDL on the docs schemas page is an older revision than hyperdx main. The repo uses hyperdx main.

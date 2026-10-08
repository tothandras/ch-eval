-- ClickStack DDL: ClickStack otel_traces (ClickHouse >= 26.2, text indexes)
-- Source: https://github.com/hyperdxio/hyperdx/blob/eeabaebe6e9226f5ca090b29566e44ec52523b01/docker/otel-collector/schema/seed/00005_otel_traces.sql
-- Version: hyperdxio/hyperdx main @ eeabaebe6e9226f5ca090b29566e44ec52523b01 (2026-10-07)
-- Applied by the ClickStack collector image's goose-based `migrate` tool (packages/otel-collector/cmd/migrate/main.go)
--   before the collector starts; the exporter itself runs with create_schema: false unless JSON/legacy mode is enabled.
-- Macros: ${DATABASE} (env HYPERDX_OTEL_EXPORTER_CLICKHOUSE_DATABASE, default 'default');
--   ${LOGS_TTL}/${TRACES_TTL}/${METRICS_TTL} (env HYPERDX_OTEL_EXPORTER_{LOGS,TRACES,METRICS}_TTL, falling back to
--   HYPERDX_OTEL_EXPORTER_TABLES_TTL, default 720h) rendered as toIntervalDay(30) etc.
-- Only `-- +goose ...` annotation lines were removed; SQL text is verbatim.
-- NOTE: newer than the DDL rendered on the docs schemas page (see docs_clickstack_otel_traces.sql).
-- ===== BEGIN VERBATIM =====
CREATE TABLE IF NOT EXISTS ${DATABASE}.otel_traces
(
    `Timestamp` DateTime64(9) CODEC(Delta(8), ZSTD(1)),
    `TraceId` String CODEC(ZSTD(1)),
    `SpanId` String CODEC(ZSTD(1)),
    `ParentSpanId` String CODEC(ZSTD(1)),
    `TraceState` String CODEC(ZSTD(1)),
    `SpanName` LowCardinality(String) CODEC(ZSTD(1)),
    `SpanKind` LowCardinality(String) CODEC(ZSTD(1)),
    `ServiceName` LowCardinality(String) CODEC(ZSTD(1)),
    `ResourceAttributes` Map(LowCardinality(String), String) CODEC(ZSTD(1)),
    `ScopeName` String CODEC(ZSTD(1)),
    `ScopeVersion` String CODEC(ZSTD(1)),
    `SpanAttributes` Map(LowCardinality(String), String) CODEC(ZSTD(1)),
    `Duration` UInt64 CODEC(ZSTD(1)),
    `StatusCode` LowCardinality(String) CODEC(ZSTD(1)),
    `StatusMessage` String CODEC(ZSTD(1)),
    `Events.Timestamp` Array(DateTime64(9)) CODEC(ZSTD(1)),
    `Events.Name` Array(LowCardinality(String)) CODEC(ZSTD(1)),
    `Events.Attributes` Array(Map(LowCardinality(String), String)) CODEC(ZSTD(1)),
    `Links.TraceId` Array(String) CODEC(ZSTD(1)),
    `Links.SpanId` Array(String) CODEC(ZSTD(1)),
    `Links.TraceState` Array(String) CODEC(ZSTD(1)),
    `Links.Attributes` Array(Map(LowCardinality(String), String)) CODEC(ZSTD(1)),
    `__hdx_materialized_rum.sessionId` String MATERIALIZED ResourceAttributes['rum.sessionId'] CODEC(ZSTD(1)),
    `SampleRate` UInt64 MATERIALIZED greatest(toUInt64OrZero(SpanAttributes['SampleRate']), 1) CODEC(T64, ZSTD(1)),
    `ResourceAttributeItems` Array(String) ALIAS arrayMap((arr) -> concat(arr.1, '=', arr.2), ResourceAttributes::Array(Tuple(String, String))),
    `SpanAttributeItems` Array(String) ALIAS arrayMap((arr) -> concat(arr.1, '=', arr.2), SpanAttributes::Array(Tuple(String, String))),
    INDEX idx_trace_id TraceId TYPE text(tokenizer = 'array'),
    INDEX idx_rum_session_id __hdx_materialized_rum.sessionId TYPE text(tokenizer = 'array'),
    INDEX idx_res_attr_key mapKeys(ResourceAttributes) TYPE text(tokenizer = 'array'),
    INDEX idx_res_attr_items ResourceAttributeItems TYPE text(tokenizer = 'array'),
    INDEX idx_span_attr_key mapKeys(SpanAttributes) TYPE text(tokenizer = 'array'),
    INDEX idx_span_attr_items SpanAttributeItems TYPE text(tokenizer = 'array'),
    INDEX idx_duration Duration TYPE minmax GRANULARITY 1,
    INDEX idx_lower_span_name SpanName TYPE text(tokenizer = 'splitByNonAlpha', preprocessor=lower(SpanName))
)
ENGINE = MergeTree
PARTITION BY toDate(Timestamp)
ORDER BY (ServiceName, SpanName, toDateTime(Timestamp))
TTL toDate(Timestamp) + ${TRACES_TTL}
SETTINGS index_granularity = 8192, ttl_only_drop_parts = 1;
-- ===== END VERBATIM =====

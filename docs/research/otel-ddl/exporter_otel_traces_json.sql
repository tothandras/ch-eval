-- Exporter-default DDL template: otel_traces (JSON attributes; exporter config json: true)
-- Source: https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/809052cc7cbee29653943daf22c8b6b8a76ea5b2/exporter/clickhouseexporter/internal/sqltemplates/traces_json_table.sql
-- Version: opentelemetry-collector-contrib main @ 809052cc7cbee29653943daf22c8b6b8a76ea5b2 (2026-10-07). internal/sqltemplates, config.go and exporter_*.go are byte-identical to release tag v0.162.0 (ae8c507510f48f433ab47dd1c6b01a59d6c388b5, 2026-09-29), the latest release tag.
-- Placeholders (Go fmt): %q.%q = database, table (Go-quoted, e.g. "default"."otel_traces"); 3rd %s = ClusterString ('' or ON CLUSTER `name`); ENGINE = %s -> TableEngine (default MergeTree()); last %s = TTL expr ('' or TTL toDateTime(Timestamp) + toIntervalDay(N)/Hour/Minute/Second).
-- Used when config `json: true` (createTraceJSONTables in exporter_traces_json.go). Also creates trace_id_ts table + MV.
-- README: may need enable_json_type=1 in endpoint/connection_params; ClickHouse v25+ recommended.
-- ===== BEGIN VERBATIM =====
CREATE TABLE IF NOT EXISTS %q.%q %s (
    Timestamp DateTime64(9) CODEC(Delta, ZSTD(1)),
    TraceId String CODEC(ZSTD(1)),
    SpanId String CODEC(ZSTD(1)),
    ParentSpanId String CODEC(ZSTD(1)),
    TraceState String CODEC(ZSTD(1)),
    SpanName LowCardinality(String) CODEC(ZSTD(1)),
    SpanKind LowCardinality(String) CODEC(ZSTD(1)),
    ServiceName LowCardinality(String) CODEC(ZSTD(1)),
    ResourceAttributes JSON CODEC(ZSTD(1)),
    ResourceAttributesKeys Array(LowCardinality(String)) CODEC(ZSTD(1)),
    ScopeName String CODEC(ZSTD(1)),
    ScopeVersion String CODEC(ZSTD(1)),
    SpanAttributes JSON CODEC(ZSTD(1)),
    SpanAttributesKeys Array(LowCardinality(String)) CODEC(ZSTD(1)),
    Duration UInt64 CODEC(ZSTD(1)),
    StatusCode LowCardinality(String) CODEC(ZSTD(1)),
    StatusMessage String CODEC(ZSTD(1)),
    Events Nested (
        Timestamp DateTime64(9),
        Name LowCardinality(String),
        Attributes JSON
    ) CODEC(ZSTD(1)),
    Links Nested (
        TraceId String,
        SpanId String,
        TraceState String,
        Attributes JSON
    ) CODEC(ZSTD(1)),

    INDEX idx_res_attr_keys ResourceAttributesKeys TYPE bloom_filter(0.01) GRANULARITY 1,
    INDEX idx_span_attr_keys SpanAttributesKeys TYPE bloom_filter(0.01) GRANULARITY 1,
    INDEX idx_duration Duration TYPE minmax GRANULARITY 1
) ENGINE = %s
PARTITION BY toDate(Timestamp)
ORDER BY (ServiceName, SpanName, toDateTime(Timestamp), Timestamp)
%s
SETTINGS index_granularity=8192, ttl_only_drop_parts = 1
-- ===== END VERBATIM =====

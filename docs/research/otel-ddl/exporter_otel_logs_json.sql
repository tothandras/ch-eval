-- Exporter-default DDL template: otel_logs (JSON attributes; exporter config json: true)
-- Source: https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/809052cc7cbee29653943daf22c8b6b8a76ea5b2/exporter/clickhouseexporter/internal/sqltemplates/logs_json_table.sql
-- Version: opentelemetry-collector-contrib main @ 809052cc7cbee29653943daf22c8b6b8a76ea5b2 (2026-10-07). internal/sqltemplates, config.go and exporter_*.go are byte-identical to release tag v0.162.0 (ae8c507510f48f433ab47dd1c6b01a59d6c388b5, 2026-09-29), the latest release tag.
-- Placeholders (Go text/template): {{ident .Database}}.{{ident .TableName}} -> backtick-quoted identifiers; {{.ClusterString}} ('' or ON CLUSTER `name`); {{.Engine}} (default MergeTree()); {{.TTL}} ('' or TTL toDateTime(Timestamp) + toIntervalDay(N)...).
-- Used when config `json: true` (createLogsJSONTable in exporter_logs_json.go). Note PRIMARY KEY differs from ORDER BY here.
-- ===== BEGIN VERBATIM =====
CREATE TABLE IF NOT EXISTS {{ident .Database}}.{{ident .TableName}} {{.ClusterString}} (
    `Timestamp` DateTime64(9) CODEC(Delta(8), ZSTD(1)),
    `TraceId` String CODEC(ZSTD(1)),
    `SpanId` String CODEC(ZSTD(1)),
    `TraceFlags` UInt8,
    `SeverityText` LowCardinality(String) CODEC(ZSTD(1)),
    `SeverityNumber` UInt8,
    `ServiceName` LowCardinality(String) CODEC(ZSTD(1)),
    `Body` String CODEC(ZSTD(1)),
    `ResourceSchemaUrl` LowCardinality(String) CODEC(ZSTD(1)),
    `ResourceAttributes` JSON CODEC(ZSTD(1)),
    `ResourceAttributesKeys` Array(LowCardinality(String)) CODEC(ZSTD(1)),
    `ScopeSchemaUrl` LowCardinality(String) CODEC(ZSTD(1)),
    `ScopeName` String CODEC(ZSTD(1)),
    `ScopeVersion` LowCardinality(String) CODEC(ZSTD(1)),
    `ScopeAttributes` JSON CODEC(ZSTD(1)),
    `ScopeAttributesKeys` Array(LowCardinality(String)) CODEC(ZSTD(1)),
    `LogAttributes` JSON CODEC(ZSTD(1)),
    `LogAttributesKeys` Array(LowCardinality(String)) CODEC(ZSTD(1)),
    `EventName` String CODEC(ZSTD(1)),

    INDEX idx_res_attr_keys ResourceAttributesKeys TYPE bloom_filter(0.01) GRANULARITY 1,
    INDEX idx_scope_attr_keys ScopeAttributesKeys TYPE bloom_filter(0.01) GRANULARITY 1,
    INDEX idx_log_attr_keys LogAttributesKeys TYPE bloom_filter(0.01) GRANULARITY 1,
    INDEX idx_body Body TYPE tokenbf_v1(32768, 3, 0) GRANULARITY 8
) ENGINE = {{.Engine}}
PARTITION BY toDate(Timestamp)
PRIMARY KEY (toStartOfFiveMinutes(Timestamp), ServiceName)
ORDER BY (toStartOfFiveMinutes(Timestamp), ServiceName, Timestamp)
{{.TTL}}
SETTINGS index_granularity = 8192, ttl_only_drop_parts = 1
-- ===== END VERBATIM =====

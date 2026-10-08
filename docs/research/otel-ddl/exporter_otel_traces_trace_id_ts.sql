-- Exporter-default DDL template: otel_traces_trace_id_ts lookup table + otel_traces_trace_id_ts_mv materialized view
-- Source (table): https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/809052cc7cbee29653943daf22c8b6b8a76ea5b2/exporter/clickhouseexporter/internal/sqltemplates/traces_id_ts_lookup_table.sql
-- Source (MV):    https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/809052cc7cbee29653943daf22c8b6b8a76ea5b2/exporter/clickhouseexporter/internal/sqltemplates/traces_id_ts_lookup_mv.sql
-- Version: opentelemetry-collector-contrib main @ 809052cc7cbee29653943daf22c8b6b8a76ea5b2 (2026-10-07). internal/sqltemplates, config.go and exporter_*.go are byte-identical to release tag v0.162.0 (ae8c507510f48f433ab47dd1c6b01a59d6c388b5, 2026-09-29), the latest release tag.
-- Table placeholders (renderCreateTraceIDTsTableSQL): "%s"."%s_trace_id_ts" = database, traces table name; %s = ClusterString;
--   ENGINE = %s -> TableEngine; last %s = TTL expr on toDateTime(Start).
-- MV placeholders (renderTraceIDTsMaterializedViewSQL): db, traces table, ClusterString, db, traces table, db, traces table.
-- The two upstream files are concatenated below, separated by ';' + blank line (upstream files have no trailing ';').
-- ===== BEGIN VERBATIM =====
CREATE TABLE IF NOT EXISTS "%s"."%s_trace_id_ts" %s (
    TraceId String CODEC(ZSTD(1)),
    Start DateTime CODEC(Delta, ZSTD(1)),
    End DateTime CODEC(Delta, ZSTD(1)),
    INDEX idx_trace_id TraceId TYPE bloom_filter(0.01) GRANULARITY 1
) ENGINE = %s
    PARTITION BY toDate(Start)
    ORDER BY (TraceId, Start)
    %s
    SETTINGS index_granularity=8192, ttl_only_drop_parts = 1;

CREATE MATERIALIZED VIEW IF NOT EXISTS "%s"."%s_trace_id_ts_mv" %s
TO "%s"."%s_trace_id_ts"
AS SELECT
              TraceId,
              min(Timestamp) as Start,
              max(Timestamp) as End
   FROM "%s"."%s"
   WHERE TraceId != ''
   GROUP BY TraceId;
-- ===== END VERBATIM =====

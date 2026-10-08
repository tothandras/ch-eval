-- Exporter INSERT statement templates (all signals)
-- Source dir: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/809052cc7cbee29653943daf22c8b6b8a76ea5b2/exporter/clickhouseexporter/internal/sqltemplates
-- Version: opentelemetry-collector-contrib main @ 809052cc7cbee29653943daf22c8b6b8a76ea5b2 (2026-10-07). internal/sqltemplates, config.go and exporter_*.go are byte-identical to release tag v0.162.0 (ae8c507510f48f433ab47dd1c6b01a59d6c388b5, 2026-09-29), the latest release tag.
-- Each upstream file is reproduced verbatim; a trailing ';' was appended after each for readability.
-- Logs: {{.FeatureColumnNames}}/{{.FeatureColumnPositions}} add ', EventName' / ', ?' if the column exists (DESC TABLE at start);
--   JSON logs additionally add ResourceAttributesKeys, ScopeAttributesKeys, LogAttributesKeys when present.
-- Traces JSON: trailing %s,%s add ', ResourceAttributesKeys, SpanAttributesKeys' and ', ?, ?, ?' (3 placeholders for 2 columns
--   upstream, exporter_traces_json.go renderInsertTracesJSONSQL).
-- Inserts are sent with clickhouse-go PrepareBatch/Append/Send (one batch INSERT per exporter push).
-- Metric/traces templates use fmt %s/%q for database and table.
-- ===== BEGIN VERBATIM =====
-- ---- traces_insert.sql  (https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/809052cc7cbee29653943daf22c8b6b8a76ea5b2/exporter/clickhouseexporter/internal/sqltemplates/traces_insert.sql)
INSERT INTO %q.%q (
    Timestamp,
    TraceId,
    SpanId,
    ParentSpanId,
    TraceState,
    SpanName,
    SpanKind,
    ServiceName,
    ResourceAttributes,
    ScopeName,
    ScopeVersion,
    SpanAttributes,
    Duration,
    StatusCode,
    StatusMessage,
    Events.Timestamp,
    Events.Name,
    Events.Attributes,
    Links.TraceId,
    Links.SpanId,
    Links.TraceState,
    Links.Attributes
) VALUES (
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?
);

-- ---- traces_json_insert.sql  (https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/809052cc7cbee29653943daf22c8b6b8a76ea5b2/exporter/clickhouseexporter/internal/sqltemplates/traces_json_insert.sql)
INSERT INTO %q.%q (
    Timestamp,
    TraceId,
    SpanId,
    ParentSpanId,
    TraceState,
    SpanName,
    SpanKind,
    ServiceName,
    ResourceAttributes,
    ScopeName,
    ScopeVersion,
    SpanAttributes,
    Duration,
    StatusCode,
    StatusMessage,
    Events.Timestamp,
    Events.Name,
    Events.Attributes,
    Links.TraceId,
    Links.SpanId,
    Links.TraceState,
    Links.Attributes
    %s
) VALUES (
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?
    %s
);

-- ---- logs_insert.sql  (https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/809052cc7cbee29653943daf22c8b6b8a76ea5b2/exporter/clickhouseexporter/internal/sqltemplates/logs_insert.sql)
INSERT INTO {{ident .Database}}.{{ident .TableName}} (
    Timestamp,
    TraceId,
    SpanId,
    TraceFlags,
    SeverityText,
    SeverityNumber,
    ServiceName,
    Body,
    ResourceSchemaUrl,
    ResourceAttributes,
    ScopeSchemaUrl,
    ScopeName,
    ScopeVersion,
    ScopeAttributes,
    LogAttributes
    {{.FeatureColumnNames}}
) VALUES (
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?
    {{.FeatureColumnPositions}}
);

-- ---- logs_json_insert.sql  (https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/809052cc7cbee29653943daf22c8b6b8a76ea5b2/exporter/clickhouseexporter/internal/sqltemplates/logs_json_insert.sql)
INSERT INTO {{ident .Database}}.{{ident .TableName}} (
    Timestamp,
    TraceId,
    SpanId,
    TraceFlags,
    SeverityText,
    SeverityNumber,
    ServiceName,
    Body,
    ResourceSchemaUrl,
    ResourceAttributes,
    ScopeSchemaUrl,
    ScopeName,
    ScopeVersion,
    ScopeAttributes,
    LogAttributes
    {{.FeatureColumnNames}}
) VALUES (
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?
    {{.FeatureColumnPositions}}
);

-- ---- metrics_gauge_insert.sql  (https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/809052cc7cbee29653943daf22c8b6b8a76ea5b2/exporter/clickhouseexporter/internal/sqltemplates/metrics_gauge_insert.sql)
INSERT INTO "%s"."%s" (
    ResourceAttributes,
    ResourceSchemaUrl,
    ScopeName,
    ScopeVersion,
    ScopeAttributes,
    ScopeDroppedAttrCount,
    ScopeSchemaUrl,
    ServiceName,
    MetricName,
    MetricDescription,
    MetricUnit,
    Attributes,
    StartTimeUnix,
    TimeUnix,
    Value,
    Flags,
    Exemplars.FilteredAttributes,
    Exemplars.TimeUnix,
    Exemplars.Value,
    Exemplars.SpanId,
    Exemplars.TraceId) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?);

-- ---- metrics_sum_insert.sql  (https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/809052cc7cbee29653943daf22c8b6b8a76ea5b2/exporter/clickhouseexporter/internal/sqltemplates/metrics_sum_insert.sql)
INSERT INTO "%s"."%s" (
    ResourceAttributes,
    ResourceSchemaUrl,
    ScopeName,
    ScopeVersion,
    ScopeAttributes,
    ScopeDroppedAttrCount,
    ScopeSchemaUrl,
    ServiceName,
    MetricName,
    MetricDescription,
    MetricUnit,
    Attributes,
    StartTimeUnix,
    TimeUnix,
    Value,
    Flags,
    Exemplars.FilteredAttributes,
    Exemplars.TimeUnix,
    Exemplars.Value,
    Exemplars.SpanId,
    Exemplars.TraceId,
    AggregationTemporality,
    IsMonotonic) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?);

-- ---- metrics_histogram_insert.sql  (https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/809052cc7cbee29653943daf22c8b6b8a76ea5b2/exporter/clickhouseexporter/internal/sqltemplates/metrics_histogram_insert.sql)
INSERT INTO "%s"."%s" (
    ResourceAttributes,
    ResourceSchemaUrl,
    ScopeName,
    ScopeVersion,
    ScopeAttributes,
    ScopeDroppedAttrCount,
    ScopeSchemaUrl,
    ServiceName,
    MetricName,
    MetricDescription,
    MetricUnit,
    Attributes,
    StartTimeUnix,
    TimeUnix,
    Count,
    Sum,
    BucketCounts,
    ExplicitBounds,
    Exemplars.FilteredAttributes,
    Exemplars.TimeUnix,
    Exemplars.Value,
    Exemplars.SpanId,
    Exemplars.TraceId,
    Flags,
    Min,
    Max,
    AggregationTemporality) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?);

-- ---- metrics_exp_histogram_insert.sql  (https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/809052cc7cbee29653943daf22c8b6b8a76ea5b2/exporter/clickhouseexporter/internal/sqltemplates/metrics_exp_histogram_insert.sql)
INSERT INTO "%s"."%s" (
    ResourceAttributes,
    ResourceSchemaUrl,
    ScopeName,
    ScopeVersion,
    ScopeAttributes,
    ScopeDroppedAttrCount,
    ScopeSchemaUrl,
    ServiceName,
    MetricName,
    MetricDescription,
    MetricUnit,
    Attributes,
    StartTimeUnix,
    TimeUnix,
    Count,
    Sum,
    Scale,
    ZeroCount,
    PositiveOffset,
    PositiveBucketCounts,
    NegativeOffset,
    NegativeBucketCounts,
    Exemplars.FilteredAttributes,
    Exemplars.TimeUnix,
    Exemplars.Value,
    Exemplars.SpanId,
    Exemplars.TraceId,
    Flags,
    Min,
    Max,
    AggregationTemporality) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?);

-- ---- metrics_summary_insert.sql  (https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/809052cc7cbee29653943daf22c8b6b8a76ea5b2/exporter/clickhouseexporter/internal/sqltemplates/metrics_summary_insert.sql)
INSERT INTO "%s"."%s" (
    ResourceAttributes,
    ResourceSchemaUrl,
    ScopeName,
    ScopeVersion,
    ScopeAttributes,
    ScopeDroppedAttrCount,
    ScopeSchemaUrl,
    ServiceName,
    MetricName,
    MetricDescription,
    MetricUnit,
    Attributes,
    StartTimeUnix,
    TimeUnix,
    Count,
    Sum,
    ValueAtQuantiles.Quantile,
    ValueAtQuantiles.Value,
    Flags) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?);

-- ===== END VERBATIM =====

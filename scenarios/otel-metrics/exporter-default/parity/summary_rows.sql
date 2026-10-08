-- Every column the exporter inserts, hashed per row and summed (order-independent).
SELECT count() AS rows, sum(cityHash64(ResourceAttributes, ResourceSchemaUrl, ScopeName, ScopeVersion, ScopeAttributes, ScopeDroppedAttrCount, ScopeSchemaUrl, ServiceName, MetricName, MetricDescription, MetricUnit, Attributes, StartTimeUnix, TimeUnix, Count, Sum, ValueAtQuantiles.Quantile, ValueAtQuantiles.Value, Flags)) AS fingerprint
FROM otel_metrics_summary

-- Every column the exporter inserts, hashed per row and summed (order-independent).
SELECT count() AS rows, sum(cityHash64(Timestamp, TraceId, SpanId, TraceFlags, SeverityText, SeverityNumber, ServiceName, Body, ResourceSchemaUrl, ResourceAttributes, ScopeSchemaUrl, ScopeName, ScopeVersion, ScopeAttributes, LogAttributes, EventName)) AS fingerprint
FROM otel_logs

-- Every column the exporter inserts, hashed per row and summed (order-independent).
SELECT count() AS rows, sum(cityHash64(Timestamp, TraceId, SpanId, ParentSpanId, TraceState, SpanName, SpanKind, ServiceName, ResourceAttributes, ScopeName, ScopeVersion, SpanAttributes, Duration, StatusCode, StatusMessage, Events.Timestamp, Events.Name, Events.Attributes, Links.TraceId, Links.SpanId, Links.TraceState, Links.Attributes)) AS fingerprint
FROM otel_traces

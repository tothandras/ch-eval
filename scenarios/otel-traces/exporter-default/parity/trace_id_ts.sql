-- The TraceId -> time-range view, merged per TraceId so the result does not depend
-- on how inserts were split into blocks (the view aggregates per insert block).
SELECT count() AS traces, sum(cityHash64(TraceId, s, e)) AS fingerprint
FROM (SELECT TraceId, min(Start) AS s, max(End) AS e FROM otel_traces_trace_id_ts GROUP BY TraceId)

-- The key/value rollup, summed per key so the result does not depend on insert
-- block boundaries or on whether SummingMergeTree has merged yet.
SELECT count() AS rows, sum(cityHash64(Timestamp, ColumnIdentifier, Key, Value, c)) AS fingerprint
FROM (SELECT Timestamp, ColumnIdentifier, Key, Value, sum(count) AS c FROM otel_logs_kv_rollup_15m GROUP BY Timestamp, ColumnIdentifier, Key, Value)

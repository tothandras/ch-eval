-- probe: Refreshable materialized view (VG-08)
-- expect: ok
CREATE MATERIALIZED VIEW {db}.rmv_probe REFRESH EVERY 1 HOUR ENGINE = MergeTree ORDER BY x AS SELECT 1 AS x

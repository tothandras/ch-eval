-- probe: A retried identical async insert is deduplicated by default since 26.2 (IN-03)
-- requires: replicated
-- expect: value:1
CREATE TABLE {db}.dedup_async_probe (k UInt8) ENGINE = ReplicatedMergeTree('/clickhouse/tables/ch_eval_probe/{database}/{table}', '{replica}') ORDER BY k;
INSERT INTO {db}.dedup_async_probe SETTINGS async_insert = 1, wait_for_async_insert = 1 VALUES (1);
INSERT INTO {db}.dedup_async_probe SETTINGS async_insert = 1, wait_for_async_insert = 1 VALUES (1);
SELECT count() FROM {db}.dedup_async_probe

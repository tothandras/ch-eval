-- probe: A retried identical synchronous insert is deduplicated on replicated storage (IN-03)
-- requires: replicated
-- expect: value:1
CREATE TABLE {db}.dedup_sync_probe (k UInt8) ENGINE = ReplicatedMergeTree('/clickhouse/tables/ch_eval_probe/{database}/{table}', '{replica}') ORDER BY k;
INSERT INTO {db}.dedup_sync_probe SETTINGS async_insert = 0 VALUES (1);
INSERT INTO {db}.dedup_sync_probe SETTINGS async_insert = 0 VALUES (1);
SELECT count() FROM {db}.dedup_sync_probe

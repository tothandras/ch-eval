-- probe: Plain MergeTree keeps a retried identical insert (non_replicated_deduplication_window = 0) (IN-04)
-- expect: value:2
-- expect.cloud: value:1
CREATE TABLE {db}.dedup_plain_probe (k UInt8) ENGINE = MergeTree ORDER BY k;
INSERT INTO {db}.dedup_plain_probe SETTINGS async_insert = 0 VALUES (1);
INSERT INTO {db}.dedup_plain_probe SETTINGS async_insert = 0 VALUES (1);
SELECT count() FROM {db}.dedup_plain_probe

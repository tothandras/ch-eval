-- probe: Quorum inserts with insert_quorum_parallel=0 combined with async inserts are rejected (RS-04); async inserts are the 26.2+ default. Via SQL VALUES the error is 2; via the native batch path clickhouse-go uses, 26.3 raised LOGICAL_ERROR 49 ("Should be checked earlier")
-- requires: replicated
-- expect: error
-- expect.cloud: ok || error
CREATE TABLE {db}.quorum_async_probe (k UInt8) ENGINE = ReplicatedMergeTree('/clickhouse/tables/ch_eval_probe/{database}/{table}', '{replica}') ORDER BY k;
INSERT INTO {db}.quorum_async_probe SETTINGS async_insert = 1, wait_for_async_insert = 1, insert_quorum = 2, insert_quorum_parallel = 0 VALUES (1)

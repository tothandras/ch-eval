-- probe: Engine actually used for ENGINE = ReplicatedMergeTree (ST-01)
-- requires: replicated
-- expect: value:ReplicatedMergeTree
-- expect.cloud: value:SharedMergeTree
CREATE TABLE {db}.rengine_probe (x UInt8) ENGINE = ReplicatedMergeTree('/clickhouse/tables/ch_eval_probe/{database}/{table}', '{replica}') ORDER BY x;
SELECT engine FROM system.tables WHERE database = '{db}' AND name = 'rengine_probe'

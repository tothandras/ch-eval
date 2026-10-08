-- probe: Engine actually used for ENGINE = MergeTree (ST-01)
-- expect: value:MergeTree
-- expect.cloud: value:SharedMergeTree
CREATE TABLE {db}.engine_probe (x UInt8) ENGINE = MergeTree ORDER BY x;
SELECT engine FROM system.tables WHERE database = '{db}' AND name = 'engine_probe'

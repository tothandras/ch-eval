-- probe: ON CLUSTER DDL fans out to every node (ST-03)
-- requires: cluster
-- expect: value:{nodes}
CREATE TABLE {db}.on_cluster_probe ON CLUSTER '{cluster}' (x UInt8) ENGINE = MergeTree ORDER BY x;
SELECT count() FROM clusterAllReplicas('{cluster}', system.tables) WHERE database = '{db}' AND name = 'on_cluster_probe'

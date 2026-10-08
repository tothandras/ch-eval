-- probe: SYSTEM STOP/START MERGES on a table (SM-02; unsupported on Cloud)
-- expect: ok
-- expect.cloud: error
CREATE TABLE {db}.merges_probe (x UInt8) ENGINE = MergeTree ORDER BY x;
SYSTEM STOP MERGES {db}.merges_probe;
SYSTEM START MERGES {db}.merges_probe

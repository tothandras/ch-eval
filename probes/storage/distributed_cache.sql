-- probe: Distributed cache switches enabled by default (ST-09); 0 = none enabled
-- expect: value:0
-- expect.cloud: ok
SELECT countIf(value IN ('1', 'true')) FROM system.settings WHERE name LIKE '%distributed_cache%' AND (name LIKE 'enable%' OR name LIKE 'read_through%' OR name LIKE 'write_through%')

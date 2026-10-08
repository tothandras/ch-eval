-- probe: Parallel replicas default (VG-07): enable_parallel_replicas
-- expect: value:0
-- expect.cloud: ok
SELECT value FROM system.settings WHERE name = 'enable_parallel_replicas'

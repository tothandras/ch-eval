-- probe: Distributed insert and shard-pruning defaults (RS-08, RS-10)
-- expect: value:false	false	false
-- expect.cloud: ok
SELECT getSetting('distributed_foreground_insert'), getSetting('optimize_skip_unused_shards'), getSetting('skip_unavailable_shards')

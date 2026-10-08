-- probe: skip_unavailable_shards = 1 silently returns the reachable shards' data (RS-09)
-- expect: value:1
-- expect.cloud: value:1 || error
SELECT count() FROM remote('127.0.0.1:1,localhost:9000', system.one) SETTINGS skip_unavailable_shards = 1, connect_timeout_with_failover_ms = 200, connections_with_failover_max_tries = 1

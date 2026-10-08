-- probe: A query over a cluster with one unreachable shard fails by default (RS-09, skip_unavailable_shards = 0)
-- expect: error
SELECT count() FROM remote('127.0.0.1:1,localhost:9000', system.one) SETTINGS connect_timeout_with_failover_ms = 200, connections_with_failover_max_tries = 1

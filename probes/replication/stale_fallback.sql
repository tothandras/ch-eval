-- probe: Stale-replica fallback defaults for distributed reads (RS-02)
-- expect: value:['fallback_to_stale_replicas_for_distributed_queries=1','max_replica_delay_for_distributed_queries=300']
-- expect.cloud: ok
SELECT groupArray(name || '=' || value) FROM (SELECT name, value FROM system.settings WHERE name IN ('fallback_to_stale_replicas_for_distributed_queries', 'max_replica_delay_for_distributed_queries') ORDER BY name)

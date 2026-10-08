-- probe: Deduplication windows by engine family (IN-04)
-- expect: value:['non_replicated_deduplication_window=0','replicated_deduplication_window=10000']
-- expect.cloud: ok
SELECT groupArray(name || '=' || value) FROM (SELECT name, value FROM system.merge_tree_settings WHERE name IN ('replicated_deduplication_window', 'non_replicated_deduplication_window') ORDER BY name)

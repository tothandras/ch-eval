-- probe: Query condition cache on by default (VG-06)
-- expect: value:1 || value:true
SELECT value FROM system.settings WHERE name = 'use_query_condition_cache'

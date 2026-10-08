-- probe: Cache populated on write by default (CS-03)
-- expect: value:0 || value:false
-- expect.cloud: value:1 || value:true
SELECT value FROM system.settings WHERE name = 'enable_filesystem_cache_on_write_operations'

-- probe: Filesystem-cache bypass setting accepted (ST-07)
-- expect: value:1
SELECT 1 SETTINGS enable_filesystem_cache = 0

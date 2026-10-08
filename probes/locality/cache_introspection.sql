-- probe: Filesystem-cache introspection readable (CS-02): rows in system.filesystem_cache, FilesystemCacheSize metric present
-- expect: regex:^[0-9]+	1$
SELECT (SELECT count() FROM system.filesystem_cache), (SELECT count() FROM system.metrics WHERE metric = 'FilesystemCacheSize')

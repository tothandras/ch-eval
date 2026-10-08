-- probe: system.query_log, part_log and asynchronous_insert_log exist after a flush (OB-01..OB-03); log tables are created lazily
-- expect: value:1	1	1
SYSTEM FLUSH LOGS;
SELECT (SELECT count() FROM system.tables WHERE database = 'system' AND name = 'query_log'), (SELECT count() FROM system.tables WHERE database = 'system' AND name = 'part_log'), (SELECT count() FROM system.tables WHERE database = 'system' AND name = 'asynchronous_insert_log')

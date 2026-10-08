-- probe: system.query_log, part_log and asynchronous_insert_log exist (OB-01..OB-03)
-- expect: value:1	1	1
SELECT (SELECT count() FROM system.tables WHERE database = 'system' AND name = 'query_log'), (SELECT count() FROM system.tables WHERE database = 'system' AND name = 'part_log'), (SELECT count() FROM system.tables WHERE database = 'system' AND name = 'asynchronous_insert_log')

-- probe: CGroupMaxCPU reported (OB-04); no row means absent (harness falls back to max_threads)
-- expect: ok
SELECT value FROM system.asynchronous_metrics WHERE metric = 'CGroupMaxCPU'

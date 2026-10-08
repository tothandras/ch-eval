# S3: What happens when a shard is unavailable (default vs skip_unavailable_shards=1)?

Target `altinity/local-cluster-2x2` (`26.3.33.10001.altinitystable`), 2026-10-08T07:41:39Z.

> Local tier: behaviour (errors, partial results, catch-up) only; timings are not production evidence.

| Step | Outcome | ms | Detail |
| --- | --- | ---: | --- |
| insert 100k rows through the Distributed table (foreground) | ok | 104 |  |
| count with all shards up | value:100000 | 0 |  |
| stop node ch-s2r1 (shard 2) | ok | 5495 |  |
| stop node ch-s2r2 (shard 2) | ok | 5012 |  |
| count, default settings (skip_unavailable_shards=0) | error:279 | 8 | code: 279, message: All connection tries failed. Log: 

Code: 32. DB::Exception: Attempt to read after eof. (ATTEMPT_TO_READ_AFTER_EOF) (version 26.3.33.10001.altinitystable (altinity build))
Code: 32. DB::Exception: Attempt to read after e… |
| count, skip_unavailable_shards=1 | value:50088 | 1055 |  |

## Findings

- With a shard down, the default query returned error:279 after 8 ms.
- With skip_unavailable_shards=1 it returned value:50088 of 100000 rows, without an error, after 1055 ms: a partial total that a billing query would not notice.

# S4: How does insert routing affect visibility (Distributed background, foreground, direct to shard)?

Target `altinity/local-cluster-2x2` (`26.3.33.10001.altinitystable`), 2026-10-08T07:41:56Z.

> Local tier: behaviour (errors, partial results, catch-up) only; timings are not production evidence.

| Step | Outcome | ms | Detail |
| --- | --- | ---: | --- |
| insert ack: Distributed, background queue (default) | ok | 18 |  |
| time from ack to queryable: Distributed, background queue (default) | value:144ms | 144 |  |
| insert ack: Distributed, distributed_foreground_insert=1 | ok | 79 |  |
| time from ack to queryable: Distributed, distributed_foreground_insert=1 | value:45ms | 45 |  |
| insert ack: direct to shard-local tables (client-side sharding) | ok | 83 |  |
| time from ack to queryable: direct to shard | value:21ms | 21 |  |
| insert with the last shard down: background queue | ok | 25 |  |
| insert with the last shard down: foreground | error:279 | 10 | code: 279, message: All connection tries failed. Log: 

Code: 32. DB::Exception: Attempt to read after eof. (ATTEMPT_TO_READ_AFTER_EOF) (version 26.3.33.10001.altinitystable (altinity build))
Code: 32. DB::Exception: Attempt to read after e… |
| distribution_queue files pending on the initiator | value:1 | 0 |  |
| time from shard restart to queued rows queryable | value:1520ms | 1520 |  |

## Findings

- Background Distributed inserts acknowledge before rows are queryable; foreground inserts acknowledge after the shards have them.
- With a shard down, the background insert was accepted (true) and queued on the initiator's disk (1 files) until the shard returned; the foreground insert returned error:279.

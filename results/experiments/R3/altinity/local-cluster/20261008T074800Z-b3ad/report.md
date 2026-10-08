# R3: What happens when Keeper loses quorum?

Target `altinity/local-cluster` (`26.3.33.10001.altinitystable`), 2026-10-08T07:47:37Z.

> Local tier: behaviour (errors, partial results, catch-up) only; timings are not production evidence.

| Step | Outcome | ms | Detail |
| --- | --- | ---: | --- |
| insert, 3/3 Keepers up | ok | 12 |  |
| insert, 2/3 Keepers up (one lost) | ok | 11 |  |
| insert, 1/3 Keepers up (quorum lost) | error:242 | 309 | code: 242, message: Table is in readonly mode: replica_path=/clickhouse/tables/1/ch_eval_exp/kq/replicas/ch-s1r1 |
| SELECT during quorum loss | value:2 | 5 |  |
| system.replicas.is_readonly during quorum loss | value:1 | 0 |  |
| time from Keeper restart until inserts succeed | value:8618ms | 8618 |  |

## Findings

- Losing one of three Keepers: insert ok.
- Losing quorum: insert error:242 after 309 ms; SELECT value:2; tables went read-only (is_readonly=1). Inserts recovered 8618ms after the Keepers restarted.

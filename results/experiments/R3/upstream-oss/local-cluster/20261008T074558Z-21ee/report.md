# R3: What happens when Keeper loses quorum?

Target `upstream-oss/local-cluster` (`26.3.33.24`), 2026-10-08T07:45:43Z.

> Local tier: behaviour (errors, partial results, catch-up) only; timings are not production evidence.

| Step | Outcome | ms | Detail |
| --- | --- | ---: | --- |
| insert, 3/3 Keepers up | ok | 17 |  |
| insert, 2/3 Keepers up (one lost) | ok | 17 |  |
| insert, 1/3 Keepers up (quorum lost) | error:242 | 312 | code: 242, message: Table is in readonly mode: replica_path=/clickhouse/tables/1/ch_eval_exp/kq/replicas/ch-s1r1 |
| SELECT during quorum loss | value:2 | 10 |  |
| system.replicas.is_readonly during quorum loss | value:1 | 0 |  |
| time from Keeper restart until inserts succeed | value:5683ms | 5683 |  |

## Findings

- Losing one of three Keepers: insert ok.
- Losing quorum: insert error:242 after 312 ms; SELECT value:2; tables went read-only (is_readonly=1). Inserts recovered 5683ms after the Keepers restarted.

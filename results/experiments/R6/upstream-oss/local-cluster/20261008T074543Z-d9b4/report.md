# R6: How does DDL propagate when a replica is down (ON CLUSTER vs Replicated database)?

Target `upstream-oss/local-cluster` (`26.3.33.24`), 2026-10-08T07:42:18Z.

> Local tier: behaviour (errors, partial results, catch-up) only; timings are not production evidence.

| Step | Outcome | ms | Detail |
| --- | --- | ---: | --- |
| stop node ch-s1r2 | ok | 5275 |  |
| CREATE TABLE ... ON CLUSTER with distributed_ddl_task_timeout=10 | error:159 | 10609 | code: 159, message: Distributed DDL task /clickhouse/task_queue/ddl/query-0000000194 is not finished on 1 of 2 hosts (0 of them are currently executing the task, 0 are inactive). They are going to execute the query in background. Was waitin… |
| CREATE TABLE in a Replicated database (no ON CLUSTER) | error:159 | 180839 | code: 159, message: ReplicatedDatabase DDL task /clickhouse/databases/ch_eval_exp_repl/log/query-0000000003 is not finished on 1 of 2 hosts (0 of them are currently executing the task, 0 are inactive). They are going to execute the query in… |
| restart node | value:2080ms | 2453 |  |
| time after restart until ch_eval_exp.ddl_t exists on the restarted node | value:5002ms | 5002 |  |
| time after restart until ch_eval_exp_repl.ddl_t exists on the restarted node | value:3ms | 3 |  |

## Findings

- ON CLUSTER with a replica down returned error:159 after 10609 ms (the client sees a timeout while the task stays queued for the missing host).
- In a Replicated database the same DDL returned error:159 after 180839 ms; the restarted replica replays it from the database's log.

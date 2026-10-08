# R6: How does DDL propagate when a replica is down (ON CLUSTER vs Replicated database)?

Target `altinity/local-cluster` (`26.3.33.10001.altinitystable`), 2026-10-08T07:47:08Z.

> Local tier: behaviour (errors, partial results, catch-up) only; timings are not production evidence.

| Step | Outcome | ms | Detail |
| --- | --- | ---: | --- |
| stop node ch-s1r2 | ok | 5316 |  |
| CREATE TABLE ... ON CLUSTER with distributed_ddl_task_timeout=10 | error:159 | 10587 | code: 159, message: Distributed DDL task /clickhouse/task_queue/ddl/query-0000000230 is not finished on 1 of 2 hosts (0 of them are currently executing the task, 0 are inactive). They are going to execute the query in background. Was waitin… |
| CREATE TABLE in a Replicated database (no ON CLUSTER), distributed_ddl_task_timeout=10 | error:159 | 10583 | code: 159, message: ReplicatedDatabase DDL task /clickhouse/databases/ch_eval_exp_repl/log/query-0000000003 is not finished on 1 of 2 hosts (0 of them are currently executing the task, 0 are inactive). They are going to execute the query in… |
| restart node | value:1510ms | 1769 |  |
| time after restart until ch_eval_exp.ddl_t exists on the restarted node | value:822ms | 822 |  |
| time after restart until ch_eval_exp_repl.ddl_t exists on the restarted node | value:2ms | 2 |  |

## Findings

- ON CLUSTER with a replica down returned error:159 after 10587 ms; see the steps for when the restarted replica executed the queued task.
- In a Replicated database the same DDL returned error:159 after 10583 ms; see the steps for when the restarted replica replayed it from the database log.

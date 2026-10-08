# AI Cost Management (internal workload)

Evaluation of AI Cost Management usage analytics: ClickHouse Cloud vs a sharded Altinity
`prod-us` cluster. **Internal; remove before the repository is made public.**

| File | Contents |
| --- | --- |
| [`BRIEF.md`](BRIEF.md) | The brief, with the SQL in `sql/` |
| [`REPORT.md`](REPORT.md) | The deliverable |
| `sql/schema/`, `sql/queries/` | Verbatim DDL and statements from the brief |
| `sql/gen/` | Deterministic synthetic data generator (no real data) |
| `cluster/` | Local 3 shards × 2 replicas copy of the proposed `prod-us` layout (Docker Compose) |
| `scripts/` | HTTP client, runner, experiments |
| `results/` | Outputs cited by the report |

## Reproduce

All runs are local and correctness-only. They are not platform performance results.

```bash
nohup dockerd >/tmp/dockerd.log 2>&1 &            # if the daemon is not running
cd internal/ai-cost-management

# Q1: feature acceptance per version (single node)
scripts/q1_features.sh altinity/clickhouse-server:26.3.33.10001.altinitystable clickhouse/clickhouse-server:26.8.2.7

# 3x2 cluster (keeper + 6 x altinity 26.3.33.10001)
(cd cluster && docker compose up -d)               # restart once if servers start before Keeper is the leader
H=$(cluster/ip.sh ch-s1r1)
scripts/run_sql.py $H default cluster/sql/00_provision.sql
for f in sql/schema/0*.sql; do scripts/run_sql.py $H ai_cost_management_local $f \
  --engine-replace 'ENGINE = ReplacingMergeTree|ENGINE = ReplicatedReplacingMergeTree' --continue; done
scripts/run_sql.py $H ai_cost_management cluster/sql/10_cluster_only.sql
scripts/run_sql.py $H default cluster/sql/20_users.sql
scripts/run_sql.py $H default sql/gen/u_function.sql
scripts/gen.py $H ai_cost_management 2000000 --set distributed_foreground_insert=1
scripts/gen.py $H ai_cost_management 2000000 --filter 'number % 20 = 0' --set distributed_foreground_insert=1
scripts/gen.py $H ai_cost_management 2000000 --filter 'number % 200 = 1' --tshift-ms 1000 --set distributed_foreground_insert=1
scripts/run_queries.py cluster3x2 $H --cluster prod_us

# teardown
(cd cluster && docker compose down -v)
```

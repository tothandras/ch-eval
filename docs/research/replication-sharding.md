# Replication and sharding on self-managed ClickHouse 26.3: cited facts

Input to design D13 and to feature-matrix section 9. These are **source
facts, not probe results.** Each one gets a probe or experiment in Phase 2.

## Sources

| Key | Repository | Commit |
| --- | --- | --- |
| **[CH]** | `ClickHouse/ClickHouse` tag `v26.3.33.24-lts` | `211b7286` |
| **[KB]** | `Altinity/altinityknowledgebase` | `bdc41270` |
| **[OP]** | `Altinity/clickhouse-operator` tag `release-0.27.4` | `82498653` |

- **Defaults** come from `src/Core/Settings.cpp` and `MergeTreeSettings.cpp`.
- **Altinity build check:** every `DECLARE` default line is identical between
  `v26.3.33.10001.altinitystable` and upstream `26.3.33.24`, so the defaults
  below hold for both builds.
- **Published URLs:**
  - [CH] docs: `clickhouse.com/docs/<slug>`
  - settings: `clickhouse.com/docs/operations/settings/settings#<name>`
  - [KB]: `kb.altinity.com/<path>`

## Replication

| Topic | Fact (default at 26.3) | Quote | Source |
| --- | --- | --- | --- |
| Insert ack | One replica acknowledges; the others fetch asynchronously | "By default, an INSERT query waits for confirmation of writing the data from only one replica." / "recently inserted data appears on the other replicas with some latency." | /engines/table-engines/mergetree-family/replication |
| `insert_quorum` | 0 (off); `'auto'` = majority | "`INSERT` succeeds only when ClickHouse manages to correctly write data to the `insert_quorum` of replicas during the `insert_quorum_timeout`." | settings |
| `insert_quorum_parallel` | **true** | — | settings |
| `select_sequential_consistency` | 0 | "Requires `insert_quorum_parallel` to be disabled (enabled by default)." | settings |
| `SYSTEM SYNC REPLICA` | Plain, STRICT, LIGHTWEIGHT and PULL forms; waits at most `receive_timeout` (300 s) | LIGHTWEIGHT "waits only for `GET_PART`, `ATTACH_PART`, `DROP_RANGE`, `REPLACE_RANGE` and `DROP_PART` entries" | /sql-reference/statements/system |
| `fallback_to_stale_replicas_for_distributed_queries` | **true** | "Forces a query to an out-of-date replica if updated data is not available." | settings |
| `max_replica_delay_for_distributed_queries` | **300** s | "If a replica's lag is greater than or equal to the set value, this replica is not used." | settings |
| `load_balancing` | **random** | "if the replicas have different data, you will also get different data." | settings |
| Operator profile override | `load_balancing=nearest_hostname`, `prefer_localhost_replica=0` | — | [OP] `config/chi/users.d/02-clickhouse-default-profile.xml` |
| Keeper unavailable | SELECTs keep working. INSERTs throw. Tables go read-only on session loss and recover automatically. | "ZooKeeper is not used in `SELECT` queries" / "If ClickHouse Keeper is unavailable during an `INSERT`… an exception is thrown." | /engines/table-engines/mergetree-family/replication |
| Keeper sizing | 3 nodes, dedicated hosts, fast log disk, low latency | "we don't recommend running more than three Keeper nodes (excluding observers)." / "USE DEDICATED FAST DISKS for the transaction log!" | [KB] altinity-kb-zookeeper/clickhouse-keeper, altinity-kb-proper-setup |
| `SYSTEM RESTORE REPLICA` | Rebuilds lost Keeper metadata | "Works only on readonly `ReplicatedMergeTree` tables." | /sql-reference/statements/system |
| ON CLUSTER with a host down | Eventually executed; the client times out after `distributed_ddl_task_timeout` = 180 s (`throw` mode) | "The local version of the query will eventually be executed on each host in the cluster, even if some hosts are currently not available." | /sql-reference/distributed-ddl |
| Replicated database | Replicates metadata only; became beta in 24.4. Its tier at 26.3 is unverified. | "the database is responsible only for metadata" | /engines/database-engines/replicated |
| Zero-copy | `allow_remote_fs_zero_copy_replication` = false | "Don't use this setting in production, because it is not ready." | MergeTreeSettings |

## Sharding

| Topic | Fact (default at 26.3) | Quote | Source |
| --- | --- | --- | --- |
| `internal_replication=true` | Write to one healthy replica per shard; replication does the rest. The operator sets it when `replicasCount>1`. | "the write operation selects the first healthy replica and writes data to it." | /engines/table-engines/special/distributed; [OP] `normalizer.go` |
| `distributed_foreground_insert` | **false**, so inserts are queued on local disk and sent in the background | "the data block is just written to the local file system. The data is sent to the remote servers in the background as soon as possible." | Distributed docs |
| Background-queue loss | — | "If the server ceased to exist or had a rough restart… the inserted data might be lost." | Distributed docs |
| `skip_unavailable_shards` | **false**: an unavailable shard is an error. With 1 the result is partial and silent. | "returns a result based on partial data and does not report node availability issues." | settings |
| `optimize_skip_unused_shards` | **false** | "Assumes that the data is distributed by sharding key, otherwise a query yields incorrect result." | settings |
| Rebalancing | None | "ClickHouse doesn't support automatic shard rebalancing." | /guides/sre/scaling-clusters |
| clickhouse-copier | Removed from the bundle in 24.3; no longer supported | "This tool is no longer supported" | ClickHouse/copier README |
| Part moves between shards | `part_moves_between_shards_enable` = 0, experimental | "Does not take into account sharding expressions." | MergeTreeSettings |
| Tenant sharding | Shard by user/tenant to co-locate a tenant's data | "the data of a single user will reside on a single shard, which simplifies running `IN` and `JOIN` by users" | Distributed docs |
| When to shard (ClickHouse) | Scale vertically first | "we generally recommend using the largest server available to prevent having to re-shard" | /guides/sizing-and-hardware-recommendations |
| When to shard (Altinity) | Shard early | "Starting with a sharded cluster from the beginning makes it easier in the future to scale the cluster up." | [KB] cluster-production-configuration-guide |
| Limits of thumb (Altinity) | — | "Less than 10TB of compressed data per server." / replicas: "2 is minimum for HA. 3 is a 'golden standard'." | [KB] how-much-is-too-much |
| Parallel replicas | `enable_parallel_replicas`=0; production tier from 25.12 (inferred: BETA flag removed); disabled with FINAL | "Parallel replicas are disabled with FINAL" | /deployment-guides/parallel-replicas |
| Operator scale-up | Schema is created on new replicas and shards automatically | "After cluster is scaled up and new ClickHouse instances are ready, `clickhouse-operator` automatically creates schema." | [OP] `docs/schema_migration.md` |
| Operator cluster names | One per `clusters[].name` (default "cluster"), plus `all-replicated`, `all-sharded`, `all-clusters` | — | [OP] `pkg/model/common/creator/cluster.go` |

## Implications to test (OpenMeter)

1. **Reads can go backwards by default.** Distributed reads may use replicas up
   to 300 s stale, and even staler ones if every replica is stale
   (`fallback_to_stale_replicas_for_distributed_queries=1`). With
   `load_balancing=random`, successive reads can decrease. The operator's
   profile changes `load_balancing` but not the stale fallback. → D7 and R1
   measure it.
2. **`select_sequential_consistency` alone does nothing on
   ReplicatedMergeTree.** It needs both `insert_quorum` and
   `insert_quorum_parallel=0`, and the latter's default is the opposite. → R2
   measures the write cost.
3. **The default Distributed insert is fire-and-forget** to local disk, and the
   queue can be lost on a hard restart. That makes the visibility delay and
   loss risk real for a billing pipeline. → S4.
4. **One misconfiguration away from partial billable totals.**
   `skip_unavailable_shards=1` returns them silently. The default is safe (it
   errors). → S3 records both behaviours.
5. **Shard pruning is not default.** It needs `optimize_skip_unused_shards=1`,
   and gives wrong results if the data is not actually distributed by the key.
   → S2 and the S1 parity gate.
6. **Upstream and Altinity disagree on when to shard.** ClickHouse says "largest
   server first"; Altinity's KB says "start sharded". The D15 rightsizing sweep
   supplies the evidence for OpenMeter's volumes.

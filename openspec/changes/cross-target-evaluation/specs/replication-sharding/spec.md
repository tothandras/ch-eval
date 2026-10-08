## ADDED Requirements

### Requirement: Self-managed topology variants
Self-managed targets SHALL support the topology variants `1x2`, `1x4`, and `2x2` (shards × replicas) from the same target manifest family. The variants SHALL differ only in `shardsCount` and `replicasCount` (and pod count), keeping image, resources, and settings identical. Sharded variants SHALL use `Distributed` tables over `ReplicatedMergeTree` local tables with `internal_replication=true`.

#### Scenario: Same manifest, different layout
- **WHEN** the `2x2` and `1x2` variants of `altinity` are deployed
- **THEN** their metadata shows identical versions, resources, and changed settings, and differs only in the cluster layout

### Requirement: Sharded-table migration helper
The migration templating SHALL provide a `shardedTable` helper. On sharded topologies it SHALL render a `<name>_local` replicated table plus a `Distributed` table `<name>` with a declared sharding key. On unsharded dialects it SHALL render the plain table `<name>`. Scenario query files SHALL NOT change between sharded and unsharded runs.

#### Scenario: Queries unchanged under sharding
- **WHEN** the `proposal` scenario runs on `2x2` and on `1x2`
- **THEN** the same query files execute, and `bench compare` reports MATCH for every query's digest

### Requirement: Replication and sharding experiments
The harness SHALL measure, on self-managed targets:
- replica staleness and the default stale-replica fallback of distributed reads;
- quorum-insert overhead;
- the effect of Keeper quorum loss (time until tables go read-only, the error surfaced to writers, and recovery time);
- replica rebuild time after pod and volume loss;
- per-replica storage amplification;
- DDL propagation with a replica down;
- shard pruning by sharding key;
- behaviour with an unavailable shard under default settings and under `skip_unavailable_shards=1`;
- time-to-queryable for Distributed background, Distributed foreground, and direct-to-shard inserts;
- the cost of adding a shard and rebalancing data.

Any configuration that can return a partial or decreasing billable total without an error SHALL be reported as a design callout.

#### Scenario: Partial result without error
- **WHEN** a shard is stopped and a sum query runs with `skip_unavailable_shards=1`
- **THEN** the experiment records whether the query succeeded, and the deficit against the known total, and flags a design callout if it succeeded with a deficit

## ADDED Requirements

### Requirement: Cross-replica visibility measurement
The `replica-consistency` experiment SHALL insert OpenMeter-shaped batches with a monotonically increasing sequence at a configured rate and, for each acknowledged batch, SHALL measure per replica the delay from insert acknowledgement until the batch is visible, using per-replica reads (`clusterAllReplicas` grouped by `hostName()`).

#### Scenario: Lag distribution recorded
- **WHEN** the experiment runs against a multi-replica target
- **THEN** the result records the p50, p99, and max ack-to-visible lag per replica and overall

### Requirement: Client-observed staleness and monotonicity
Concurrent readers SHALL repeatedly read the namespace's cumulative total through the target's normal load-balanced endpoint and SHALL record stale reads (an observed total below the total acknowledged before the read started), monotonicity violations (a total lower than the same reader's previous observation), and the deficit size in rows and seconds.

#### Scenario: Decreasing total detected
- **WHEN** a reader observes 100 and then 90 for a strictly increasing stream
- **THEN** a monotonicity violation is counted with deficit 10

### Requirement: Consistency policies compared
The experiment SHALL run each applicable policy: ReplicatedMergeTree default, ReplicatedMergeTree with `insert_quorum` equal to the replica count plus `insert_quorum_parallel=0` plus `select_sequential_consistency=1`, ReplicatedMergeTree with `SYSTEM SYNC REPLICA ... LIGHTWEIGHT` before reads, SharedMergeTree default, and SharedMergeTree with `select_sequential_consistency=1`. Each policy runs with synchronous and asynchronous inserts, under steady ingest, and during a replica restart. The insert and read latency overhead of each policy SHALL be reported next to its staleness results.

#### Scenario: Policy overhead reported
- **WHEN** a policy run completes
- **THEN** the result contains its stale-read rate, violation count, lag percentiles, and insert/read latency p50/p99 next to the default policy's figures

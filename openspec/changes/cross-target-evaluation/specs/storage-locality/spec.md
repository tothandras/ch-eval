## ADDED Requirements

### Requirement: Locality configurations
The evaluation SHALL compare three storage-locality configurations:
- `separated`: `cloud`, SharedMergeTree on object storage;
- `colocated`: self-managed ReplicatedMergeTree on gp3 EBS;
- `self-separated`: self-managed ReplicatedMergeTree on an S3 disk with a local filesystem cache.

They SHALL use the same workload, data, and shape. The result metadata SHALL record the storage policy and cache configuration where readable.

#### Scenario: Locality recorded
- **WHEN** a perf run completes on any of the three configurations
- **THEN** `metadata.json` names the configuration and records the disks, storage policies, and cache settings visible on the target

### Requirement: Cold, warm, and working-set measurements
For each query class, the harness SHALL measure:
- latency and CPU with caches warm and with caches dropped by the strongest mechanism the target allows, recording which mechanism was used;
- bytes served from cache versus from remote or local storage;
- p95 latency as a function of the query window (1h/24h/7d/30d);
- on `self-separated` only, p95 latency as a function of the filesystem cache size relative to the table's compressed size.

#### Scenario: Asymmetric cache control recorded
- **WHEN** the OS page cache cannot be dropped on `cloud`
- **THEN** the cold result records the mechanism used (`filesystem-cache-drop`) and the report labels the comparison as asymmetric

### Requirement: Write path and warm-up
The harness SHALL measure, per locality configuration:
- insert latency, time-to-queryable, merge throughput, and merge CPU;
- object-store requests per 1M events;
- the first-query latency curve after a replica is added, restarted, or resumed from idle, until the curve reaches its steady-state p95.

#### Scenario: Warm-up curve
- **WHEN** a new replica starts serving
- **THEN** the result contains latency per query over time until three consecutive windows are within 10% of the steady-state p95

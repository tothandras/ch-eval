## MODIFIED Requirements

### Requirement: Multi-replica log collection
Log collectors SHALL read `system.query_log` and `system.part_log` across all replicas via `clusterAllReplicas(<cluster>, ...)`, where `<cluster>` comes from the target manifest (`default` on ClickHouse Cloud, the CHI cluster name on operator-managed targets). They SHALL flush logs on all replicas before collecting. Available CPU SHALL be computed as the detected replica count × per-replica vCPUs × the measure window. When the target uses S3-backed storage, collectors SHALL also aggregate S3 request counters (`S3GetObject`, `S3PutObject`, `S3UploadPart`, `S3ListObjects`, ...) from query and part logs and from `system.events` deltas.

#### Scenario: merge scheduled on a non-connected replica
- **WHEN** a merge of the scenario table executes on a replica other than the one the harness is connected to
- **THEN** its CPU is included in the merge attribution

#### Scenario: cluster-wide flush unavailable
- **WHEN** `SYSTEM FLUSH LOGS ON CLUSTER` is not permitted on the target
- **THEN** the collector falls back to a local flush plus settle delay and the result records `log_flush: "local-only"`

#### Scenario: operator cluster name
- **WHEN** the target manifest declares cluster `ch-eval`
- **THEN** collectors read `clusterAllReplicas(ch-eval, system.query_log)` rather than `default`

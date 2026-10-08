## ADDED Requirements

### Requirement: Target manifests
Each target SHALL be described by a manifest at `targets/<name>/target.json` that declares its platform (`cloud`, `altinity`, `upstream-oss`, or a named control such as `upstream-cloudver`), server image or service, topology tier (`local-single`, `local-cluster`, `kind`, `aws`, `cloud`), DDL dialect (`single`, `replicated`, `cloud`), the cluster name used for `ON CLUSTER` and `clusterAllReplicas`, the storage mode (`local`, `ebs`, `s3`, `shared`), and the intended shape (replicas, vCPU and GiB per replica). Manifests SHALL NOT contain credentials or connection strings. Connection details SHALL come only from environment variables (`CLICKHOUSE_DSN_<TARGET>` or `CLICKHOUSE_DSN`), typically set in a git-ignored `.envrc`.

#### Scenario: Manifest without secrets
- **WHEN** a target manifest is committed
- **THEN** it contains no password, token, or DSN, and the harness resolves the connection from the environment at run time

### Requirement: Single up/down entry point per target tier
Every target tier SHALL be provisioned and torn down through one `up` and one `down` entry point (`targets/<name>/<tier>/up` and `.../down`). `down` SHALL be idempotent and SHALL verify deletion by listing the provider's resources filtered by the run-id tag, exiting non-zero if anything remains.

#### Scenario: Verified teardown
- **WHEN** `down` completes for a paid tier
- **THEN** a provider listing filtered by the run id returns no resources, and otherwise `down` exits non-zero naming the survivors

### Requirement: Paid-resource guardrails
Provisioning scripts for paid tiers SHALL tag every resource with the run id, SHALL print an estimated cost for the requested shape and duration before creating anything, and SHALL refuse to run without an explicit budget acknowledgement (`BUDGET_ACK=<run-id>`). Paid resources SHALL NOT be created without the repository owner's approval of the estimate.

#### Scenario: Missing budget acknowledgement
- **WHEN** `up` for the `aws` or `cloud` tier is invoked without `BUDGET_ACK` matching the run id
- **THEN** it prints the estimate and exits without creating resources

### Requirement: Live run metadata
Every benchmark, cogs, probe, migration, and experiment run SHALL capture live target metadata into `metadata.json`: `version()`, build options, cluster topology (`system.clusters`, `system.macros`), disks and storage policies, changed session, MergeTree, and server settings, CPU and memory limits, the harness commit, the client host, and the manifest. When the live target contradicts the manifest (version, replica count, dialect), the run SHALL fail unless `--allow-drift` is given, in which case the drift SHALL be recorded.

#### Scenario: Version drift detected
- **WHEN** the manifest pins `26.3.33.10001.altinitystable` and the server reports another version
- **THEN** the run aborts before measuring, unless `--allow-drift` was passed

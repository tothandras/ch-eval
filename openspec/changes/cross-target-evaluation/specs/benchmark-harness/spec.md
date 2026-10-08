## MODIFIED Requirements

### Requirement: Scenario discovery

The benchmark harness SHALL discover scenarios by scanning `scenarios/<workload>/<variant>/` at runtime. The user MAY restrict the run with `--workload` and `--scenario` flags (repeatable). Scenarios that do not satisfy the `scenario-format` contract SHALL be skipped with a logged warning, never aborting the run.

#### Scenario: Discover all scenarios
- **WHEN** the harness is invoked with no workload or scenario filter
- **THEN** it runs every scenario of every workload that satisfies the `scenario-format` contract

#### Scenario: Filter to one scenario
- **WHEN** the harness is invoked with `--workload openmeter --scenario baseline-openmeter`
- **THEN** it runs only that scenario and skips the others without error

### Requirement: Result persistence

Results SHALL be written to `results/<workload>/<scenario>/<target>/<run-id>/run.json` with a sibling `report.md` and `metadata.json`. Each run file SHALL include the scenario name, workload, target name and tier, the cluster fingerprint (ClickHouse version, cluster topology via `system.clusters`, and an `is_single_node` flag determined by probing actual shard and replica counts), the harness git commit (with a `-dirty` suffix when the working tree has uncommitted changes), the concurrency levels used, the sweep id used for CPU correlation, and the run start and end timestamps. Per-query CPU (`cpu_p50_us`, `cpu_p95_us`) and memory (`mem_avg_bytes`) SHALL be persisted alongside the latency and throughput fields, or set to null when `query_log` was unavailable. The `results/` tree SHALL be tracked in version control.

#### Scenario: Result file is self-describing
- **WHEN** a run directory is read in isolation
- **THEN** a reviewer can determine the workload, scenario, target, tier, ClickHouse version, harness commit (clean or dirty), concurrency, and timing without any other file

#### Scenario: CPU recorded or explicitly null
- **WHEN** a result file is read for a run where `query_log` was disabled on the target
- **THEN** every query's `cpu_p50_us`, `cpu_p95_us`, and `mem_avg_bytes` are `null` (not absent, not zero)

#### Scenario: Run produces JSON, markdown, and metadata
- **WHEN** a scenario run completes
- **THEN** `run.json`, `report.md`, and `metadata.json` are written under the run directory, and none is excluded by `.gitignore`

## ADDED Requirements

### Requirement: Results per workload, scenario, and target
Every run SHALL write to `results/<workload>/<scenario>/<target>/<run-id>/` containing `run.json` (the existing run schema plus a `target` block), `report.md`, and `metadata.json`. The run id SHALL be `<UTC timestamp>-<short random suffix>`, and that same id SHALL be used to tag any cloud resources the run creates.

#### Scenario: Self-describing run directory
- **WHEN** a reviewer opens a run directory
- **THEN** they can tell the workload, scenario, target, tier, server version, shape, seed, settings, and harness commit without any other file

### Requirement: Cross-target compare with a parity gate
`bench compare --workload W --scenario S --targets a,b[,c]` SHALL compare the latest valid run of S per target and print per-query latency and CPU deltas. It SHALL diff value digests across targets over an identical seeded window, and exit non-zero naming the queries whose digests differ. A digest mismatch SHALL block any performance comparison of that workload and scenario until investigated.

#### Scenario: Digest mismatch blocks the run
- **WHEN** one query's digest differs between `cloud` and `altinity`
- **THEN** compare exits non-zero, names the query, and the report generator refuses to state a performance verdict for that scenario

### Requirement: Local tiers never back performance conclusions
Runs from `local-single`, `local-cluster`, or `kind` tiers SHALL be usable only for parity, correctness, and smoke checks. Comparisons that mix tiers, or that use only local tiers, SHALL be labelled "smoke, not performance".

#### Scenario: Local-only compare
- **WHEN** compare runs on two local-tier runs
- **THEN** the output header states the comparison is a smoke check, not a performance result

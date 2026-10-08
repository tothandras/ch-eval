## ADDED Requirements

### Requirement: Compute scaling experiments
The harness SHALL measure compute scaling on each platform:

- **Cloud:**
  - vertical autoscaling response to a load step (time to scale up and down, errors and latency during the change, billed compute units);
  - horizontal scale-out time-to-serve;
  - resume-from-idle latency.
- **Self-managed:**
  - vertical scaling through a pod-resource change (rolling-restart duration, failed queries and inserts);
  - adding a replica (schema creation, time to fetch the full dataset, impact on existing replicas);
  - adding a shard.

Scaling behaviour that cannot be triggered on the tier in use SHALL be reported as "assessed, not measured" with citations.

#### Scenario: Load-step on Cloud
- **WHEN** load steps from idle to the declared heavy level
- **THEN** the result contains a time series of per-replica CPU and memory limits, query p95, and error counts, plus the billed compute units for the window

### Requirement: Storage scaling and headroom
The harness SHALL measure:
- self-managed EBS online volume expansion (duration, disruption);
- peak disk usage during merges and during each online-change migration, as a fraction of steady-state table size. This headroom factor SHALL be used when sizing provisioned storage in cost models.

Provider limits (modification cooldowns, inability to shrink) SHALL be cited.

#### Scenario: Migration headroom
- **WHEN** the `data String` → `JSON` migration runs on a populated table
- **THEN** the result records peak bytes on disk relative to the pre-migration size per replica

### Requirement: Rightsizing and load-profile cost
For each workload and target, the harness SHALL determine the smallest replica shape (from a declared sweep) that meets the declared SLO at the declared target load. The report SHALL compare costs at each target's smallest compliant shape as well as at identical shapes. A load-profile cost model SHALL compute monthly cost for a parameterized diurnal or weekly curve and report the crossover points (peak-to-average ratio, utilization, ops-hours per month) at which the cheaper target changes. Cloud's curve SHALL use the measured autoscaling and idle behaviour; self-managed SHALL be provisioned for peak.

#### Scenario: Crossover reported
- **WHEN** the cost model runs with default parameters
- **THEN** the report lists, per workload, the parameter values at which the cheaper target flips, and the verdict's flip conditions reference them

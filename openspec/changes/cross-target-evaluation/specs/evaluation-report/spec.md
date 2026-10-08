## ADDED Requirements

### Requirement: Generated report
`REPORT.md` SHALL be produced by `bench report` from `results/` and the feature-matrix data. It SHALL NOT contain hand-typed numbers. It SHALL contain, in order: OpenMeter design-decision callouts (any probed or measured difference that changes an OpenMeter design decision), an executive summary, a verdict per workload, feature-matrix highlights, a cost comparison with utilization crossover points, caveats, and open questions.

#### Scenario: Number provenance
- **WHEN** a number appears in `REPORT.md`
- **THEN** it is traceable to a result file path rendered next to the table it appears in

### Requirement: Verdicts with flip conditions
Each workload verdict SHALL state a recommendation (`cloud`, `altinity`, or "no material difference") using decision rules documented in `docs/methodology.md`, and SHALL list the computed conditions under which it would flip (for example a utilization threshold, ops-hours per month, or a feature becoming available). Verdict inputs SHALL be limited to `aws` and `cloud` tier runs that passed the parity gate.

#### Scenario: Insufficient evidence
- **WHEN** a workload lacks parity-clean paid-tier runs on both cloud and altinity
- **THEN** its verdict reads "insufficient evidence" and names the missing runs

### Requirement: Measured versus assessed labelling
Every statement in the report SHALL be marked as measured (with a result path) or as "assessed, not measured" (with a citation).

#### Scenario: Unmeasurable aspect
- **WHEN** an aspect such as compute-compute separation cannot be measured on a target
- **THEN** the report states "assessed, not measured" and cites the source

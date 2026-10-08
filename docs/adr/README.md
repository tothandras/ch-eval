# Architecture decision records

| ADR | Title | Status |
| --- | --- | --- |
| [0001](0001-fork-existing-harness.md) | Fork the existing harness; extend, don't rewrite | Accepted |
| [0002](0002-target-dimension.md) | Target platform as a first-class dimension | Proposed (Phase 1) |
| [0003](0003-migration-tool.md) | ClickHouse migrations: golang-migrate layout, custom dialect-aware runner | Proposed (Phase 1) |
| [0004](0004-otel-scenarios.md) | OTel workloads: verbatim upstream schemas plus one evidence-justified tuned variant | Proposed (Phase 1) |
| [0005](0005-topology-and-shapes.md) | Local tiers for correctness only; matched paid shapes for performance and cost | Proposed (Phase 1) |

New ADRs are numbered sequentially. An ADR is required before changing the
runner, digest, or COGS attribution semantics inherited from the source
harness.

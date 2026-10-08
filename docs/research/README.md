# Research inputs

Material that shaped the plan but is **not** measured evidence. Every claim
here is a hypothesis until a probe or workload in this repo confirms or
refutes it; the feature matrix and `REPORT.md` cite results, not these notes.

| File | Origin | Status |
| --- | --- | --- |
| [`sharedmergetree-usage-collection.md`](sharedmergetree-usage-collection.md) | Provided by the repo owner, 2026-10-08 | Hypotheses → tested by the `replica-consistency` workload (see `openspec/changes/cross-target-evaluation`) |
| [`openmeter-clickhouse-ddl.md`](openmeter-clickhouse-ddl.md) | Source reading of `openmeterio/openmeter@5a0e27e6` | Input to ADR-0003 (migrations) |
| [`otel-schemas.md`](otel-schemas.md) | Source reading of collector-contrib v0.162.0 and ClickStack (hyperdx) | Input to the OTel scenario variants |
| [`versions.md`](versions.md) | Docker Hub, GitHub tags/releases, ClickHouse docs source | Pins the versions under test |
| [`replication-sharding.md`](replication-sharding.md) | Source reading: ClickHouse v26.3.33.24 docs and settings, Altinity KB, Altinity operator 0.27.4 | Input to design D13 and feature-matrix §9 |
| [`compute-storage-separation.md`](compute-storage-separation.md) | Source reading: ClickHouse docs and settings (master, v26.3.44.2) | Input to design D14 and feature-matrix §10 |
| [`scaling-rightsizing.md`](scaling-rightsizing.md) | Source reading: ClickHouse docs and Cloud OpenAPI, operator 0.27.4, botocore, AWS Price List API | Input to design D15, feature-matrix §11 and cost estimates |

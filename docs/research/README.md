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

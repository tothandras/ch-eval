## ADDED Requirements

### Requirement: OTel scenario variants
The `otel-traces`, `otel-metrics`, and `otel-logs` workloads SHALL each provide an `exporter-default` variant (verbatim DDL from the collector-contrib ClickHouse exporter at a pinned release), a `clickstack` variant (verbatim DDL from ClickStack at a pinned commit), and one `tuned` variant whose every deviation from its parent schema is justified by a recorded A/B result. Only database and engine placeholders MAY differ from the cited source DDL.

#### Scenario: Verbatim provenance
- **WHEN** a reviewer opens a variant's init migration
- **THEN** its header cites the source file and commit, and a diff against the source shows only placeholder substitutions

### Requirement: Seeded OTel generators
Generators for traces, metrics, and logs SHALL be pure functions of (seed, index) so that any target receives byte-identical data. They SHALL expose cardinality knobs: services, operations per service, instances per service, high-cardinality attributes, trace depth and fanout, metric series count (including a high-cardinality profile of at least 100k series), and log template and severity mix. Log records SHALL be able to correlate with generated traces.

#### Scenario: Deterministic output
- **WHEN** a generator runs twice with the same seed and range
- **THEN** both runs produce identical rows

### Requirement: Real ingest path through the collector
Each OTel workload SHALL be ingested at least once per target through an OpenTelemetry Collector (contrib, pinned version) with the ClickHouse exporter configured with `create_schema: false`, batching, and async inserts. A small dataset ingested through the collector and through direct insert SHALL produce identical digests for every query.

#### Scenario: Path parity
- **WHEN** the same generated dataset is loaded via OTLP→collector and via direct insert into two databases
- **THEN** `bench compare` reports MATCH for every query

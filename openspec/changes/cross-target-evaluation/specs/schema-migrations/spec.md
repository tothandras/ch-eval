## ADDED Requirements

### Requirement: One templated source of truth
Schema SHALL be defined once per workload and scenario as ordered migration files named `YYYYMMDDHHMMSS_<name>.up.sql[.tmpl]` with optional `.down.sql[.tmpl]`, rendered per target dialect with a closed set of template helpers (`onCluster`, `engine`, `db`, `dialect`). Dialect differences SHALL be expressed only through these helpers, never by forking files per target. Rendering SHALL produce plain SQL that golang-migrate can apply.

#### Scenario: Same file, three dialects
- **WHEN** `bench migrate render` runs one migration for the `single`, `replicated`, and `cloud` dialects
- **THEN** the output uses `MergeTree`, `ON CLUSTER` + `ReplicatedMergeTree(<keeper path>, '{replica}')`, and `SharedMergeTree` respectively, from the same source file

### Requirement: Recorded, ordered, idempotent application
The migration runner SHALL apply pending migrations in version order, recording each in a `schema_migrations_ch` table (version, name, checksum of the rendered SQL, dirty flag, applied_at, duration_ms, target) whose engine follows the dialect so that state replicates with the schema. Re-running an applied migration SHALL be a no-op. A checksum mismatch for an applied version SHALL be an error. A failed migration SHALL leave the dirty flag set and block further migrations until resolved. Up migrations SHALL use `IF [NOT] EXISTS` forms, enforced by a lint.

#### Scenario: Re-apply is a no-op
- **WHEN** `bench migrate up` runs twice against the same target
- **THEN** the second run applies nothing and exits zero

### Requirement: Migration tests on every target
The test suite SHALL apply up, down (where a down file exists), and re-apply for every workload's migrations on every dialect, and CI SHALL run it against `altinity` and `upstream-oss` local tiers on every pull request.

#### Scenario: Round trip
- **WHEN** the migration test runs on a target
- **THEN** up → down → up → up succeeds and the final schema equals the schema after the first up

### Requirement: Measured online changes
Each workload SHALL include at least one realistic online change applied to a populated table: `data String` → `JSON` on `om_events`, adding a skip index with `MATERIALIZE INDEX`, adding a materialized column with `MATERIALIZE COLUMN`, or adding a rollup MV with backfill. The runner SHALL record, per statement and target, the wall duration until all mutations finish, CPU and bytes read and written from `query_log`/`part_log`, and the effect on concurrent query latency when a load is running.

#### Scenario: Online change cost recorded
- **WHEN** an online-change migration runs with `--measure`
- **THEN** the result records duration, CPU seconds, and read/written bytes per statement for that target

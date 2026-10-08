## MODIFIED Requirements

### Requirement: Scenario directory layout

A scenario SHALL be a directory `scenarios/<workload>/<variant>/` whose `<workload>` and `<variant>` names are kebab-case identifiers. Every scenario directory SHALL contain an `init/` migration set (see the `schema-migrations` capability), a `scenario.json` naming the workload, database, logical tables (the first is the seeded table) and registered generator, and a `queries/` subdirectory. It MAY contain `queries/mix.json` for cogs replay. A single-file `init.sql` is no longer accepted: every shipped scenario is ported, and a second code path would let the migration and legacy init drift apart. Data is loaded by a registered generator or by an external mechanism.

#### Scenario: Discoverable scenario
- **WHEN** the benchmark driver scans `scenarios/`
- **THEN** every `scenarios/<workload>/<variant>/` directory containing `init/` is treated as a runnable scenario of that workload

#### Scenario: Scenario missing init
- **WHEN** a directory under `scenarios/<workload>/` contains no `init/`
- **THEN** the driver SHALL skip it and log a warning

### Requirement: init.sql contract

A scenario's init SHALL be an ordered migration set that creates every table, materialized view, projection, and dictionary the scenario requires, rendered for the target's dialect and database. It MUST be idempotent: re-applying it leaves existing objects unchanged and succeeds.

#### Scenario: Re-running init
- **WHEN** the driver applies a scenario's init against a database that already has the scenario's objects
- **THEN** every statement SHALL succeed without error and the existing objects SHALL be unchanged

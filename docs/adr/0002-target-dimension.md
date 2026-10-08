# ADR-0002: Target platform as a first-class dimension

- Status: Proposed (Phase 1)
- Date: 2026-10-08

## Context

The source harness treats "the ClickHouse at `$CLICKHOUSE_DSN`" as implicit.
Results record only `version()` and cluster names. Comparing Cloud and
self-managed fairly needs more on every run:

- the exact build and topology;
- storage mode;
- changed settings;
- shape.

It also needs a guarantee that local runs never leak into performance claims.

## Decision

- **Manifest.** `targets/<name>/target.json` declares platform, image or
  service, tier, dialect, cluster name, storage mode, shape and location. It
  never contains credentials. The DSN comes from `CLICKHOUSE_DSN_<TARGET>` or
  `CLICKHOUSE_DSN`.
- **Live metadata, captured every run.** `bench/target` writes
  `metadata.json` with:
  - `version()` and `system.build_options`;
  - `system.clusters` and `system.macros`;
  - disks and policies;
  - changed `system.settings`, `system.merge_tree_settings` and
    `system.server_settings`;
  - the cgroup CPU and memory limits;
  - the client host and harness commit.
- **Drift check.** Manifest-vs-live drift fails the run unless
  `--allow-drift` is given.
- **Results layout.** Results go to
  `results/<workload>/<scenario>/<target>/<run-id>/`. The run ID is
  `<UTC timestamp>-<short ULID>`, and it is the same tag applied to any paid
  resource.
- **Tier gate.** Tiers `local-single`, `local-cluster` and `kind` are
  correctness-only. `bench report` ignores them for performance and cost.

## Consequences

- The `runner.Run` JSON schema is unchanged. A `target` block is added
  alongside it, so older tooling still reads `run.json`.
- The tier gate is enforced in code, not left to reviewer discipline.

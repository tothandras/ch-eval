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

## Implementation notes (Phase 2)

These record where the implementation differs from the decision text above.

- **Manifests.** They are `targets/<platform>/<tier>.json`, one per tier,
  named `<platform>/<tier>`. Each one names its DSN variable in `dsn_env`
  (for example `CLICKHOUSE_DSN_ALTINITY_LOCAL_CLUSTER`). DSN resolution falls
  back to `CLICKHOUSE_DSN`. Local tiers then fall back to the passwordless
  `default` user on the tier's first native port. No connection string is
  stored in the repository.
- **Run IDs.** The format is `<UTC timestamp>-<4 hex>`.
- **Results path.** The target segment is two directories: `.../<platform>/<tier>/<run-id>/`.
- **`run.json`.** It embeds the forked `runner.Run` fields flat, plus:
  - `run_id`, `workload` and `target`;
  - `seed`: generator, rows, RNG seed, namespaces, `time_end`, table and
    insert settings;
  - `migrations`: the steps applied, with per-statement timings;
  - `smoke_only`.
- **Drift.** Drift fails `bench` and `bench cogs` unless `--allow-drift` is
  given. It covers:
  - version;
  - mixed replica versions;
  - reachable replica count against shards × replicas;
  - presence of the manifest's storage policy.
- **Cluster-aware CPU probe.** On a multi-node target, per-query CPU is summed
  per `initial_query_id` over `clusterAllReplicas(<cluster>, system.query_log)`.
  Remote shards log their own threads, so a single-node read would undercount
  distributed queries.
- **COGS accounting** reads the manifest's cluster (`ch_eval` locally,
  `default` on Cloud) instead of always using `default`.

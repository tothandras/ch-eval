# Brief: ClickHouse Cloud versus sharded Altinity for AI Cost Management usage analytics

> **Internal.** Provided by the repository owner on 2026-10-08 as a new requirement for this
> evaluation.
>
> **How this copy differs from the original:**
> - The prose is the original text except for one change, noted at section 11. Every SQL statement
>   is stored once, verbatim, under [`sql/`](sql/), and is referenced below by file name.
> - Section 11's first out-of-scope item (a network-access setting) is omitted here on purpose. It is
>   out of scope and must not be re-raised.
>
> The deliverable is [`REPORT.md`](REPORT.md).

## 1. Task

AI Cost Management stores every metered AI Gateway call as one row in a multi-tenant ClickHouse table. It serves console dashboards, budget balances, budget alerts, and file exports from that table. Production US launched on 2026-10-07 on a dedicated ClickHouse Cloud service and has served almost no traffic so far. Within weeks, the team plans to move it to `prod-us`, a self-managed Altinity cluster of 3 shards × 2 replicas that other teams' production workloads share.

Evaluate three things:

1. **Platform comparison per use case.** For each use case in section 9, compare ClickHouse Cloud (one logical shard, SharedMergeTree) with Altinity `prod-us` (ReplicatedMergeTree tables on three shards behind `Distributed` tables). Cover latency, resource use, correctness risk, operational burden, and failure modes. Say where either platform is clearly better and where the difference does not matter.
2. **Schema and query performance.** Review the table design, its secondary structures, and the SQL the service actually sends. Identify the queries that will scale worst as event volume and tenant count grow, and say what to change.
3. **Sharded fitness.** Judge whether the schema, the queries, and the proposed cluster design (D-06 to D-17 in section 5) behave correctly and efficiently on `prod-us`. Cover deduplication under `FINAL`, aggregation and pagination over `Distributed` tables, row policies on remote shards, read-after-write for the notifier, migrations and mutations across replicas, and the data move.

Decisions D-01 to D-05 are taken: production ends up on `prod-us`, sharded. Do not argue for staying on Cloud in general. Do say, for a specific use case, when a measurement or documented ClickHouse behavior argues against the move or for a different way of doing it. The proposed decisions D-06 to D-17 are open, and you may challenge any of them with evidence.

## 2. How to work

- You cannot run anything against the team's clusters. When a conclusion depends on a measurement, write the exact read-only statement the team should run, such as an `EXPLAIN indexes = 1`, a `system.query_log` query, or a `clusterAllReplicas` catalog check. Say which result would change your recommendation. If you have a local ClickHouse or chDB, you may run the statements in sections 8 and 9 against it; say which version you used.
- Four ClickHouse versions are in play (section 4), and behavior changes between releases. Tie every version-sensitive claim to a version, and cite a changelog entry, documentation page, or source file. When you cannot confirm a behavior for 26.3, say so.
- Label every finding **verified** (you ran it, or a primary source states it), **inferred** (reasoned from documented behavior), or **assumed**.
- Respect the constraints in section 6. If a recommendation breaks one, for example by adding a second migration set for Altinity, say so and explain why it is worth it.
- Prefer the simplest change that meets the requirement. The team would rather delete or consolidate a structure than add a layer; removing a projection that no query uses counts as a good finding.
- Do not re-raise the items in section 11.

## 3. The service

### Data flow

1. A Kong AI Gateway data plane reports each AI call to Konnect analytics. A Flink forwarder, gated per organization, publishes the call as a CloudEvent to the Kafka topic `konnect.ai-cost-management.usage-events`. An HTTP ingestion endpoint also publishes to that topic; it takes 1 to 100 events per request with `time` between 72 hours ago and 5 minutes from now.
2. The **worker** (consumer group `ai-cost-management-usage-events`) inserts each Kafka partition batch into `usage_events` and commits offsets only after the insert succeeds.
3. A materialized view keeps `usage_event_field_values`, a lookup of observed dimension values that powers the console's pickers.
4. The **API** serves the console: aggregates, forecasts, budget balances, the event list, and field values.
5. The **notifier** (consumer group `ai-cost-management-notifier`) reads the same topic, waits until the worker has stored a batch, and then evaluates budget alert thresholds.
6. The **exporter** streams export chunks (CSV or Parquet) to S3.

### ClickHouse users

| Process | User | Access |
| --- | --- | --- |
| Worker | writer | `INSERT` on `usage_events` only, no `SELECT` |
| API, notifier | reader | Reader role; queries also set `max_execution_time = 20` |
| Exporter | reader | Reader role; queries set the export limits in section 9.8 |
| Materialized view | `ai_cost_management_field_values_definer` | `HOST NONE`; `SELECT` on events, `INSERT` on the lookup table |
| Migration job | stage 1: the Cloud service's admin; `prod-us`: a planned least-privilege migrator | DDL, access objects, field-value backfill |

The API client waits 10 seconds longer than `max_execution_time`, so the server's timeout error ends a slow query before the client gives up on it.

### Semantics the SQL depends on

- **One table for all tenants.** A tenant is a Konnect organization. Every reader query sends the tenant twice: as a typed `{tenant: UUID}` predicate, and as the query-level custom setting `SQL_tenant_id`, which a restrictive row policy reads (section 8.3).
- **Identity and deduplication.** An event's identity is `(tenant_id, source, id)`. Events are immutable; a producer retry with a changed `time` stores a second row. Kafka delivery is at-least-once and a redelivered batch is simply inserted again. Duplicates collapse because the `ReplacingMergeTree` sorting key ends in the identity, and every authoritative read uses `FINAL`. Monthly partitions on `time` are meant to keep `FINAL` partition-local (question 3 asks you to verify that).
- **Absence is a sentinel, not `NULL`.** It is `''` for strings and missing map keys, the nil UUID for identifiers, and `[]` for arrays. `empty()` and `notEmpty()` test absence, and the API renders it as `null`.
- **Exact numbers.** Costs are stored as `Decimal(18, 9)` and summed as `Decimal256(9)`. Token quantities are stored as `UInt64` and summed as `UInt128`. All values leave the API as decimal strings. A report, a budget balance, and an alert for the same scope and range must show the same number, so they share the same SQL builders.
- **Subjects and teams.** `subject` is a typed reference: `principal/{uuid}`, `consumer/{uuid}`, or `oidc/{sub}`, split at the first `/`. `teams` is an array of typed references. An event with several teams counts fully in each team's group, so team groups can sum to more than the tenant total.
- **Tags.** `tags` is a `Map(LowCardinality(String), String)`. The materialized column `tag_items` holds one token `base64(key):base64(value)` per pair, with a text index on it. Tag `eq` and `in` filters compile to `has` and `hasAny` on the tokens; every other tag operator reads the map. Ingestion rejects empty tag values.
- **Late usage.** `arrival` is `late` when `ingested_at > time + 1 hour`. Late events change totals of windows that have already ended.
- **Timestamps.** The worker writes `ingested_at = stored_at =` the Kafka message timestamp (`LogAppendTime`). `time` is the event's own time.
- **Unpriced events.** An event is unpriced when it reports a positive token quantity of any kind but its stored `reported_cost_usd` is zero.

## 4. Platforms and versions

| | ClickHouse Cloud, stage 1 production | ClickHouse Cloud, US dev | Altinity `prod-us` | Integration tests | Local check for this brief |
| --- | --- | --- | --- | --- | --- |
| Topology | Dedicated service in AWS us-east-2 | 3 replicas | 3 shards × 2 replicas, shared with other teams' production workloads | 1 node | `clickhouse local` |
| `usage_events` | `SharedReplacingMergeTree` (inferred: Cloud maps `Replicated*` engines to `Shared*`) | `SharedReplacingMergeTree`, observed | Planned: `ReplicatedReplacingMergeTree` local tables behind `Distributed` tables | `ReplacingMergeTree` | `ReplacingMergeTree` |
| Version | Not recorded | 26.4 (reported, unverified) | `26.3.16.10001.altinitystable`; the plan requires 26.3.33.10001 or later before the move | 26.8.2.7 | 26.7.4.58 |
| Network | HTTPS on 8443, public CA certificate | Same | HTTPS on 8443 over AWS PrivateLink through Altinity's edge proxy; certificate must chain to a public CA | — | — |
| State | Migrated and running since 2026-10-07; almost no traffic yet | Demo data | Trial-tier environment; no AI Cost Management objects yet; no Datadog in its account | — | — |

Version facts:

- 26.3.16 lacks two row-policy fixes that the tenant-isolation contract needs: the `mergeTreeProjection` fix (ClickHouse #112329, in 26.3.20.7) and the `mergeTreeIndex` fix (ClickHouse #115304, in 26.3.22.7; upstream 26.6.4.10 and 26.7.6.5). Altinity 26.3.33.10001 has both.
- Nobody has checked whether 26.3 accepts the table's features: `text(tokenizer = 'array')` indexes, `_part_offset` projections, `map_serialization_version = 'with_buckets'`, and `deduplicate_merge_projection_mode = 'rebuild'`.
- ClickHouse Cloud rejects `Distributed(...)`, so a sharded layout cannot come from migrations that run unchanged on both platforms.
- ClickHouse Cloud disables `_part_offset` columns in projections by default; the table enables them with `allow_part_offset_column_in_projections = 1`.
- ClickHouse Cloud fixes the custom-setting prefix `SQL_`. A self-managed server must list it in `custom_settings_prefixes`.
- Observed on US dev Cloud on 2026-09-09: an `ALTER ... SETTINGS mutations_sync = 2` returned before its part mutation ran. One `ALTER` that changed a column's type and added an index on that column applied its metadata at once, but its part mutation then failed forever, and every later `ALTER` was refused with error 517. Plain `MergeTree` in the integration suite accepted the same statement. Since then the migration runner polls `system.mutations` after every migration.

## 5. Decisions

Taken on 2026-09-28:

- **D-01** Production US launches on ClickHouse Cloud (stage 1), then moves its setup and data to `prod-us` once the data team's provisioning automation and quotas are ready, estimated at two to three weeks (stage 2).
- **D-02** Stage 1 uses a Cloud service dedicated to AI Cost Management. It stays up for one week after the switch-over as the rollback path.
- **D-03** US dev stays on ClickHouse Cloud. So `prod-us` will be the first sharded deployment anywhere.
- **D-04** One migration set, one backend image, and one Helm chart serve both platforms. Differences in platform and topology are configuration.
- **D-05** On `prod-us`, AI Cost Management uses its own databases on the cluster's six nodes and shards its data across the three shards.

Proposed, open to challenge:

- **D-06** Stage 1 migrates as the Cloud service's `default` user, because the provisioning provider cannot grant `CREATE ROLE`. This is a recorded exception to least privilege.
- **D-07** On a cluster there are two `Replicated` databases. `ai_cost_management_local` is sharded like the cluster and holds everything the existing migrations create, replicated within each shard. `ai_cost_management` has one shard across all six nodes and holds `Distributed` tables named like the local ones (`usage_events`, `usage_event_field_values`) plus the runner's `_migrations` table. The service always connects to `ai_cost_management`, so application SQL is the same on every platform.
- **D-08** On a cluster, regular migrations run in the local database. Cluster-only migrations (`NNN_name.cluster.sql`, with `${CLICKHOUSE_CLUSTER}` and `${CLICKHOUSE_LOCAL_DATABASE}` placeholders) run only when a cluster is configured. Any later change to a table behind a `Distributed` table ships with a cluster-only companion migration. The field-value backfill is skipped on a cluster, because the materialized view fires on each shard as rows arrive through the `Distributed` table, copied history included.
- **D-09** The `Distributed` tables shard by `cityHash64(tenant_id, id)`. Every copy of an event lands on one shard, so per-shard `FINAL` stays exact, and each tenant's usage spreads over all shards, so every tenant read fans out to all three. The worker inserts through the `Distributed` table with `distributed_foreground_insert = 1` and commits Kafka offsets only after every target shard has written its part.
- **D-10** `prod-us` defines an interserver secret, so a query that reaches another shard runs there as the user who started it. Row policies and grants stay on the local tables, where every shard enforces them. The reader keeps `serialize_query_plan = 0`.
- **D-11** The runner refuses to migrate a cluster unless both databases have the expected engines and shard counts and `clusterAllReplicas` shows the migrator's user on every replica. After migrating, it refuses a `Distributed` table whose columns differ from its local table.
- **D-12** Every migration statement runs with `alter_sync = 2` and `mutations_sync = 2`, so a replica outage fails the migration instead of leaving a partial change. The `system.mutations` poll stays for Cloud.
- **D-13** On a cluster, the notifier's threshold query checks visibility itself: it fails through `throwIf` (error 395) when the stored Kafka offset it needs is not visible yet, and the consumer retries within its existing visibility budget. The read sets `load_balancing = 'in_order'`, so the check and the evaluation read the same replica of each shard; an append-only migration lets the reader profile change that setting. The separate stored-offset probe (B5) is removed.
- **D-14** The data team provisions both databases, the roles, the users, and a least-privilege migrator on `prod-us` through reviewed automation.
- **D-15** The service connects over HTTPS on port 8443 and never uses the native port 9440.
- **D-16** `prod-us` runs Altinity Stable 26.3.33.10001 or later before the move, and a sharded integration suite runs that build.
- **D-17** The move copies `usage_events` from the stage 1 service into `prod-us` through the `Distributed` table twice: a bulk copy up to an `ingested_at` watermark while stage 1 serves production, and a final overlapping copy after the switch-over. Rows copied twice, and batches the worker retries across the switch-over, collapse under `FINAL` because D-09 sends every copy of an event to the same shard. Ingestion does not pause. The copy path is still undecided: `prod-us` pulling with `remoteSecure`, or an export through S3. Reconciliation compares per-tenant row counts and `FINAL` spend totals below the watermark.

Rejected alternatives (do not re-propose them without new evidence):

- Waiting for `prod-us` before launching.
- Launching on OpenMeter's shared tier-0 Cloud service.
- One shard over the six nodes. It needs no `Distributed` tables, but keeps six copies of the data and leaves the cluster's shards unused.
- An Altinity-only schema, or `Distributed` tables created by hand. Either breaks D-04.
- Pausing the worker for the move.
- Quorum inserts with `select_sequential_consistency` instead of D-13. On `ReplicatedMergeTree`, sequential consistency requires `insert_quorum_parallel = 0`, which rejects concurrent inserts into a table, and a quorum insert fails while a replica restarts.
- Accepting stale evaluations as delays. On a replicated cluster the probe no longer proves that the evaluating replica holds the usage, so replica lag would delay a threshold crossing until the tenant's next usage arrives.

Sharding by `tenant_id` alone does not appear among the rejected alternatives.

## 6. Requirements and constraints

Correctness:

- An authoritative read counts each event exactly once, whatever the redelivery, retry, or copy history.
- A report, a budget balance, and an alert evaluation over the same scope and range agree to the last decimal place.
- The move to `prod-us` keeps every tenant's usage history and neither loses nor double-counts an event.

Tenant isolation:

- The isolation contract in section 8.3 holds on every shard and replica. The reader and writer users hold only their roles, the row policies and the reader settings profile stay as specified, and the verification gate in section 9.10 passes before a platform serves production data.

Platform and migrations:

- One migration set, one image, and one chart (D-04). Migrations are forward-only and append-only: an applied migration is never edited, and corrections ship as new migrations.
- Each migration file holds exactly one statement. The runner sends it with `wait_end_of_query = 1` and records its checksum.
- Migrations never use `ON CLUSTER` today. On Cloud, the database engine shares definitions across replicas.

Operations:

- `prod-us` is shared. AI Cost Management's load, migrations, and data copy must not degrade other teams' workloads, and changes to the cluster go through the data team.
- Provisioning on both platforms comes from reviewed automation, not console changes.
- The writer runs with `readonly = 0` and no `async_insert`, and replicated insert deduplication stays enabled.
- An insert block may span at most 100 partitions, which is why HTTP ingestion bounds `time`. Gateway usage has no such bound.
- An export connection can stay idle for up to 330 seconds while the server finishes a chunk, so every proxy on the path must tolerate that. Whether Altinity's edge proxy does is unknown.

## 7. Volumes, targets, and measurements

Tags: **[M]** measured, **[T]** target, **[P]** placeholder default, **[U]** unknown.

- **[U]** Production event rate, tenant count, largest tenant, and growth are unknown. Production is in a pilot: Kong's internal organizations first, then named customers one at a time. The stage 1 service's size is not recorded. The production load test (3× the expected peak for one hour on US dev, recording worker throughput, insert latency, lag recovery, ClickHouse CPU and memory, and API p95 per route for the largest organization) has not run.
- **[T]** API: 99.9% SLA with 99.95% objectives for availability (share of requests without a 5xx over 30 days) and latency (p95 under 2 s in each one-minute slice). Monitors warn at an API p95 of 1 s and page at 2 s; they page when the worker's time lag exceeds 900 s, the notifier's lag exceeds 1,800 s, or consumer lag exceeds 100,000 records. GA requires at least 3× headroom over the pilot peak.
- **[P]** Kafka: 3 partitions in dev, 12 proposed for production; 3 worker replicas, scalable to 12; batches of up to 10,000 records. The documents disagree on topic retention (one day versus seven days); treat retention as at least one day.
- **[P]** Reader limits: `max_execution_time` 20 s for API and notifier queries. Server-side reader limits (`max_execution_time`, `max_memory_usage`, `max_result_rows`) and a 13-month TTL on `usage_events` are planned but not built.
- **[M]** Tag filters on a local 26.7.4 server with 3M synthetic events, read with `FINAL`: `tags[k] = v` read all 372 granules (265 MiB), while `has(tag_items, token)` read 3 of 372 granules (1.7 to 3 MiB) with identical counts. `tag_items` plus `idx_tag_items` took 29% of the table's bytes in that data set.
- **[M]** Field-value lookup on a local server, 50,000 events, medians of 5 runs: inserts took 128 ms without the materialized view and 173 to 200 ms with it. A picker lookup took 2.3 to 3.6 ms, against 5.5 to 6.1 ms for searching raw events. With 300,000 subjects and 2,000,000 tag values on 26.7, a subject page read 24,576 rows instead of 300,000.
- **[M]** Exports on a local open-source server, one day of 1M synthetic events: CSV peaked at 311 MB of memory with 2 threads and 116 MB with 1. Parquet peaked at 401 MB with 64 MiB row groups and 351 MB with 32 MiB. The default 512 MiB Parquet row group exceeded the 512 MiB query limit, and ordering by `time, source, id` alone also exceeded it. The CSV result was 609 MB; server-side gzip added 0.9 s and produced 79 MB.

## 8. Schema

### 8.1 `usage_events`

The team's migrations produce this table. The DDL is `SHOW CREATE TABLE` output after replaying the migrations on ClickHouse 26.7.4, so some details reflect that build: it renders the text indexes' granularity as `GRANULARITY 100000000`, and it inlines the `tag_token` function into the `tag_items` expression. In deployed environments the engine is `ReplicatedReplacingMergeTree` with server-default replication paths, which ClickHouse Cloud turns into `SharedReplacingMergeTree`.

DDL: [`sql/schema/01_usage_events.sql`](sql/schema/01_usage_events.sql).

Notes:

- The original migration carries two unresolved comments. It says the `by_subject` and `by_model` projections need evaluating ("TODO: evaluate whether these are needed"), and it asks for the best `max_buckets_in_map` and `map_buckets_strategy`.
- The current `by_subject` projection was added in a later migration and materialized with `ALTER TABLE usage_events MATERIALIZE PROJECTION by_subject SETTINGS mutations_sync = 2`.
- `subject` replaced the columns `subject_type` and `subject` through an add, materialize, drop, and rename sequence of single-statement migrations. `teams` changed from `Array(UUID)` to `Array(String)` in place.
- Only the event list (C1) and the event export (D1) read `type`, `traceparent`, `tracestate`, `workspace_id`, `plugin_id`, `request_id`, and the semantic-cache and compression columns. No statement reads `stored_at`, and only the probe B5 reads `kafka_topic`, `kafka_partition`, and `kafka_offset`.

### 8.2 Field-value lookup

- **Lookup table:** [`sql/schema/02_field_values.sql`](sql/schema/02_field_values.sql). It has no partitions and no TTL. Its rows come from the materialized view, which fires on every insert into `usage_events`.
- **Materialized view:** [`sql/schema/05_mv.sql`](sql/schema/05_mv.sql).
- **SQL UDFs:** [`sql/schema/03_functions.sql`](sql/schema/03_functions.sql). The view calls `usage_event_field_value_observations`; `tag_items` was built from `tag_token`.
- **Backfill progress table:** [`sql/schema/06_backfill_table.sql`](sql/schema/06_backfill_table.sql). The one-time backfill records its progress there.

### 8.3 Access objects and tenant isolation

Statements: [`sql/schema/04_access.sql`](sql/schema/04_access.sql). They define:
- the reader and writer roles;
- the permissive base and restrictive tenant-guard row policies on both tables;
- the definer user, which has an unprovisioned hash and `HOST NONE`;
- the `ai_cost_management_reader_settings` profile.

The isolation contract:

- Every reader query carries a typed tenant predicate and the `SQL_tenant_id` setting. The restrictive policy rejects the other tenants' rows even when the predicate is missing, and a query without the setting fails, because the profile's default `''` makes `toUUID` throw.
- The reader cannot write, mutate, copy partitions, use `url`, `remote`, or `file`, or change protected settings. `serialize_query_plan = 0` is pinned.
- For `Distributed` tables, the policies must exist on every local table, because a policy on the `Distributed` table is not forwarded to the shards.
- `ai_cost_management_reader_settings` declares the export bounds as constraints: `max_execution_time MAX 300`, `max_memory_usage MAX 536870912`, and `max_threads MAX 2`, each changeable in read-only mode. The profile does not set their values.
- Tenant erasure is `ALTER TABLE ... DELETE WHERE tenant_id = ... SETTINGS mutations_sync = 2` on both tables. Only the single-table form is documented; there is no form for a cluster's local tables.

## 9. Use cases and the SQL they send

The statements come from the service's own query builders, run against a fake client, and every one executes against the schema above on ClickHouse 26.7.4. The service sends server-side parameters such as `{tenant: UUID}`, `{from: DateTime64(3, 'UTC')}`, or `{candidates: Array(String)}`; the stored files replace them with equivalent literals, so the statements run as written. Every reader query also runs under the restrictive row policy, so the real plan carries a second tenant predicate. The example tenant is `00000000-0000-4000-8000-000000000001`.

### 9.1 Ingestion

- **Trigger:** each Kafka partition batch of up to 10,000 records. There is no buffering across batches and no flush interval. A batch mixes tenants and may span several monthly partitions.
- **Statement:** `INSERT INTO usage_events FORMAT JSONEachRow` with `date_time_input_format = 'best_effort'`. There is no `async_insert`, no `insert_deduplication_token`, and no quorum setting. The worker writes every stored column except the two `MATERIALIZED` columns `arrival` and `tag_items`.
- **Failures:** a transport or server failure backs off from 1 s to 30 s and then redelivers the batch. A content error bisects the batch recursively, so one bad record in 10,000 costs about 28 inserts. Records the store refuses on their own are logged and dropped.
- **Deduplication:** `non_replicated_deduplication_window` locally, the replicated window in production. Insert deduplication only catches a byte-identical block, and a redelivered batch can have different boundaries, so merges and `FINAL` do the real work.
- **On a cluster (D-09):** inserts go through the `Distributed` table with `distributed_foreground_insert = 1`. Each block splits by shard and by partition, and the materialized view fires on each shard.

Judge: part creation and merge pressure per insert on each platform; how insert deduplication behaves when a `Distributed` table splits a block; whether `distributed_foreground_insert = 1` makes one slow shard stall every worker; whether the materialized view's `SELECT DISTINCT` and `ARRAY JOIN` cost grows with tag and team counts; and the cost of `ReplacingMergeTree` merges on a shared cluster.

### 9.2 Console aggregates

The console's usage explorer and dashboard call `POST /v0/events/query`. One explorer page view sends three statements built from the same URL state: ungrouped totals with buckets for the chart header (A1), the top few subjects with buckets for the chart (A2), and a paginated subject table without buckets (A3).

Request limits:

- At most two dimensions: `subject`, `subject_type`, `subject_id`, `provider`, `model`, `model_id`, `request_mode`, `region`, `client`, `source`, `team`, the four gateway UUIDs, or `tags.<key>`.
- Measures: `events`, `unpriced_events`, both costs, and any of the 22 quantity kinds.
- Optional granularity `PT1H`, `P1D`, `P1W`, or `P1M`, in an IANA time zone, with at most 400 buckets.
- Page size 20 by default and at most 100. Pagination uses keyset cursors over the sort attributes (default `reported_cost_usd desc`), with the dimensions as tiebreakers and `null` groups last.
- Filters combine with AND and allow nested `and` and `or` groups, at most ten per group, with up to 100 values per list.

The statement shape: the inner select aggregates per bucket and group, the middle select folds the buckets into range totals plus an `arraySort(groupArray(tuple(...)))` series, and the outer select adds a null flag per dimension and applies the keyset, order, and `limit size + 1`. Without granularity, the middle select aggregates the range directly. Every page recomputes every group in the range.

| Statement | Description | File |
| --- | --- | --- |
| A1 | Explorer totals: ungrouped, daily buckets in `Europe/Budapest` | [`sql/queries/a1_explorer_totals.sql`](sql/queries/a1_explorer_totals.sql) |
| A2 | Explorer chart: top subjects, daily buckets | [`sql/queries/a2_explorer_top_subjects.sql`](sql/queries/a2_explorer_top_subjects.sql) |
| A3 | Explorer table: subjects without buckets, with provider, model, team, and tag filters. The tag `in` filter compiles to `hasAny(tag_items, ...)`. | [`sql/queries/a3_explorer_table_filtered.sql`](sql/queries/a3_explorer_table_filtered.sql) |
| A4 | Two dimensions: `team` (an `arrayJoin`) by `tags.cost_center` (a map lookup), weekly buckets over a quarter, sorted by dimension | [`sql/queries/a4_team_by_costcenter_weekly.sql`](sql/queries/a4_team_by_costcenter_weekly.sql) |

Judge: the cost of `FINAL` with and without the projections; whether the `by_subject` and `by_model` projections are ever chosen for these shapes under `FINAL` (`_part_offset` projections act as indexes, not as pre-aggregations); how the three-level aggregation, `groupArray` series, and keyset `limit` execute over a `Distributed` table, and which part runs on the initiator; whether `optimize_skip_unused_shards` or `distributed_group_by_no_merge` can ever apply under D-09; and how cost grows with range length, group count, and page depth.

### 9.3 Forecast

`GET /v0/events/forecast` sends one daily aggregate (A5) from the earlier of the previous month's start and 28 days ago up to the start of today, then one table-free kernel query (A6) when enough history exists.

| Statement | Description | File |
| --- | --- | --- |
| A5 | Forecast history on 2026-10-07 in UTC | [`sql/queries/a5_forecast_history.sql`](sql/queries/a5_forecast_history.sql) |
| A6 | The three forecast kernels, with four training days and a 25-day horizon | [`a6a_kernel_weekday.sql`](sql/queries/a6a_kernel_weekday.sql), [`a6b_kernel_linear.sql`](sql/queries/a6b_kernel_linear.sql), [`a6c_kernel_decay.sql`](sql/queries/a6c_kernel_decay.sql) |

### 9.4 Budget balances

- **Trigger:** HTTP requests from the console. A budget has a scope (filters, optionally split by one dimension, `each`) and a window. The window is any ISO 8601 interval from 1 second to 100 years, anchored in an IANA time zone and computed with calendar arithmetic; a calendar month is the common case.
- **Balance** (`GET /budgets/{id}/balance`): B1a, then B1b only when the scope has an `each` dimension and the window has events. That makes one or two statements per request.
- **Entity balances** (`GET /budgets/{id}/balances`): a totals statement and B2 run in parallel, so two statements per request. B2 adds zero rows for named entities through `union all` and pages with keyset cursors.
- Team scope predicates compile to `has` or `hasAny` on `teams`; negations also require a non-empty list.

| Statement | Description | File |
| --- | --- | --- |
| B1a | Budget balance: the scoped window total | [`sql/queries/b1a_budget_total.sql`](sql/queries/b1a_budget_total.sql) |
| B1b | Budget balance: the top subject groups in the same scope | [`sql/queries/b1b_budget_top_subjects.sql`](sql/queries/b1b_budget_top_subjects.sql) |
| B2 | Entity balances: team groups with a named candidate team | [`sql/queries/b2_entity_balances_teams.sql`](sql/queries/b2_entity_balances_teams.sql) |

### 9.5 Budget alert evaluation (notifier)

- **Trigger:** every Kafka batch the notifier consumes, up to 10,000 records from one partition. The notifier groups the batch by tenant and holds it until 30 s after its newest record. It then processes tenants one after another, and only tenants with alert rules on limit budgets touch ClickHouse.
- **Per tenant:** the visibility probe B5 runs until the stored offset reaches the batch's highest offset for that tenant, sleeping 3 s between tries, for at most 30 s shared by the whole batch. After that it logs a warning and evaluates anyway. Then one threshold statement (B3, or B4 for a scope without `each`) runs per rule, sequentially. The window's end is the tenant's newest record time, not the wall clock. The minimum is the lowest threshold of the rule, so the query returns only groups that can fire, and JavaScript compares the exact thresholds.
- **Fan-out:** per batch, the sum over its tenants with rules of (1 to 10 probes + one statement per rule). Producers assign partitions at random or round-robin, so one tenant's events land on every partition, and a tenant can be evaluated once per partition in each hold window: up to 12 times per cycle with the proposed production partition count.
- A ClickHouse timeout skips that tenant for the batch. Hold, visibility wait, and failure backoff must stay under Kafka's 300 s maximum poll interval.
- D-13 replaces B5 on a cluster with a `throwIf` inside the threshold statement and `load_balancing = 'in_order'`.

| Statement | Description | File |
| --- | --- | --- |
| B3 | Threshold: subjects in an OpenAI-scoped monthly budget that reached $800 | [`sql/queries/b3_threshold_each_subject.sql`](sql/queries/b3_threshold_each_subject.sql) |
| B4 | Threshold: a pooled scope, with no `each` dimension and no minimum pushdown | [`sql/queries/b4_threshold_pooled.sql`](sql/queries/b4_threshold_pooled.sql) |
| B5 | Visibility probe: the worker's stored offset for one tenant and partition. It runs without `FINAL` and relies on the `idx_ingested_at` minmax index. | [`sql/queries/b5_visibility_probe.sql`](sql/queries/b5_visibility_probe.sql) |

Judge: this is the service's main recurring load. Every batch reruns full-window `FINAL` aggregations per rule, and the cost grows with the window's size, the number of rules, and the partition count. Compare how read-after-write visibility works on Cloud SharedMergeTree and on `prod-us` replicas, judge D-13 on both, and say whether this workload justifies a different design (an incremental or pre-aggregated form) and what that would cost in correctness and complexity.

### 9.6 Usage event list

`GET /v0/events` returns newest first with keyset pagination on `(time desc, source, id)`. The time filter is optional, so a request without one reads every month of the tenant.

| Statement | Description | File |
| --- | --- | --- |
| C1 | Event list: anthropic, `tags.env = prod`, since 2026-10-01 | [`sql/queries/c1_event_list.sql`](sql/queries/c1_event_list.sql) |

Judge: whether `ORDER BY time desc LIMIT 21` with `FINAL` reads in order, on one node and over a `Distributed` table; what an unbounded request costs for a large tenant; and whether the `id` bloom filter helps any query.

### 9.7 Field-value pickers

`GET /v0/events/field-values` reads the lookup table without `FINAL` and relies on `DISTINCT`, ordered by the physical key so that ClickHouse can stream the `DISTINCT`. Fields are `subject`, the context strings, `teams`, the gateway UUIDs, `consumer_group_ids`, `tags` without a key (tag keys), or `tags` with a key (that key's values). Search is a case-insensitive substring.

| Statement | Description | File |
| --- | --- | --- |
| C2 | Subject picker search | [`sql/queries/c2_picker_subject.sql`](sql/queries/c2_picker_subject.sql) |
| C3 | Tag keys | [`sql/queries/c3_picker_tag_keys.sql`](sql/queries/c3_picker_tag_keys.sql) |

Without a cluster, the migration job backfills the lookup table once, one `usage_events` partition at a time, after migrating (`SELECT DISTINCT partition_id FROM system.parts ...` lists the partitions). It runs with `max_threads = 2`, `max_memory_usage = 536870912`, `max_bytes_before_external_group_by = 67108864`, and `max_execution_time = 300`. Statement: [`sql/queries/z_backfill_partition.sql`](sql/queries/z_backfill_partition.sql).

Judge: duplicate lookup rows across shards under D-09, since every shard observes the same tenant values; whether `DISTINCT` over a `Distributed` table still streams; the lookup table's growth with no TTL; and whether the materialized view adds measurable insert cost on a shared cluster.

### 9.8 Exports

- **Trigger:** the exporter claims one export at a time and polls every 5 s. An export covers at most 13 months. It runs one statement per chunk, one after another: a civil day for event exports (D1) and a calendar month for FOCUS cost exports (D2). The result streams to S3.
- **Settings per statement:** `max_execution_time = 300`, `max_memory_usage = 536870912`, `max_threads = 2`, `wait_end_of_query = 1`, and `cancel_http_readonly_queries_on_client_close = 1`. The reader profile pins the Parquet settings (zstd, 64 MiB row groups) and CSV settings. Because of `wait_end_of_query`, the server holds the whole chunk in temporary storage outside the memory limit, and the connection is idle until the chunk is finished.

| Statement | Description | File |
| --- | --- | --- |
| D1 | Event export, one day, Parquet | [`sql/queries/d1_event_export_day.sql`](sql/queries/d1_event_export_day.sql) |
| D2 | FOCUS 1.3 cost export, one month, daily charge periods, two named tag keys | [`sql/queries/d2_focus_export_month.sql`](sql/queries/d2_focus_export_month.sql) |

Judge: temporary storage and memory on each platform for the largest plausible tenant-day and tenant-month; whether the `order by toStartOfMinute(time), time, source, id` read-in-order trick survives a `Distributed` table, where the initiator must merge sorted streams from three shards; the idle-connection risk through Altinity's edge proxy; and whether exports can starve interactive queries on a shared cluster.

### 9.9 Migrations, mutations, and retention

- The runner sends one statement per file, records an MD5 checksum per file, and refuses an applied migration that changed on disk.
- Before and after every migration it polls `system.mutations` on the replica it reached, and fails on any `latest_fail_reason`.
- Mutation-producing migrations so far: `MODIFY COLUMN` of a decimal precision, a combined drop index, drop column, change type, and add index (the one that failed on Cloud), `DROP PROJECTION`, `MATERIALIZE COLUMN`, and `MATERIALIZE PROJECTION`, each with `mutations_sync = 2`.
- Operators run `ALTER ... DELETE` for tenant erasure and `KILL MUTATION` for a stuck mutation.
- Planned: a 13-month TTL on `usage_events` and server-side reader limits, both through new migrations.
- On a cluster, D-08, D-11, and D-12 apply.

Judge: how each mutation type behaves on Cloud SharedMergeTree compared with a 3 × 2 `ReplicatedMergeTree` cluster with `Replicated` databases; what a TTL on monthly partitions costs on each platform; and how tenant erasure must work across local tables.

### 9.10 Tenant isolation across shards

The verification gate, which must pass before a platform serves production data and again after every upgrade or move to `Distributed` tables:

1. A reader query without `SQL_tenant_id` fails.
2. With tenant A's setting, a query returns only tenant A's rows, even without the predicate.
3. Tenant A's setting with tenant B's predicate returns nothing.
4. Subqueries, joins, projections, `FINAL`, and every exposed table or view keep the policy.
5. The reader cannot write, mutate, copy partitions, use privileged table functions, or change protected settings.
6. Application queries send matching tenant values.
7. The access system tables match the intended state.
8. `EXPLAIN indexes = 1` shows tenant and time pruning.
9. The whole gate reruns after every upgrade and after any move to `Distributed` tables.

The projection-index path (enabled at `min_table_rows_to_use_projection_index`, 1M rows by default on 26.8) is not covered by the team's tests.

Judge: whether custom settings such as `SQL_tenant_id` reach remote shards with the interserver secret; how a shard treats settings constraints such as `max_threads MAX 2` on settings it receives from the initiator; whether access objects, SQL user-defined functions, and the definer user must be created per replica or can be replicated (`user_defined_zookeeper_path`, replicated access storage); and which gate items need new cross-shard cases.

### 9.11 The move from Cloud to `prod-us`

See D-17. Judge: `remoteSecure` pull against an S3 round trip for the copy, including network path, credentials, throughput, and the load on a shared cluster; whether the `ingested_at` watermark plus an overlapping final copy is exact under `FINAL` given D-09; how the materialized view behaves on copied history; and what the reconciliation must compare to prove nothing was lost or doubled.

## 10. Questions to answer

1. Does 26.3.33 accept every table feature in section 8.1, with the same behavior as 26.4 and 26.7? Which settings or experimental flags does each need?
2. Under D-09, every tenant read fans out to three shards. Compare `cityHash64(tenant_id, id)` with sharding by `tenant_id` alone, which would route each read to one shard through `optimize_skip_unused_shards`. Cover tail latency, skew from large tenants, `FINAL` exactness, and the move. Which would you choose for this workload, and on what evidence?
3. Is `FINAL` partition-local here without `do_not_merge_across_partitions_select_final`? What does `FINAL` cost on each platform at plausible volumes, and do settings such as `enable_vertical_final` change the answer?
4. Are the `by_subject` and `by_model` projections used by any statement in section 9? If not, should they go? What do they cost in inserts, merges, and storage?
5. Do the reader profile's constraints (`max_threads MAX 2`, `max_memory_usage MAX 536870912`) limit interactive queries, either on the node that receives the query or on remote shards?
6. What is the best `max_buckets_in_map` and `map_buckets_strategy` for `tags`, given that most queries read one key or none?
7. Is the notifier's recurring full-window evaluation (section 9.5) sustainable at 12 partitions, many rules, and monthly windows on each platform? If not, what is the simplest exact alternative?
8. Is D-13 (`throwIf` plus `load_balancing = 'in_order'`) correct on `prod-us`, and what is the equivalent guarantee on Cloud SharedMergeTree, where the probe runs today?
9. Which use cases does the shared `prod-us` cluster put at risk of noisy-neighbor effects in either direction, and which quotas, settings profiles, or workload limits should the data team set?
10. What does each platform need for backups, restores, upgrades, and monitoring, given that `prod-us` has no Datadog and its environment is still marked as a trial?
11. Which measurements should the team take on US dev Cloud and on `prod-us` before the move, and what results would block it?

## 11. Out of scope

Do not re-raise these:

- *(First item omitted from this copy: an accepted network-access risk on the Cloud services.)*
- Tag filters use tokens encoded by the service rather than the `tag_token` SQL function, and an empty tag value means an absent tag. Both are decided.
- Text indexes on `mapKeys(tags)` or `mapValues(tags)`. They are avoided because ClickHouse #122358 drops rows that lack the key for patterns that match the empty string.
- The rejected alternatives in section 5.
- Query-time pricing, enforcement, and a credit ledger. None is built; do not design them here.

## 12. Output

1. **Summary:** at most ten bullets with the conclusions that matter most.
2. **Use-case comparison:** a table with the columns use case, ClickHouse Cloud, Altinity `prod-us`, verdict, and confidence, with one row per section 9 subsection.
3. **Findings**, ranked by severity. For each: the use case, the platforms affected, the evidence label (verified, inferred, or assumed), the impact, the recommendation, the effort, and any constraint or decision it touches.
4. **Proposed decisions:** for each of D-06 to D-17 that you have a view on, keep, change, or replace, with the reason.
5. **Measurement requests:** exact read-only statements for the team to run on US dev Cloud, the stage 1 Cloud service, and `prod-us`, each with the result that would change a recommendation.
6. **Open questions** you could not resolve, each with what would resolve it.

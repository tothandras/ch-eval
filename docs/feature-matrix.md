# Feature matrix: ClickHouse Cloud vs Altinity Stable vs upstream OSS

> **Status: DRAFT (Phase 1). Every cell is a hypothesis taken from a cited
> primary source, not a probed result.** In Phase 2 this file becomes
> generated output: `bench feature-matrix` renders it from
> `docs/feature-matrix.yaml` plus the probe results under
> `results/probes/<target>/<run-id>/`. Once that happens, each cell shows the
> probed status and the exact `version()` it was probed on. Rows whose probe has
> not run on a target keep the hypothesis marker.

Targets and versions (pinned in [`research/versions.md`](research/versions.md)):

| Target | Planned version |
| --- | --- |
| **cloud** | ClickHouse Cloud. 26.6 on the Fast/Regular channels and 26.4 on Slow as of 2026-10-08. SharedMergeTree on object storage. |
| **altinity** | `26.3.33.10001.altinitystable`, Altinity operator 0.27.4 plus ClickHouse Keeper |
| **upstream-oss** | `26.3.33.24`, same base version and same topology as altinity |

Legend:

| Mark | Meaning |
| --- | --- |
| ✅ | available |
| ❌ | not available |
| ⚠️ | available with caveats |
| ❓ | unverified in sources |
| *(h)* | hypothesis, not yet probed |
| 🔬 | observed ad hoc in the 2026-10-08 local smoke run; a formal probe is still pending |

Citations point to the published docs URL. Docs text was verified against the
docs source (`ClickHouse/ClickHouse/docs` at master `cd023af`) because
clickhouse.com is unreachable from the build sandbox.

**OpenMeter-relevant rows are bold.** Each is a candidate for the design-decision
callout at the top of `REPORT.md` once probed.

## 1. Storage and replication

| ID | Capability | cloud | altinity / upstream-oss | Source | Planned probe |
| --- | --- | --- | --- | --- | --- |
| **ST-01** | **Engine actually used for `ENGINE = MergeTree` / `ReplicatedMergeTree`** | SharedMergeTree; Cloud "automatically substitutes the corresponding `Shared*` engine" *(h)* | As written. Plain `MergeTree` on a 2-replica cluster creates independent, unreplicated tables per replica. This is OpenMeter's hard-coded engine. *(h)* | [cloud-compatibility](https://clickhouse.com/docs/products/cloud/guides/cloud-compatibility) | `probes/storage/engine_conversion.sql` |
| ST-02 | Database engine | `Shared` database engine "builds on and improves the behavior of the Replicated database engine" *(h)* | `Atomic` by default; `Replicated` available *(h)* | [shared-catalog](https://clickhouse.com/docs/products/cloud/features/infrastructure/shared-catalog) | `probes/storage/database_engine.sql` |
| **ST-03** | **`ON CLUSTER` DDL** | Accepted with cluster `default`. "Mostly no-op" wording is gone from the docs, so ❓ | Required to fan DDL out to replicas, unless a `Replicated` DB is used *(h)* | [execute-system-queries-in-cloud](https://clickhouse.com/docs/resources/support-center/knowledge-base/cloud-services/execute-system-queries-in-cloud) | `probes/storage/on_cluster.sql` |
| **ST-04** | **Part replication model** | Shared object storage. Replicas share parts and synchronize metadata via Keeper. *(h)* | Each replica fetches and stores its own copy of every part. On S3 disks without zero-copy, each replica writes its own S3 objects. *(h)* | [shared-merge-tree](https://clickhouse.com/docs/products/cloud/features/infrastructure/shared-merge-tree), [storing-data](https://clickhouse.com/docs/concepts/features/configuration/server-config/storing-data) | Storage bytes per replica in every perf run (`system.parts` per host) |
| ST-05 | Zero-copy replication on S3 | n/a | ⚠️ Off by default, "not recommended for production use" *(h)* | [storing-data](https://clickhouse.com/docs/concepts/features/configuration/server-config/storing-data) | `probes/storage/zero_copy_setting.sql` |
| **ST-06** | **Cross-replica read-after-write visibility** | Metadata visibility is asynchronous. `select_sequential_consistency` and `SYSTEM SYNC REPLICA LIGHTWEIGHT` are available. *(h)* | Asynchronous replication with no fixed bound. `insert_quorum` + `select_sequential_consistency` is the documented hardening. *(h)* | [replication](https://clickhouse.com/docs/reference/engines/table-engines/mergetree-family/replication), [SMT consistency](https://clickhouse.com/docs/products/cloud/features/infrastructure/shared-merge-tree#consistency) | **`replica-consistency` workload** (stale-read rate, visibility lag, monotonicity violations) |
| ST-07 | Filesystem cache for remote disks; `enable_filesystem_cache=0` bypass | ✅ *(h)* | ✅ on S3-backed storage policies only ("cache is supported only for remote disk types") *(h)* | [storing-data](https://clickhouse.com/docs/concepts/features/configuration/server-config/storing-data), [settings](https://clickhouse.com/docs/reference/settings/session-settings) | `probes/storage/fs_cache.sql` |
| ST-08 | Dropping caches for cold reads | `SYSTEM DROP FILESYSTEM CACHE` ✅. `DROP MARK/UNCOMPRESSED CACHE` ❓. OS page cache is not reachable. *(h)* | All `SYSTEM DROP … CACHE`, plus the OS page cache on the node *(h)* | [execute-system-queries-in-cloud](https://clickhouse.com/docs/resources/support-center/knowledge-base/cloud-services/execute-system-queries-in-cloud) | `probes/storage/drop_caches.sql` |
| ST-09 | Distributed cache | Cloud-only ("Only has an effect in ClickHouse Cloud"). GA ❓ *(h)* | ❌ *(h)* | [distributed-cache settings](https://clickhouse.com/docs/reference/settings/session-settings/distributed-cache) | `probes/storage/distributed_cache.sql` |

## 2. Version-gated features (relevant to these workloads)

| ID | Feature | Introduced → GA | cloud (26.4/26.6) | altinity / upstream-oss (26.3) | Source | Planned probe |
| --- | --- | --- | --- | --- | --- | --- |
| **VG-01** | **`JSON` type** (OpenMeter proposal schema) | 24.8 → **25.3** | ✅ *(h)* | ✅ *(h)* | PR [#77785](https://github.com/ClickHouse/ClickHouse/pull/77785) "`JSON` data type is production-ready." | `probes/version/json_type.sql` |
| **VG-02** | **`ALTER … MODIFY COLUMN` String → JSON** | 24.11 | ✅ *(h)* | ✅ *(h)* | PR [#70442](https://github.com/ClickHouse/ClickHouse/pull/70442); [newjson](https://clickhouse.com/docs/reference/data-types/newjson) "only `ALTER` from a `String` type is supported" | `migrations/openmeter/*_data_string_to_json` (measured online change) |
| VG-03 | `text` (full-text) index | 23.1 (`inverted`) → **26.2** | ✅ *(h)* | ✅ *(h)* | PR [#96794](https://github.com/ClickHouse/ClickHouse/pull/96794) "The text index is now GA." | `probes/version/text_index.sql` |
| VG-04 | Lightweight `DELETE` | 22.8 → **23.3** | ✅ *(h)* | ✅ *(h)* | 23.3 changelog | `probes/version/lightweight_delete.sql` |
| VG-05 | Lightweight `UPDATE … SET` | 25.7 → **beta (not GA through 26.9)** | ❓ Beta features on Cloud *(h)* | ⚠️ beta, `enable_lightweight_update` *(h)* | [update statement](https://clickhouse.com/docs/reference/statements/update) "Lightweight updates are currently beta." | `probes/version/lightweight_update.sql` |
| VG-06 | Query condition cache | 25.3 → default on **25.4** | ✅ *(h)* | ✅ *(h)* | PR [#79080](https://github.com/ClickHouse/ClickHouse/pull/79080) | `probes/version/query_condition_cache.sql` |
| VG-07 | Parallel replicas | beta 24.10 → production tier ≈25.12 (inferred from source: BETA tier flag present at 25.11, absent from 25.12) | ✅ with `cluster_for_parallel_replicas='default'` *(h)* | ⚠️ needs a cluster definition; reads replicated parts *(h)* | [parallel-replicas](https://clickhouse.com/docs/products/cloud/features/infrastructure/parallel-replicas) | `probes/version/parallel_replicas.sql` |
| VG-08 | Refreshable materialized views | 23.12 → **24.10** | ✅ *(h)* | ✅ *(h)* | PR [#70550](https://github.com/ClickHouse/ClickHouse/pull/70550) | `probes/version/refreshable_mv.sql` |
| VG-09 | Projections over `JSON` subcolumns | ❓ (not documented) | ❓ | ❓ | [newjson](https://clickhouse.com/docs/reference/data-types/newjson) mentions a projection sort key that reads a subcolumn, which is indirect evidence only | `probes/version/projection_json.sql` |
| VG-10 | `_part_offset` projections; lightweight update/delete with projections | 25.5; LW mutations "aren't supported for tables with projections" | ✅ / ⚠️ *(h)* | ✅ / ⚠️ *(h)* | [projections](https://clickhouse.com/docs/concepts/features/projections/projections) | `probes/version/projection_lw_mutation.sql` |
| VG-11 | Vector similarity index (low priority) | 24.8 → **25.8** | ✅ *(h)* | ✅ *(h)* | PR [#85888](https://github.com/ClickHouse/ClickHouse/pull/85888) | — |

## 3. Ingest semantics

| ID | Behavior | cloud | altinity / upstream-oss | Source | Planned probe |
| --- | --- | --- | --- | --- | --- |
| **IN-01** | **`async_insert` default** | 1 since 26.2 *(h)* | 1 since 26.2 🔬 (`getSetting('async_insert')` returned `true` on both 26.3 images) | [async-insert setting](https://clickhouse.com/docs/reference/settings/session-settings/async-insert) "26.2 … Enable async inserts by default." | `probes/ingest/async_defaults.sql` |
| **IN-02** | **`wait_for_async_insert` default; flush timeout** | 1; "200 ms or 1000 ms on Cloud" *(h)* | 1; 200 ms *(h)* | [asyncinserts](https://clickhouse.com/docs/concepts/features/operations/insert/asyncinserts) | `probes/ingest/async_defaults.sql` plus time-to-queryable measurement |
| **IN-03** | **Insert deduplication default** | On for all inserts since 26.2 (`deduplicate_insert=enable`). `async_insert_deduplicate` superseded. *(h)* | Same *(h)* | PRs [#95970](https://github.com/ClickHouse/ClickHouse/pull/95970), [#94413](https://github.com/ClickHouse/ClickHouse/pull/94413) | `probes/ingest/dedup_retry.go` (identical block retried; OpenMeter-style retry with fresh `store_row_id`) |
| **IN-04** | **Deduplication window by engine** | SharedMergeTree follows replicated semantics ❓ *(h)* | `replicated_deduplication_window=10000` (since 25.9). Plain `MergeTree` has `non_replicated_deduplication_window=0`, so it "deduplicates nothing". *(h)* | [deduplicating-inserts-on-retries](https://clickhouse.com/docs/concepts/features/operations/insert/deduplicating-inserts-on-retries) | `probes/ingest/dedup_window.sql` |
| IN-05 | `ReplacingMergeTree … FINAL` read cost | measured | measured | — | Probe query pair with and without `FINAL` on a populated table, per target |
| **IN-06** | **Time-to-queryable vs `wait_for_async_insert`** | measured | measured | — | `ingest` experiment (ack → visible p50/p99 per replica) |

## 4. Cloud-only capabilities

| ID | Capability | cloud | self-managed | Source | Probe / evidence |
| --- | --- | --- | --- | --- | --- |
| CO-01 | Compute-compute separation (warehouses) | ✅ Scale tier and above. Read-only services "don't execute background merges". *(h)* | ❌ No equivalent. The closest options are dedicated read replicas, or Antalya swarms for object-storage data. *(h)* | [warehouses](https://clickhouse.com/docs/products/cloud/features/infrastructure/warehouses) | assessed, not measured (optional Cloud experiment) |
| CO-02 | Idle scaling and autoscaling | Idling pauses compute billing. Vertical autoscaling on Scale/Enterprise. Horizontal autoscaling is "Private Preview". *(h)* | ❌ Manual; HPA is not meaningful for a stateful CH | [idling](https://clickhouse.com/docs/products/cloud/features/autoscaling/idling), [autoscaling](https://clickhouse.com/docs/products/cloud/features/autoscaling/overview) | Idle-floor COGS cell plus cost curves |
| CO-03 | ClickPipes managed ingest | ✅ *(h)* | ❌ | [clickpipes](https://clickhouse.com/docs/integrations/clickpipes/home) | assessed; OpenMeter's sink does not use it |
| CO-04 | Managed backups | Daily by default with 24 h retention. Scale/Enterprise: 1–45 days retention, every 6–48 h. Restores go to a new service. BYO-bucket export is not available for TDE services. *(h)* | ❌ (self-run: clickhouse-backup or `BACKUP`) | [backups](https://clickhouse.com/docs/products/cloud/features/backups/overview) | `backup-restore` experiment |
| CO-05 | Encryption at rest | Default with provider-managed AES-256 keys. TDE/CMEK on Enterprise, chosen at service creation. *(h)* | Self-provided: EBS encryption, S3 SSE, or encrypted disks *(h)* | [cmek](https://clickhouse.com/docs/products/cloud/guides/security/cmek) | assessed + config evidence |
| CO-06 | TLS in transit | Native 9440 and HTTPS 8443 require TLS. Strict TLS-only is ❓ (one page mentions an unencrypted native connection). *(h)* | Operator-configured *(h)* | [connection details](https://clickhouse.com/docs/get-started/quickstarts/obtain-your-cloud-connection-details) | `probes/security/tls.go` |

## 5. Self-managed-only control

| ID | Capability | cloud | self-managed | Source | Probe / evidence |
| --- | --- | --- | --- | --- | --- |
| SM-01 | Server config and profiles | "most system settings aren't configurable"; no config files *(h)* | Full control *(h)* | [cloud-compatibility](https://clickhouse.com/docs/products/cloud/guides/cloud-compatibility) | `probes/control/server_settings.sql` |
| SM-02 | Statements and features Cloud restricts | Experimental features disabled. `SYSTEM STOP/START MERGES`, `STOP/START FETCHES`, `SHUTDOWN` unsupported. Custom storage policies and tiered-storage TTL unsupported. *(h)* | ✅ *(h)* | [system statements](https://clickhouse.com/docs/reference/statements/system), cloud-compatibility | `probes/control/restricted_statements.sql` (expects an error code on cloud) |
| SM-03 | Executable UDFs | ✅ **Contradicts the starting hypothesis:** "SQL and executable user-defined functions are generally available in ClickHouse Cloud". WASM and driver-based UDFs unsupported. *(h)* | ✅ via server config *(h)* | cloud-compatibility | `probes/control/udf.sql` |
| SM-04 | Altinity-specific builds (Antalya: Iceberg, swarms, Parquet metadata cache, Hybrid engine) | ❌ | Antalya only (`.altinityantalya` tags, latest 26.6.4.20001), **not** in Stable *(h)* | [Antalya.md](https://github.com/Altinity/ClickHouse/blob/antalya-26.6/Antalya.md) | assessed, out of the main comparison |
| SM-05 | `clickhouse-backup` | n/a (v2.8.1 can **restore** Cloud BYO-bucket backups onto self-managed) | ✅ v2.8.1 *(h)* | [ChangeLog](https://github.com/Altinity/clickhouse-backup/blob/master/ChangeLog.md) | `backup-restore` experiment |
| SM-06 | Operator-driven upgrades | Release channels: Fast / Regular / Slow (Enterprise) *(h)* | Change the image in the CHI pod template. The operator reconciles shard by shard; 0.27.4 adds quorum-preserving Keeper rolling updates. *(h)* | [chi_update_clickhouse_version](https://github.com/Altinity/clickhouse-operator/blob/master/docs/chi_update_clickhouse_version.md), [release-0.27.4](https://github.com/Altinity/clickhouse-operator/releases/tag/release-0.27.4), [upgrades](https://clickhouse.com/docs/products/cloud/features/admin-features/upgrades) | `reliability` experiment (rolling upgrade) |
| SM-07 | Support window | Continuous upgrades | Altinity Stable "Supported for three years"; upstream LTS for 1 year *(h)* | [Altinity/ClickHouse README](https://github.com/Altinity/ClickHouse) | assessed |

## 6. Observability of ClickHouse itself (what the COGS runner needs)

| ID | Need | cloud | altinity / upstream-oss | Source | Planned probe |
| --- | --- | --- | --- | --- | --- |
| **OB-01** | `system.query_log` (per replica) | ✅ via `clusterAllReplicas(default, …)` *(h)* | ✅ enabled in shipped config. The cluster name comes from the CHI, not always `default`. *(h)* | [system-tables](https://clickhouse.com/docs/products/cloud/features/monitoring/system-tables) | `probes/observability/system_tables.sql` |
| **OB-02** | `system.part_log` (merge accounting) | ✅ *(h)* | ✅ enabled in shipped `config.xml` 🔬 (table present on both 26.3 images); operator config not yet verified *(h)* | same | same |
| OB-03 | `system.asynchronous_insert_log` | ❓ | ✅ enabled in shipped `config.xml` *(h)* | — | same |
| OB-04 | `CGroupMaxCPU` | ✅ ("Allocated CPU") *(h)* | ✅ with pod CPU limits; may be absent under compose, which falls back to `max_threads` *(h)* | [cloud-console monitoring](https://clickhouse.com/docs/products/cloud/features/monitoring/cloud-console) | `probes/observability/capacity.sql` |
| OB-05 | `SYSTEM FLUSH LOGS ON CLUSTER` | ❓ (the harness already falls back to local flush + settle) | ✅ *(h)* | — | `probes/observability/flush_logs.sql` |
| OB-06 | Merge CPU in `part_log.ProfileEvents` | ❓ SMT merges run on one replica; completeness unknown | ✅ *(h)* | — | COGS runs: compare merge CPU with `system.events` deltas |
| OB-07 | S3 request counters (`S3GetObject`, `S3PutObject`, …) for S3 cost | ✅ in ProfileEvents *(h)* | ✅ when S3 disks are used *(h)* | — | COGS runs on `s3` storage mode |

## 7. Multi-tenancy and security

| ID | Capability | cloud | self-managed | Source | Planned probe |
| --- | --- | --- | --- | --- | --- |
| SE-01 | Row policies by `namespace` | ✅ *(h)* | ✅ *(h)* | [multitenancy](https://clickhouse.com/docs/products/cloud/guides/best-practices/multitenancy) | `probes/security/row_policy.sql` |
| SE-02 | RBAC | ✅ except `GRANT ALL`, `EXECUTE AS`, table-engine grants *(h)* | ✅ *(h)* | cloud-compatibility | `probes/security/rbac.sql` |
| SE-03 | Quotas | ❓ | ✅ *(h)* | — | `probes/security/quota.sql` |
| SE-04 | Workload scheduling (`CREATE WORKLOAD` / `CREATE RESOURCE`, since 24.11; config-based setup removed in 26.7) | ❓ | ✅ *(h)* | PR [#69187](https://github.com/ClickHouse/ClickHouse/pull/69187) | `probes/security/workload.sql` |
| SE-05 | Authentication methods | No LDAP, Kerberos or X.509 *(h)* | All *(h)* | cloud-compatibility | assessed |

## 8. Client compatibility (`clickhouse-go` v2.48.0, OpenMeter's version)

| ID | Behavior | Planned probe |
| --- | --- | --- |
| CL-01 | Native vs HTTP protocol, TLS ports | `probes/client/protocols_test.go` |
| CL-02 | lz4 / zstd compression | same |
| CL-03 | `WithAsync(wait)` semantics vs the server defaults in IN-01/IN-02 | `probes/client/async_test.go` |
| CL-04 | Batch insert of JSON text into a `JSON` column | `probes/client/json_insert_test.go` |

## Starting hypotheses checked against sources so far

- **Refuted:** executable UDFs as "self-managed only" (SM-03). Cloud documents them as GA.
- **Refined:** Cloud uses the **Shared** database engine, not `Replicated` (ST-02).
- **New:** `async_insert=1` and deduplicate-all-inserts became defaults in 26.2 (IN-01, IN-03). Every target in this evaluation is ≥ 26.2, so OpenMeter's inserts are async on all of them unless the sink sets `async_insert=0`.
- **New:** lightweight `UPDATE` is still beta in 26.x (VG-05). Lightweight mutations fail on tables with projections (VG-10).

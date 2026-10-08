# Compute/data separation (Cloud) vs colocation (self-managed): cited facts

Input to design D14 and to feature-matrix section 10. These are **source
facts, not results.**

## Sources

- **Docs:** `ClickHouse/ClickHouse` `docs/` at `cd023af8` (2026-10-08),
  published at `clickhouse.com/docs/<path>`.
- **Code defaults:** checked at master and at tag `v26.3.44.2-lts`.
- **Blog posts:** clickhouse.com was unreachable, so blog wording was seen
  only in search snippets. It is marked *unverified verbatim* and is not used
  as evidence.

## ClickHouse Cloud (separated)

| Topic | Fact | Quote | Source |
| --- | --- | --- | --- |
| Storage | Data lives in object storage. Compute metadata is stateless ("Nothing is persisted on local disk"). | "Storage backed by object store … Significantly lower price point for storing data" | /products/cloud/reference/architecture; …/shared-catalog |
| Local cache | Per-replica SSD filesystem cache, plus page cache metrics in the console. Cache size vs replica size is **not documented**. A Cloud blog says compute nodes "no longer use local disks for caching" (*unverified verbatim*) — **conflicting; probe needed**. | "The data themselves are stored in object storage, and the SSD cache allows us to speed up queries." | …/infrastructure/parallel-replicas; …/monitoring/cloud-console |
| Distributed cache | Cloud-only. Preview since 2025-05; GA unverified. Server switches `enable_read_through_distributed_cache` / `enable_write_through_distributed_cache` (default false). | "We are enlisting preview participants for Distributed cache for AWS and GCP." | Cloud changelog 2025; Settings.cpp |
| Scaling invalidates caches | Scaling events restart the process. Make-before-break adds new replicas, which start cold (inference). | "scaling takes time and invalidates local caches" | …/autoscaling/scaling-recommendations; v25.5 changelog |
| Cloud-only warm-up levers | `cache_populated_by_fetch` (MergeTree, 0), `cache_warmer_threads` (4), `prefer_warmed_unmerged_parts_seconds` (0), `merge_*_bytes_to_prewarm_cache` | "cause all nodes to load new data parts from storage into their filesystem cache without requiring a query" | merge-tree settings |
| Cache on write | `enable_filesystem_cache_on_write_operations` defaults to 0 upstream; "Cloud default value: `1`" | — | session settings |
| Inserts | Every SharedMergeTree insert is written to shared storage. Async-insert flush defaults differ on Cloud: busy timeout 1000 ms vs 200 ms, max data 100 MiB vs 10 MiB. | "all inserts to SharedMergeTree are quorum inserts (written to shared storage)" | …/shared-merge-tree; settings |
| Merges | Read-write services merge; read-only warehouses do not | "Doesn't perform background merge operations outside of system tables" | …/infrastructure/warehouses |
| Scale-out | No data to copy | "add and remove replicas nearly instantaneously" | …/parallel-replicas |
| Billing | Compute, storage (compressed), backups and data transfer. **No object-storage request charges. No intra-region transfer charges.** | "no charges for intra-region data transfer" | …/billing/billing-overview, …/network-data-transfer |

## Self-managed

**Colocated (EBS)**
- Data and compute share a node, and the OS page cache serves warm reads.
- ClickHouse cannot drop the OS page cache. The docs' own benchmarks "flush
  the Linux page cache prior to each cold query run" from outside ClickHouse
  (/concepts/core-concepts/academic-overview).

**S3 disk with a cache disk ("self-separated")**
- **Configuration:** `<type>cache</type>` over `<disk>s3</disk>` with `path`
  and `max_size`.
- **Cache defaults:** `cache_on_write_operations` = false, policy SLRU.
- **Metadata stays local per replica, so replicas are not stateless:**
  "When metadata is stored locally per replica (the default), each replica
  independently manages its pointers to blobs" (/concepts/…/storing-data).
- **Zero-copy** is "not ready for production", so each replica stores its own
  objects.
- **Introspection:** `system.filesystem_cache`, `system.remote_data_paths`,
  `system.blob_storage_log`.
- **ProfileEvents for request accounting:**
  - `S3GetObject`, `S3PutObject`, `S3UploadPart`, `S3ListObjects`,
    `S3HeadObject`, and the `DiskS3*` equivalents;
  - `S3ReadRequestsCount`, `S3WriteRequestsCount`;
  - `CachedReadBufferReadFromCacheBytes`,
    `CachedReadBufferReadFromSourceBytes`.
- **The docs position this mode as for colder data:** "especially useful for
  use cases where query performance on 'cold' data is less critical"
  (/guides/oss/deployment-and-scaling/separation-storage-compute).

## Consequences for the experiment design

- **Fairness: cache on write.** Cloud writes inserted and merged parts into the
  cache by default; self-managed does not. The self-managed `s3` mode is run
  twice, with `cache_on_write_operations` off (the default) and on, so the
  comparison does not hide a configuration difference.
- **Locality probes run first.** These Cloud facts decide how C1, C2 and C4
  are read:
  - whether `system.filesystem_cache` and the `FilesystemCacheSize` metric are
    readable;
  - which distributed-cache settings exist and whether they are enabled;
  - `enable_filesystem_cache_on_write_operations`.
- **Cloud-only warm-up levers are a tuned-Cloud sub-experiment.** For C4, warm-up
  with `cache_populated_by_fetch` and `prewarm_mark_cache` on is reported
  separately from the defaults.
- **Cost lines differ by design.** Self-managed S3 pays per request and Cloud
  does not, so the request count per 1M events (C3) is what makes the
  comparison fair.

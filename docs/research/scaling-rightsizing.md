# On-demand scaling and rightsizing: cited facts

Input to design D15 and to feature-matrix section 11. These are **source
facts, not results.**

## Sources

| Key | Repository | Commit |
| --- | --- | --- |
| **[CH]** | `ClickHouse/ClickHouse` `docs/` (published at `clickhouse.com/docs/<slug>`) | `cd023af8` (2026-10-08) |
| **[API]** | `docs/_specs/cloud-openapi.json` in [CH] | same |
| **[OP]** | `Altinity/clickhouse-operator` tag `release-0.27.4` | `82498653` |
| **[KB]** | `Altinity/altinityknowledgebase` | `bdc41270` |
| **[CSI]** | `kubernetes-sigs/aws-ebs-csi-driver` | `bbf7d7b9` |
| **[BC]** | botocore EC2 API reference | `bfb412fa` |
| AWS prices | AWS Price List API | snapshot 2026-10-07 |

## ClickHouse Cloud

| Topic | Fact | Quote | Source |
| --- | --- | --- | --- |
| Vertical autoscaling: signals | CPU target 53%: up above 75%, down below 37.5%. Memory from peak query memory, RSS and OOMs with 1.1–1.5× headroom. | "Service usage is constantly monitored over a lookback window to make scaling decisions." | /cloud/features/autoscaling/vertical, …/scaling-recommendations |
| Vertical autoscaling: windows | **3 h small window, 30 h large window.** Scale-down latency was cut from up to 30 h to as little as 3 h. | "scaling takes time and invalidates local caches; so it is safer to scale up in a single step." | scaling-recommendations; Cloud changelog |
| Make-before-break | New-size replicas are added before old ones are removed. Old replicas wait up to 1 h for running queries. | "This adds one or more replicas of the new size before removing the old replicas" | /manage/scaling, /cloud/features/mbb |
| Memory bounds | Multiples of 4 GiB, minimum 8, maximum 356 (paid). Setting min = max disables autoscaling. | "Must be a multiple of 4 and lower than or equal to … 356 for paid services." | [API] `replicaScaling` |
| Tiers | Vertical autoscaling on Scale, and on Enterprise with standard 1:4 profiles. Basic is fixed. | "Scale tier services and Enterprise tier services with standard `1:4` profiles can also be autoscaled vertically" | /cloud/features/autoscaling/vertical |
| Horizontal scaling | Manual, 1–50 replicas (50 per warehouse is a soft limit). The scaling service converges on the final count. | "the scaling service will ignore the intermediate states and converge on the final replica count." | /cloud/features/autoscaling/horizontal; [API] |
| Why scale-out is fast | No data or metadata copy to new replicas | "metadata doesn't need to be replicated as your service scales up and down… faster… scale-up operations." | /cloud/reference/shared-merge-tree |
| Horizontal autoscaling | **Private preview.** Recommendations every 10 min, target 60%. Needs a pinned vertical size. The changelog says it currently scales on concurrent query count and needs 26.2+. | "in Private Preview" | /cloud/features/autoscaling/horizontal-autoscaling; changelog |
| Scheduled scaling (beta) | `POST …/scalingSchedule` with entries: `weekdays`, `startHourUtc`/`endHourUtc` (whole UTC hours, no overlap), memory bounds, replicas, idle settings. "the schedule takes precedence at its trigger time." | "Creates or fully replaces the autoscaling schedule" | [API]; /cloud/features/autoscaling/scheduled-scaling |
| Idling | `idleTimeoutMinutes` ≥ 5. Adaptive idling may raise it to 15/30/60 min. Wakes on the first connection. Storage and backups are still billed. | "Wake-up typically takes a few tens of seconds" (Go client docs) / "the first query may take a minute" (CLI docs) | [API]; /integrations/language-clients/go/config-reference; /cloud/features/autoscaling/idling |
| Storage | Metered on compressed bytes; no provisioning. Backups are billed separately; the default is 1 backup kept 1 day. Example: "1 TB of data + 1 backup $50.60" (us-east-1). | "usage is metered on the compressed size of data stored in ClickHouse tables." / "you don't have to size your storage partition upfront" | /cloud/manage/billing/overview; cloud-compatibility |
| Rightsizing guidance | No dedicated Cloud sizing page was found. Generic memory:storage ratio 1:100–1:130, or 1:30–1:50 for frequently accessed data. | "Find a vertical size at which a few of your typical queries run comfortably without memory errors." | horizontal-autoscaling; /guides/sizing-and-hardware-recommendations |

## Self-managed (Altinity operator on EKS)

| Topic | Fact | Quote | Source |
| --- | --- | --- | --- |
| Vertical change | Hosts are reconciled one at a time within a shard: excluded from the cluster, then running queries drain, with a 300 s StatefulSet update timeout and `onFailure: abort`. The first shard is reconciled alone. | "to be excluded from a ClickHouse cluster… to complete all running queries" | [OP] `config/config.yaml`; `worker-reconciler-chi.go` |
| Settings changes | Do not restart without `spec.restart: RollingUpdate` | "will not trigger a server restart" | [OP] `docs/operator_configuration.md` |
| Add replica or shard | Schema is created automatically. A new replica copies all data from its peers. There is no rebalancing. | "the data will be copied from the other replicas to the new one" / "ClickHouse doesn't support automatic shard rebalancing." | [OP] `docs/schema_migration.md`; /engines/…/replication; /guides/sre/scaling-clusters |
| Volume growth | Operator-managed PVs (`storageManagement.provisioner: Operator`, since 0.20.0) allow resize without downtime, given CSI `allowVolumeExpansion` | "Allows to re-scale volumes w/o a downtime" | [OP] changelog; [KB] altinity-kb-kubernetes |
| EBS modifications | **Up to 4 per rolling 24 h.** The old fixed 6-hour wait is outdated (changed January 2026). Volumes cannot shrink. | "You can modify a volume up to four times within a rolling 24-hour period" / "The target volume size must be greater than or equal to the existing size of the volume." | [BC] |
| gp3 limits | 3,000–80,000 IOPS, 125–2,000 MiB/s, 1–65,536 GiB. Baseline 3,000 IOPS and 125 MiB/s included. | — | [BC]; [CSI] |
| Merge headroom | Merges reserve 2× the size of the parts being merged. When space is short, merges stall and part counts grow. Mutations rewrite whole parts or columns. | "Merges book disk space by doubling the total merged parts sizes" | /reference/settings/merge-tree-settings; /sql-reference/statements/alter |
| Keeper sizing | 3 dedicated nodes with fast log disks and ≥4 GB RAM. The operator example limits Keeper to 2 CPU / 4 Gi / 10 Gi. | "USE DEDICATED FAST DISKS for the transaction log!" | [KB] altinity-kb-proper-setup; [OP] chk example |

## Price snapshot (AWS Price List API, 2026-10-07; on-demand Linux)

| Item | us-east-1 | eu-central-1 |
| --- | ---: | ---: |
| m7i.2xlarge ($/h) | 0.4032 | 0.483 |
| m7i.large ($/h) | 0.1008 | 0.12075 |
| m6i.2xlarge ($/h) | 0.384 | 0.46 |
| EKS control plane ($/cluster-h) | 0.10 | 0.10 |
| gp3 storage ($/GB-month) | 0.08 | 0.0952 |
| gp3 IOPS above 3,000 ($/IOPS-month) | 0.005 | 0.006 |
| gp3 throughput above 125 ($/MiBps-month) | 0.04 | 0.0476 |

ClickHouse Cloud compute-unit rates need re-checking before any paid run. The
source repo's profiles use $0.2985 per unit-hour (Scale, as of 2026-06-09),
which is **unverified for today**.

## Consequences for the experiment design

- **E1 (load step).** Vertical autoscaling decides over 3 h and 30 h windows,
  so a minutes-long load step cannot trigger it. E1 is split:
  - **E1a:** scheduled scaling (beta API) gives a deterministic trigger at a
    UTC hour. Measure the transition time, errors and latency, and the billed
    units including make-before-break overlap.
  - **E1b (optional, budget-dependent):** a multi-hour sustained-load test that
    lets vertical autoscaling act on its own.
- **E15 (load-profile model).** Diurnal peaks shorter than the 3 h window are
  followed only by scheduled scaling or idling, not by vertical autoscaling.
  The cost model treats these as separate levers.
- **E11 (headroom).** The 2× reservation for merges and mutations sets the
  EBS provisioning factor. Measured peak usage replaces the rule of thumb in
  cost models.

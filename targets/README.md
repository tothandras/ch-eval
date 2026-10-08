# targets

One directory per platform. Each tier is a manifest (`<tier>.json`). Every
platform has one `up` and one `down` entry point:

```bash
targets/altinity/up local-cluster      # start
targets/altinity/down local-cluster    # stop, remove volumes, verify nothing is left
```

| Platform | Tiers | Notes |
| --- | --- | --- |
| `altinity` | `local-single`; `local-cluster` (1×2 + 3 Keeper + SeaweedFS S3); `local-cluster-s3` (same deployment, S3-backed storage policy); `local-cluster-s3-cow` (S3 policy, `cache_on_write_operations=1`, Cloud's default); `local-cluster-s3-smallcache` (S3 policy, 256 MiB filesystem cache); `local-cluster-2x2`; `kind` (Altinity operator 0.27.4, CHK ×3, CHI 1×2) | Image `altinity/clickhouse-server:26.3.33.10001.altinitystable` |
| `upstream-oss` | same tiers except the two s3 cache variants, which test storage knobs rather than build differences | Image `clickhouse/clickhouse-server:26.3.33.24`, the same 26.3.33 base. The kind tier runs this image under the same operator. |
| `cloud` | `scale` (paid) | `up` refuses until Phase 3 approval and `BUDGET_ACK` |

`targets/collector` runs the pinned OpenTelemetry Collector (contrib 0.162.0)
for the OTel collector ingest path (`bench --ingest-path collector`).

## Using the local tiers

- **Local tiers are for correctness, parity and smoke only.** The harness
  refuses to use them for performance conclusions (ADR-0005).
- **Connection strings:**
  - Each manifest names its DSN environment variable (`dsn_env`).
  - Local tiers fall back to the passwordless `default` user on the tier's
    first native port. No connection string is stored in the repository.
  - Paid tiers have no default and read secrets only from the environment
    (`.envrc`).
- **Ports:**

  | Platform | Native base | HTTP |
  | --- | --- | --- |
  | `altinity` | 19000 (single), 19100 (1×2), 19200 (2×2), 19400 (kind), 19500 (s3-cow), 19600 (s3-smallcache) | native − 877 |
  | `upstream-oss` | 29000 (single), 29100 (1×2), 29200 (2×2), 29400 (kind) | native − 877 |

  Node *i* of a tier listens on `native_base + i`. On kind, `targets/kind/up`
  holds one `kubectl port-forward` per pod (`targets/kind/forward` restarts
  them).

  The two platforms can run side by side.
- **Generated state:** compose files, kind configs, CHI/CHK manifests and the
  kubeconfig go to `targets/.state/` (git-ignored). This includes a
  random per-deployment interserver secret. The generator is
  `targets/compose/gen.py`.

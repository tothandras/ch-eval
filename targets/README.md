# targets

One directory per platform. Each tier is a manifest (`<tier>.json`). Every
platform has one `up` and one `down` entry point:

```bash
targets/altinity/up local-cluster      # start
targets/altinity/down local-cluster    # stop, remove volumes, verify nothing is left
```

| Platform | Tiers | Notes |
| --- | --- | --- |
| `altinity` | `local-single`, `local-cluster` (1×2 + 3 Keeper + SeaweedFS S3), `local-cluster-s3` (same deployment, S3-backed storage policy), `local-cluster-2x2` | Image `altinity/clickhouse-server:26.3.33.10001.altinitystable` |
| `upstream-oss` | same tiers | Image `clickhouse/clickhouse-server:26.3.33.24`, the same 26.3.33 base |
| `cloud` | `scale` (paid) | `up` refuses until Phase 3 approval and `BUDGET_ACK` |

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
  | `altinity` | 19000 (single), 19100 (1×2), 19200 (2×2) | native − 877 |
  | `upstream-oss` | 29000 (single), 29100 (1×2), 29200 (2×2) | native − 877 |

  The two platforms can run side by side.
- **Generated state:** compose files and configs go to
  `targets/.state/<platform>-<deployment>/` (git-ignored). This includes a
  random per-deployment interserver secret. The generator is
  `targets/compose/gen.py`.

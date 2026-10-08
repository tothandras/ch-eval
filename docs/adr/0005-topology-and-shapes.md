# ADR-0005: Local tiers for correctness only; matched paid shapes for performance and cost

- Status: Proposed (Phase 1). Region, replica count and size await owner approval.
- Date: 2026-10-08

## Context

- **What the brief allows:**
  - Performance and COGS conclusions require matched shapes in the same cloud
    region: equal vCPU and RAM per replica, equal replica count, same client
    location.
  - Self-managed must be tested on both EBS gp3 and S3-backed MergeTree.
  - Local runs are for correctness, parity and smoke tests only.
- **The build sandbox:**
  - 4 vCPU / 16 GiB, Docker and kind available.
  - cgroup v1 with no CAP_SYS_RESOURCE, so kind needs `failCgroupV1: false`
    and containerd `restrict_oom_score_adj = true`.
  - `minio/minio` is no longer pullable.

## Decision

**Local tiers:**
- `local-single`: compose, one node. Used by CI.
- `local-cluster`: compose, 2 replicas × 1 shard, 3 Keeper nodes and an S3
  emulator.
- `kind`: Altinity operator 0.27.4, CHI 1×2 and CHK 3.

**Paid tier (proposed; needs approval):**
- **Location:** one AWS region and AZ, plus a client VM in the same AZ.
- **cloud:**
  - Scale-tier service, N replicas × 16 GiB / 4 vCPU.
  - Idle scaling off, except in the idle-floor cell.
  - Release channel recorded.
- **altinity and upstream-oss:**
  - EKS with the Altinity operator, identical manifests apart from the image.
  - One ClickHouse pod per dedicated node, Guaranteed QoS at 4 vCPU / 16 GiB.
  - Keeper: 3 pods on small dedicated nodes.
- **Storage modes:**
  - `ebs`: gp3 with recorded IOPS and throughput.
  - `s3`: s3 disk + filesystem cache, no zero-copy, so each replica stores its
    own objects.
- **Optional `upstream-cloudver` control.** Same self-managed topology, running
  the upstream image matching Cloud's `version()`. It separates version skew
  from architecture.

## Consequences

- Cloud's hardware generation is not observable. Self-managed uses a current
  general-purpose instance family, and the CPU model is recorded as a caveat.
- Every topology is a single shard. Sharding risks are assessed, not measured.

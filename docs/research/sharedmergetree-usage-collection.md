<!--
Provenance: provided verbatim by the repository owner on 2026-10-08 as an input
to this evaluation. Only change: relative OpenMeter source links were rewritten
to absolute links pinned at openmeterio/openmeter@5a0e27e6 (all ten targets
verified to exist at that commit). The claims below are HYPOTHESES for this
repo: the `replica-consistency` workload measures them on every target. Do not
cite this file as evidence; cite results/ instead.
-->

# Why we need SharedMergeTree for reliable usage collection

## When is usage data available?

When a customer asks when submitted usage becomes available, our guidance should be:

> Usage is typically queryable within about 10 seconds of submission. Ingestion is asynchronous, so poll for the expected data with backoff and a timeout rather than assuming it is available after a fixed delay. Ten seconds describes typical latency, not a guaranteed deadline.

The 10-second figure is our stated typical behavior, not a ClickHouse guarantee or a measurement established by this analysis. Observing an event confirms that event is visible to that read. An unchanged aggregate or several identical responses does not prove that every earlier event has arrived. Polling a total is only a completion check when the caller knows what result to expect.

For billing, we need a stronger property: **an authoritative usage snapshot must include all eligible data through its cutoff, and changing the server answering the query must not make previously visible usage disappear.** ReplicatedMergeTree makes this harder because each replica has its own asynchronously updated copy of the data.

## The problems with ReplicatedMergeTree

ClickHouse normally acknowledges an insert after the receiving replica writes it. Other replicas obtain the data asynchronously, with no fixed upper bound on the delay. A successful read from another replica can return an incomplete result without an error. [ClickHouse replication documentation](https://clickhouse.com/docs/reference/engines/table-engines/mergetree-family/replication).

| Problem | Observable effect | Consequence for OpenMeter |
|---|---|---|
| A query reaches a lagging replica | An accepted event is absent from the result. | Polling and collection can report less usage than actually exists. |
| Consecutive reads reach different replicas | Cumulative positive usage appears to decrease. | Customers see inconsistent totals; collection can fail on a negative calculated delta. |
| Replication exceeds the collection grace period | A successful snapshot omits eligible events. | Billing persists an incomplete result. |
| Invoice collection and a meter query see different data | The displayed usage does not reconcile with the invoice. | Customers cannot verify their bill; support must reconstruct the collection snapshot. |
| The final snapshot is incomplete | The replica catches up after the charge completes. | Missing usage is not automatically recollected. |

For example, one collection sees 100 units; the next reads a stale replica and sees 90, although actual cumulative usage is 120. At a simple positive unit price, subtracting previously billed usage produces a negative total. The current run-creation path rejects it with `ErrChargeTotalIsNegative`. If the stale read returns 105 instead, collection succeeds but undercollects. See [run creation](https://github.com/openmeterio/openmeter/blob/5a0e27e6f50b301e9bdb93772a6a970fd4a1f58f/openmeter/billing/charges/usagebased/service/run/create.go) (`CreateRatedRun`).

Missing data can also cause overcharging: a missing negative event increases net usage, while missing usage can prevent a customer from reaching a cheaper volume tier. Average and latest-value meters can move in either direction. The consequence depends on the meter and price.

## Sharding adds another completeness requirement

Replication copies a shard; it does not make every server contain the entire dataset. A distributed read normally selects a replica from each shard. Background writes through a `Distributed` table can add a delivery queue before replication begins. [ClickHouse Distributed engine](https://clickhouse.com/docs/reference/engines/table-engines/special/distributed).

Billing consequently depends on correct shard coverage and sufficiently fresh reads from every relevant shard. A local-table query can omit other shards entirely. A distributed query can combine fresh and stale results. Missing shards must cause a failure rather than a partial billable total. Routing changes and data migration add opportunities to omit historical events or count copied events twice.

These are topology-dependent risks, not evidence that the deployed cluster currently has a routing error. Waiting longer cannot repair incorrect shard coverage.

## How charges snapshot usage today

A charge realization persists a cumulative usage checkpoint. Its query covers:

```text
charge service-period start <= event time < ServicePeriodTo
stored_at < StoredAtLT
```

The collection schedule provides an additional one-minute wait:

| Boundary | Mid-period invoice collection | Final collection |
|---|---|---|
| `ServicePeriodTo` | The current collection’s service-period end | The charge’s service-period end |
| `StoredAtLT` | The current collection’s service-period end | The charge’s service-period end plus the configured collection interval |
| Scheduled snapshot readiness | `StoredAtLT + 1 minute` | `StoredAtLT + 1 minute` |

See [run boundaries](https://github.com/openmeterio/openmeter/blob/5a0e27e6f50b301e9bdb93772a6a970fd4a1f58f/openmeter/billing/charges/usagebased/service/creditheninvoice.go) (`StartInvoiceRun`), [snapshot scheduling](https://github.com/openmeterio/openmeter/blob/5a0e27e6f50b301e9bdb93772a6a970fd4a1f58f/openmeter/billing/charges/usagebased/service/statemachine.go) (`getCurrentRunSnapshotAfter`), and [quantity query](https://github.com/openmeterio/openmeter/blob/5a0e27e6f50b301e9bdb93772a6a970fd4a1f58f/openmeter/billing/charges/usagebased/service/rating/quantitysnapshot.go) (`snapshotQuantity`).

**The one-minute wait is a timing assumption, not a completeness check.** The sink assigns `stored_at` before its storage operation completes. Storage retries, downstream buffering, shard delivery, and replica catch-up can consume or exceed the grace period. See [sink flushing](https://github.com/openmeterio/openmeter/blob/5a0e27e6f50b301e9bdb93772a6a970fd4a1f58f/openmeter/sink/sink.go) (`flush`).

An illustrative failure:

| Time | Event |
|---|---|
| 11:59:50 | The sink assigns `stored_at` to a batch containing 20 units. |
| 12:00:00 | Collection cutoff; actual cumulative usage is 100 units. |
| 12:01:00 | Billing reads a replica that sees 80 units and snapshots that result. |
| 12:02:00 | The remaining 20 units become visible. |

Mid-period collection has some recovery: the next run re-queries cumulative usage and subtracts already booked detailed lines. If it sees 130 units after billing 80, a simple unit price produces a 50-unit delta: 30 new units plus the missing 20. That catch-up is attributed to the later run’s period. See [delta rating](https://github.com/openmeterio/openmeter/blob/5a0e27e6f50b301e9bdb93772a6a970fd4a1f58f/openmeter/billing/charges/usagebased/service/rating/delta/README.md).

Final collection has no equivalent automatic next run. Once the charge completes, replication catching up does not reopen it. A fixed `StoredAtLT` also does not freeze the database: repeating the same query later can return more eligible rows. PostgreSQL transactions and charge locks protect billing state, but cannot make the ClickHouse input complete.

## Why the invoice and meter query can disagree

In the example above, an issued invoice based on the 80-unit snapshot can remain at 80 while a later meter query shows 100 for the same event-time period. The invoice preserves the result used for billing; the live query sees the newly visible rows. The discrepancy can persist after every replica has caught up.

The reverse is also possible: billing reads a fresh replica and records 100 units, then the customer queries a stale replica and sees 80. The invoice appears to charge for usage the customer cannot see. Both requests can succeed, and retries can change the displayed total.

This creates billing disputes, weakens trust in usage reporting, and makes reconciliation difficult. A later catch-up charge may also look unexplained: it includes usage missing from a previous snapshot rather than only new usage in the displayed period.

**Compare equivalent quantities before attributing a discrepancy to replication.** Match the customer, meter, group filters, event-time bounds, and ingestion cutoff. Account for prior runs: the stored run quantity is cumulative, while an invoice line can represent a delta. Unit conversion, discounts, commitments, and price tiers also mean an invoice amount is not directly comparable to a raw meter value. See [charge quantity semantics](https://github.com/openmeterio/openmeter/blob/5a0e27e6f50b301e9bdb93772a6a970fd4a1f58f/openmeter/billing/charges/README.md).

Even after matching query boundaries, replication can still cause disagreement because the cutoff filters the rows currently visible; it does not identify an immutable snapshot. Reliable cross-replica visibility is therefore necessary for invoice reconciliation. Exact historical reproduction additionally depends on preserving the billing inputs and their cutoff semantics.

## Changed meter values also affect enforcement

Usage is an input to access decisions, not just reporting and invoices. For hard-limit metered entitlements, the current implementation allows access only while the calculated balance is positive; soft limits continue to allow access. See [access calculation](https://github.com/openmeterio/openmeter/blob/5a0e27e6f50b301e9bdb93772a6a970fd4a1f58f/openmeter/entitlement/metered/connector.go) (`MeteredEntitlementValue.HasAccess`) and [entitlement semantics](https://github.com/openmeterio/openmeter/blob/5a0e27e6f50b301e9bdb93772a6a970fd4a1f58f/openmeter/entitlement/README.md).

Consider a customer with a 100-unit grant, no resets or grant changes, and 100 units of actual consumption. Assume the relevant usage is within the evaluated time range:

| Read | Visible consumption | Calculated remaining balance | Hard-limit decision |
|---|---|---|---|
| Lagging replica | 90 units | 10 units | Allow: the customer can exceed the intended limit. |
| Fresh replica | 100 units | 0 units | Deny: the limit has been reached. |
| Lagging replica on a subsequent check | 90 units | 10 units | Allow again, despite no replenishment or reset. |

This illustrates the risk when successive balance calculations consume different visible usage for the same range. Access can flip between allowed and denied based on replica selection. A dashboard, enforcement check, and invoice may each show a different view of the customer's usage.

The main consequences are:

- **Delayed enforcement:** missing consumption makes remaining balance look larger, allowing additional service after exhaustion. Later billing cannot undo service already delivered.
- **Premature denial:** where an enforcement calculation supports downward usage corrections or decreasing meter values, missing that reduction can make usage appear too high and block a customer who should have access. Missing positive consumption alone has the opposite effect.
- **Unstable downstream behavior:** applications acting on changing values may repeatedly block and unblock access, or emit inconsistent threshold notifications. Caching an access result can extend the impact of a stale read.

The charge collection grace period does not protect these decisions: enforcement needs an answer when access is checked. Metered entitlements also have a separate **one-minute temporal resolution**, so typical 10-second ingestion visibility is not a promise of sub-minute enforcement. See [credit time semantics](https://github.com/openmeterio/openmeter/blob/5a0e27e6f50b301e9bdb93772a6a970fd4a1f58f/openmeter/credit/README.md).

SharedMergeTree strengthens the case for a consistent read policy on enforcement queries as well as billing queries. It does not make asynchronous metering an atomic quota reservation: concurrent requests can consume the same apparent remaining balance before their usage is recorded. Strict prevention of overshoot would require a separate admission/reservation mechanism; it is not a property supplied by changing the table engine. When freshness cannot be established, the enforcement integration also needs an explicit policy for allowing or denying access.

## Why SharedMergeTree is the required direction

SharedMergeTree stores table data in shared object storage and coordinates metadata through ClickHouse Keeper. Compute replicas access the same data rather than copying table parts between themselves. It supports scaling compute without manually sharding the table, removing the corresponding routing and shard-coverage burden. [SharedMergeTree architecture](https://clickhouse.com/docs/products/cloud/features/infrastructure/shared-merge-tree).

**Metadata visibility remains asynchronous.** For billing reads that must avoid stale results across replicas, SharedMergeTree supports `select_sequential_consistency`; alternatively, synchronize the serving replica with `SYSTEM SYNC REPLICA LIGHTWEIGHT`. These operations add coordination overhead and do not prove that upstream ingestion has finished. [SharedMergeTree consistency](https://clickhouse.com/docs/products/cloud/features/infrastructure/shared-merge-tree#consistency).

Our recommendation is therefore **SharedMergeTree with a deliberate consistency policy for authoritative billing, reconciliation, and enforcement reads**. For the 100-unit example, the read must see the completed write’s metadata before rating or deciding access. The engine change alone cannot guarantee that every submitted event has reached storage or that a live query uses the invoice’s boundaries.

ReplicatedMergeTree can be hardened with stricter write/read coordination and complete shard coverage, but those responsibilities remain ours. Increasing grace periods only reduces the frequency of failures. SharedMergeTree is the preferred architecture because it removes major sources of routine inconsistency instead of making billing depend on their typical delay.

## Late-event support is separate work

There are two distinct cases:

- **Eligible but not yet visible:** an event has `stored_at < StoredAtLT`, but the snapshot cannot see it. Storage visibility and snapshot readiness must address this case.
- **Arrived after the cutoff:** an event belongs to the service period but reaches the sink after `StoredAtLT`. Repeating the same bounded query will still exclude it; retrospective billing needs a later ingestion cutoff.

**Planned late-event support:** During the next billing cycle—for example, one month later—we would check whether additional events have appeared for an earlier service period and bill them retrospectively. The current developer estimate is **2–3 weeks of work**. This was intended for customers submitting data late or recovery after an outage, rather than as a dependency of normal daily operations.

That work remains necessary with SharedMergeTree. It addresses genuinely late data and recovery after final collection. Routine billing should not need to wait until the next month to correct data that was already written but invisible to a selected replica.

## Acceptance criteria

- Customer guidance distinguishes typical 10-second availability from a guarantee and describes polling with backoff and a timeout.
- Authoritative billing, reconciliation, and enforcement reads use an explicit consistency policy, verified across server changes.
- Enforcement validation covers exhausted balances appearing replenished after replica switching, delayed blocking, and supported downward corrections. The integration defines behavior when freshness cannot be established and does not assume atomic quota enforcement.
- Invoice-versus-meter comparisons use equivalent filters and cutoffs, distinguish cumulative usage from billed deltas, and explain later corrections.
- Snapshot readiness accounts for in-flight ingestion through the cutoff; elapsed grace alone is not proof of completeness.
- Validation covers delayed visibility, fresh-to-stale reads, invoice/meter discrepancies in both directions, and events appearing after final collection.
- Retrospective late-event billing remains a separately planned capability with the scope and estimate above.

This analysis is grounded in the inspected implementation and ClickHouse documentation. It does not verify deployed table engines or settings; deployment verification is needed before claiming these properties in production.

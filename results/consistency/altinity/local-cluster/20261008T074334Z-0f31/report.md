# Replica consistency: altinity/local-cluster (`26.3.33.10001.altinitystable`), run `20261008T074334Z-0f31`

> Local tier: this run checks that the experiment works and shows the shape of the behaviour; it is not evidence about production latencies.

| Policy | Insert | Restart | Batches | Insert p50/p99 ms | Insert errors | Lag p50/p99/max ms | Never visible | Reads | Stale | Stale rate | Violations /10k | Deficit p99 rows / ms | Read p50/p99 ms | Read errors |
| --- | --- | --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| default | sync | true | 199 | 13.7 / 64.4 | 0 | 39 / 2674 / 2922 | 0 | 2621 | 135 | 5.15% | 49.6 | 1400 / 2686 | 6.2 / 46.0 | 98 |
| default | async | true | 199 | 66.8 / 136.7 | 0 | 36 / 4221 / 4595 | 0 | 2553 | 142 | 5.56% | 50.9 | 2300 / 4487 | 6.5 / 34.4 | 139 |

Lag is measured from insert acknowledgement to the batch being complete on a replica (per-replica polls through `clusterAllReplicas`). A read is stale when its total is below the total acknowledged before it started; a violation is a total lower than the same reader's previous read.

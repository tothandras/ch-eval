# Replica consistency: altinity/kind (`26.3.33.10001.altinitystable`), run `20261008T074513Z-5128`

> Local tier: this run checks that the experiment works and shows the shape of the behaviour; it is not evidence about production latencies.

| Policy | Insert | Restart | Batches | Insert p50/p99 ms | Insert errors | Lag p50/p99/max ms | Never visible | Reads | Stale | Stale rate | Violations /10k | Deficit p99 rows / ms | Read p50/p99 ms | Read errors |
| --- | --- | --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| default | sync | false | 99 | 19.7 / 88.8 | 0 | 32 / 107 / 110 | 0 | 1316 | 63 | 4.79% | 0.0 | 100 / 56 | 8.5 / 34.8 | 0 |
| quorum | sync | false | 90 | 112.4 / 1601.1 | 1 | 6 / 185 / 185 | 0 | 848 | 0 | 0.00% | 0.0 | 0 / 0 | 28.0 / 390.3 | 1 |
| sync-replica | sync | false | 99 | 25.0 / 103.0 | 0 | 40 / 148 / 151 | 0 | 1048 | 0 | 0.00% | 0.0 | 0 / 0 | 17.9 / 116.5 | 2 |

Lag is measured from insert acknowledgement to the batch being complete on a replica (per-replica polls through `clusterAllReplicas`). A read is stale when its total is below the total acknowledged before it started; a violation is a total lower than the same reader's previous read.

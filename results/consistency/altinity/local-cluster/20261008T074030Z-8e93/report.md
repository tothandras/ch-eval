# Replica consistency: altinity/local-cluster (`26.3.33.10001.altinitystable`), run `20261008T074030Z-8e93`

> Local tier: this run checks that the experiment works and shows the shape of the behaviour; it is not evidence about production latencies.

| Policy | Insert | Restart | Batches | Insert p50/p99 ms | Insert errors | Lag p50/p99/max ms | Never visible | Reads | Stale | Stale rate | Violations /10k | Deficit p99 rows / ms | Read p50/p99 ms | Read errors |
| --- | --- | --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| default | sync | false | 149 | 9.6 / 28.6 | 0 | 23 / 100 / 147 | 0 | 2108 | 80 | 3.80% | 4.7 | 100 / 75 | 5.3 / 23.8 | 0 |
| default | async | false | 149 | 65.3 / 81.9 | 0 | 38 / 81 / 97 | 0 | 2073 | 89 | 4.29% | 0.0 | 100 / 37 | 6.0 / 23.1 | 0 |
| quorum | sync | false | 149 | 43.9 / 114.0 | 0 | 12 / 53 / 57 | 0 | 1915 | 0 | 0.00% | 0.0 | 0 / 0 | 9.6 / 43.7 | 0 |
| quorum | async | false | 0 | 0.0 / 0.0 | 149 | 0 / 0 / 0 | 0 | 2082 | 0 | 0.00% | 0.0 | 0 / 0 | 5.8 / 26.5 | 1 |
| sync-replica | sync | false | 149 | 12.7 / 54.5 | 0 | 35 / 103 / 230 | 0 | 1893 | 0 | 0.00% | 0.0 | 0 / 0 | 9.6 / 55.5 | 0 |
| sync-replica | async | false | 149 | 65.7 / 90.0 | 0 | 38 / 93 / 107 | 0 | 1916 | 1 | 0.05% | 0.0 | 100 / 1 | 9.1 / 47.1 | 0 |

Lag is measured from insert acknowledgement to the batch being complete on a replica (per-replica polls through `clusterAllReplicas`). A read is stale when its total is below the total acknowledged before it started; a violation is a total lower than the same reader's previous read.

# Replica consistency: upstream-oss/kind (`26.3.33.24`), run `20261008T081836Z-7386`

> Local tier: this run checks that the experiment works and shows the shape of the behaviour; it is not evidence about production latencies.

| Policy | Insert | Restart | Batches | Insert p50/p99 ms | Insert errors | Lag p50/p99/max ms | Never visible | Reads | Stale | Stale rate | Violations /10k | Deficit p99 rows / ms | Read p50/p99 ms | Read errors |
| --- | --- | --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| default | sync | false | 99 | 17.6 / 141.9 | 0 | 37 / 100 / 378 | 0 | 1333 | 72 | 5.40% | 45.0 | 100 / 312 | 7.5 / 31.6 | 0 |
| quorum | sync | false | 97 | 54.6 / 519.7 | 0 | 19 / 50 / 51 | 0 | 1174 | 0 | 0.00% | 0.0 | 0 / 0 | 11.5 / 144.0 | 0 |
| sync-replica | sync | false | 99 | 18.6 / 83.0 | 0 | 39 / 126 / 129 | 0 | 1103 | 0 | 0.00% | 0.0 | 0 / 0 | 17.7 / 78.4 | 0 |

Lag is measured from insert acknowledgement to the batch being complete on a replica (per-replica polls through `clusterAllReplicas`). A read is stale when its total is below the total acknowledged before it started; a violation is a total lower than the same reader's previous read.

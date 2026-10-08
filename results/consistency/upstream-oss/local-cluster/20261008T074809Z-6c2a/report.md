# Replica consistency: upstream-oss/local-cluster (`26.3.33.24`), run `20261008T074809Z-6c2a`

> Local tier: this run checks that the experiment works and shows the shape of the behaviour; it is not evidence about production latencies.

| Policy | Insert | Restart | Batches | Insert p50/p99 ms | Insert errors | Lag p50/p99/max ms | Never visible | Reads | Stale | Stale rate | Violations /10k | Deficit p99 rows / ms | Read p50/p99 ms | Read errors |
| --- | --- | --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| default | sync | false | 149 | 11.8 / 56.5 | 0 | 32 / 103 / 197 | 0 | 2049 | 125 | 6.10% | 19.5 | 100 / 196 | 6.0 / 32.9 | 0 |
| default | async | false | 149 | 68.3 / 95.4 | 0 | 34 / 93 / 218 | 0 | 2054 | 86 | 4.19% | 9.7 | 200 / 191 | 6.1 / 28.7 | 0 |
| quorum | sync | false | 149 | 45.2 / 89.8 | 0 | 8 / 52 / 53 | 0 | 1968 | 0 | 0.00% | 0.0 | 0 / 0 | 8.8 / 33.6 | 0 |
| quorum | async | false | 0 | 0.0 / 0.0 | 149 | 0 / 0 / 0 | 0 | 2079 | 0 | 0.00% | 0.0 | 0 / 0 | 6.1 / 24.6 | 0 |
| sync-replica | sync | false | 149 | 16.4 / 48.6 | 0 | 41 / 97 / 105 | 0 | 1871 | 0 | 0.00% | 0.0 | 0 / 0 | 10.5 / 43.2 | 0 |
| sync-replica | async | false | 149 | 71.8 / 118.8 | 0 | 41 / 131 / 185 | 0 | 1816 | 0 | 0.00% | 0.0 | 0 / 0 | 10.9 / 71.7 | 0 |

Lag is measured from insert acknowledgement to the batch being complete on a replica (per-replica polls through `clusterAllReplicas`). A read is stale when its total is below the total acknowledged before it started; a violation is a total lower than the same reader's previous read.

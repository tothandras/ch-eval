# E6: add a replica through the operator (replicasCount 2 -> 3)

Target `altinity/kind` (`26.3.33.10001.altinitystable`), operator 0.27.4, 1000000 rows of (UInt64, String). Local tier: mechanics only, timings are not production evidence.

| Milestone after the CHI patch | ms |
| --- | ---: |
| third host listed in the `ch-eval` cluster | 82488 |
| table present on all three replicas (created by the operator, no DDL issued) | 82644 |
| all 1000000 rows fetched on the new replica | 82777 |

CHI status at the end: Completed.

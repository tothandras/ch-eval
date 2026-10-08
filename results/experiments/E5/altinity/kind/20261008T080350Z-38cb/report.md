# E5: vertical change through the operator (pod memory limit)

Target `altinity/kind` (`26.3.33.10001.altinitystable`), 3 replicas. Local tier: mechanics only.

- CHI patch to `Completed`: 140043 ms (the operator rolls one host at a time).
- Pod start times after the roll:

```
chi-cheval-ch-eval-0-0-0 started 2026-10-08T08:05:36Z
chi-cheval-ch-eval-0-1-0 started 2026-10-08T08:06:23Z
chi-cheval-ch-eval-0-2-0 started 2026-10-08T08:07:00Z
```

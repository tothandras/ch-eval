> Target `upstream-oss/kind` is a local tier: this run backs parity and correctness only, never performance.

# Benchmark: clickstack

| Field | Value |
| --- | --- |
| Harness commit | `effdf02` |
| ClickHouse | 26.3.33.24 (multi-node) |
| Clusters | all-clusters, all-replicated, all-sharded, ch-eval, default |
| Sweep id | `20261008T082427Z` |
| Concurrency | 1 |
| Started | 2026-10-08T08:24:27Z |
| Finished | 2026-10-08T08:25:03Z |

## Ingest

Source `otel-direct`: 500000 rows in 30.96s (16150 events/sec), batch=10000.

## Queries

| Query | c | cache | p50 ms | p95 ms | p99 ms | QPS | CPU p50 ms | CPU src | Mem MB |
| --- | ---: | --- | ---: | ---: | ---: | ---: | ---: | --- | ---: |
| body_contains_24h | 1 | warm | 28.0 | 30.0 | 30.0 | 13.5 | 40.9 | os_cpu_cluster | 12.1 |
| logs_for_trace | 1 | warm | 10.0 | 13.0 | 13.0 | 18.7 | 11.2 | os_cpu_cluster | 0.1 |
| needle_search | 1 | warm | 7.0 | 8.0 | 8.0 | 20.4 | 6.2 | os_cpu_cluster | 0.0 |
| rollup_services_24h | 1 | warm | 11.0 | 16.0 | 16.0 | 16.7 | 12.9 | os_cpu_cluster | 0.1 |
| service_tail_1h | 1 | warm | 15.0 | 37.0 | 37.0 | 15.3 | 17.2 | os_cpu_cluster | 0.1 |
| severity_by_service_24h | 1 | warm | 27.0 | 28.0 | 28.0 | 14.1 | 29.3 | os_cpu_cluster | 7.8 |
| top_error_patterns_24h | 1 | warm | 22.0 | 30.0 | 30.0 | 14.0 | 32.3 | os_cpu_cluster | 0.1 |
| user_logs | 1 | warm | 19.0 | 19.0 | 19.0 | 15.2 | 24.8 | os_cpu_cluster | 0.1 |
| user_logs_items | 1 | warm | 16.0 | 18.0 | 18.0 | 16.5 | 21.2 | os_cpu_cluster | 0.1 |
| volume_by_hour | 1 | warm | 26.0 | 34.0 | 34.0 | 13.4 | 47.5 | os_cpu_cluster | 17.2 |

<details>
<summary><code>body_contains_24h</code></summary>

```sql
SELECT count() AS logs, uniqExact(ServiceName) AS services
FROM otel_logs
WHERE Timestamp >= toDateTime({from_24h:UInt32}) AND Timestamp < toDateTime({to:UInt32}) AND Body ILIKE '%timeout%'
```

</details>

<details>
<summary><code>logs_for_trace</code></summary>

```sql
SELECT Timestamp, ServiceName, SpanId, SeverityText, Body
FROM otel_logs
WHERE TraceId = {trace_id:String}
ORDER BY Timestamp, Body
```

</details>

<details>
<summary><code>needle_search</code></summary>

```sql
SELECT Timestamp, ServiceName, SeverityText, Body
FROM otel_logs
WHERE hasToken(lower(Body), lower({needle:String}))
ORDER BY Timestamp, Body
LIMIT 100
```

</details>

<details>
<summary><code>rollup_services_24h</code></summary>

```sql
SELECT Value, sum(count) AS logs
FROM otel_logs_kv_rollup_15m
WHERE Key = 'ServiceName' AND Timestamp >= toDateTime({from_24h:UInt32}) AND Timestamp < toDateTime({to:UInt32})
GROUP BY Value
ORDER BY logs DESC, Value
LIMIT 20
```

</details>

<details>
<summary><code>service_tail_1h</code></summary>

```sql
SELECT Timestamp, SeverityText, Body
FROM otel_logs
WHERE ServiceName = {service:String} AND Timestamp >= toDateTime({from_1h:UInt32}) AND Timestamp < toDateTime({to:UInt32})
ORDER BY Timestamp DESC, Body
LIMIT 100
```

</details>

<details>
<summary><code>severity_by_service_24h</code></summary>

```sql
SELECT ServiceName, SeverityText, count() AS logs
FROM otel_logs
WHERE Timestamp >= toDateTime({from_24h:UInt32}) AND Timestamp < toDateTime({to:UInt32})
GROUP BY ServiceName, SeverityText
ORDER BY ServiceName, SeverityText
```

</details>

<details>
<summary><code>top_error_patterns_24h</code></summary>

```sql
SELECT replaceRegexpAll(Body, '[0-9]+', 'N') AS pattern, count() AS logs
FROM otel_logs
WHERE Timestamp >= toDateTime({from_24h:UInt32}) AND Timestamp < toDateTime({to:UInt32}) AND SeverityNumber >= 17
GROUP BY pattern
ORDER BY logs DESC, pattern
LIMIT 10
```

</details>

<details>
<summary><code>user_logs</code></summary>

```sql
SELECT ServiceName, count() AS logs
FROM otel_logs
WHERE LogAttributes['user.id'] = {user_id:String}
GROUP BY ServiceName
ORDER BY ServiceName
```

</details>

<details>
<summary><code>user_logs_items</code></summary>

```sql
SELECT ServiceName, count() AS logs
FROM otel_logs
WHERE has(LogAttributeItems, concat('user.id=', {user_id:String}))
GROUP BY ServiceName
ORDER BY ServiceName
```

</details>

<details>
<summary><code>volume_by_hour</code></summary>

```sql
SELECT toStartOfHour(Timestamp) AS hour, SeverityText, count() AS logs
FROM otel_logs
WHERE Timestamp >= toDateTime({from:UInt32}) AND Timestamp < toDateTime({to:UInt32})
GROUP BY hour, SeverityText
ORDER BY hour, SeverityText
```

</details>

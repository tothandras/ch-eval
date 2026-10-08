> Target `altinity/local-cluster-2x2` is a local tier: this run backs parity and correctness only, never performance.

# Benchmark: clickstack

| Field | Value |
| --- | --- |
| Harness commit | `3701df2` |
| ClickHouse | 26.3.33.10001.altinitystable (multi-node) |
| Clusters | ch_eval |
| Sweep id | `20261008T075628Z` |
| Concurrency | 1 |
| Started | 2026-10-08T07:56:28Z |
| Finished | 2026-10-08T07:57:10Z |

## Ingest

Source `otel-direct`: 500000 rows in 35.59s (14050 events/sec), batch=10000.

## Queries

| Query | c | cache | p50 ms | p95 ms | p99 ms | QPS | CPU p50 ms | CPU src | Mem MB |
| --- | ---: | --- | ---: | ---: | ---: | ---: | ---: | --- | ---: |
| body_contains_24h | 1 | warm | 61.0 | 75.0 | 75.0 | 9.1 | 52.4 | os_cpu_cluster | 12.3 |
| logs_for_trace | 1 | warm | 26.0 | 33.0 | 33.0 | 15.3 | 22.9 | os_cpu_cluster | 0.9 |
| needle_search | 1 | warm | 15.0 | 22.0 | 22.0 | 18.4 | 16.0 | os_cpu_cluster | 0.6 |
| rollup_services_24h | 1 | warm | 21.0 | 23.0 | 23.0 | 15.9 | 23.9 | os_cpu_cluster | 0.6 |
| service_tail_1h | 1 | warm | 27.0 | 30.0 | 30.0 | 15.4 | 32.8 | os_cpu_cluster | 0.7 |
| severity_by_service_24h | 1 | warm | 43.0 | 45.0 | 45.0 | 11.9 | 46.8 | os_cpu_cluster | 10.8 |
| top_error_patterns_24h | 1 | warm | 45.0 | 48.0 | 48.0 | 11.6 | 59.5 | os_cpu_cluster | 0.9 |
| user_logs | 1 | warm | 36.0 | 61.0 | 61.0 | 11.3 | 36.6 | os_cpu_cluster | 6.7 |
| user_logs_items | 1 | warm | 33.0 | 34.0 | 34.0 | 13.3 | 36.2 | os_cpu_cluster | 0.7 |
| volume_by_hour | 1 | warm | 48.0 | 60.0 | 60.0 | 11.1 | 75.1 | os_cpu_cluster | 23.2 |

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

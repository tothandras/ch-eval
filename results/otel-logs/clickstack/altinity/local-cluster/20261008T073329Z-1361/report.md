> Target `altinity/local-cluster` is a local tier: this run backs parity and correctness only, never performance.

# Benchmark: clickstack

| Field | Value |
| --- | --- |
| Harness commit | `d75d88e` |
| ClickHouse | 26.3.33.10001.altinitystable (multi-node) |
| Clusters | ch_eval |
| Sweep id | `20261008T073329Z` |
| Concurrency | 1 |
| Started | 2026-10-08T07:33:29Z |
| Finished | 2026-10-08T07:34:12Z |

## Ingest

Source `otel-direct`: 500000 rows in 37.57s (13310 events/sec), batch=10000.

## Queries

| Query | c | cache | p50 ms | p95 ms | p99 ms | QPS | CPU p50 ms | CPU src | Mem MB |
| --- | ---: | --- | ---: | ---: | ---: | ---: | ---: | --- | ---: |
| body_contains_24h | 1 | warm | 17.0 | 19.0 | 19.0 | 16.6 | 29.1 | os_cpu_cluster | 12.1 |
| logs_for_trace | 1 | warm | 9.0 | 10.0 | 10.0 | 20.0 | 10.7 | os_cpu_cluster | 0.1 |
| needle_search | 1 | warm | 6.0 | 9.0 | 9.0 | 21.3 | 4.7 | os_cpu_cluster | 0.0 |
| rollup_services_24h | 1 | warm | 8.0 | 13.0 | 13.0 | 19.5 | 10.0 | os_cpu_cluster | 0.1 |
| service_tail_1h | 1 | warm | 11.0 | 14.0 | 14.0 | 20.3 | 17.6 | os_cpu_cluster | 0.1 |
| severity_by_service_24h | 1 | warm | 14.0 | 20.0 | 20.0 | 17.5 | 23.1 | os_cpu_cluster | 10.6 |
| top_error_patterns_24h | 1 | warm | 22.0 | 25.0 | 25.0 | 15.5 | 31.4 | os_cpu_cluster | 0.2 |
| user_logs | 1 | warm | 11.0 | 12.0 | 12.0 | 19.4 | 16.4 | os_cpu_cluster | 0.1 |
| user_logs_items | 1 | warm | 12.0 | 17.0 | 17.0 | 18.1 | 16.7 | os_cpu_cluster | 0.1 |
| volume_by_hour | 1 | warm | 24.0 | 26.0 | 26.0 | 16.1 | 42.8 | os_cpu_cluster | 14.1 |

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

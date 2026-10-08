> Target `altinity/local-single` is a local tier: this run backs parity and correctness only, never performance.

# Benchmark: clickstack

| Field | Value |
| --- | --- |
| Harness commit | `d75d88e` |
| ClickHouse | 26.3.33.10001.altinitystable (single-node) |
| Clusters | default |
| Sweep id | `20261008T072750Z` |
| Concurrency | 1 |
| Started | 2026-10-08T07:27:50Z |
| Finished | 2026-10-08T07:28:11Z |

## Ingest

Source `otel-direct`: 500000 rows in 17.57s (28452 events/sec), batch=10000.

## Queries

| Query | c | cache | p50 ms | p95 ms | p99 ms | QPS | CPU p50 ms | CPU src | Mem MB |
| --- | ---: | --- | ---: | ---: | ---: | ---: | ---: | --- | ---: |
| body_contains_24h | 1 | warm | 26.0 | 31.0 | 31.0 | 14.8 | 37.1 | os_cpu | 12.1 |
| logs_for_trace | 1 | warm | 9.0 | 10.0 | 10.0 | 20.5 | 9.8 | os_cpu | 0.1 |
| needle_search | 1 | warm | 8.0 | 9.0 | 9.0 | 19.5 | 5.6 | os_cpu | 0.0 |
| rollup_services_24h | 1 | warm | 8.0 | 10.0 | 10.0 | 20.2 | 9.8 | os_cpu | 0.1 |
| service_tail_1h | 1 | warm | 11.0 | 15.0 | 15.0 | 19.4 | 16.5 | os_cpu | 0.1 |
| severity_by_service_24h | 1 | warm | 17.0 | 32.0 | 32.0 | 15.2 | 24.7 | os_cpu | 9.2 |
| top_error_patterns_24h | 1 | warm | 19.0 | 21.0 | 21.0 | 16.6 | 30.5 | os_cpu | 0.1 |
| user_logs | 1 | warm | 14.0 | 15.0 | 15.0 | 17.3 | 20.5 | os_cpu | 0.1 |
| user_logs_items | 1 | warm | 12.0 | 14.0 | 14.0 | 18.3 | 17.9 | os_cpu | 0.1 |
| volume_by_hour | 1 | warm | 25.0 | 29.0 | 29.0 | 15.5 | 44.3 | os_cpu | 12.8 |

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

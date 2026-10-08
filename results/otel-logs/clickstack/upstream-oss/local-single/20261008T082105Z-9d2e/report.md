> Target `upstream-oss/local-single` is a local tier: this run backs parity and correctness only, never performance.

# Benchmark: clickstack

| Field | Value |
| --- | --- |
| Harness commit | `effdf02` |
| ClickHouse | 26.3.33.24 (single-node) |
| Clusters | default |
| Sweep id | `20261008T082105Z` |
| Concurrency | 1 |
| Started | 2026-10-08T08:21:05Z |
| Finished | 2026-10-08T08:21:26Z |

## Ingest

Source `otel-direct`: 500000 rows in 17.71s (28231 events/sec), batch=10000.

## Queries

| Query | c | cache | p50 ms | p95 ms | p99 ms | QPS | CPU p50 ms | CPU src | Mem MB |
| --- | ---: | --- | ---: | ---: | ---: | ---: | ---: | --- | ---: |
| body_contains_24h | 1 | warm | 29.0 | 30.0 | 30.0 | 14.5 | 37.4 | os_cpu | 10.7 |
| logs_for_trace | 1 | warm | 13.0 | 13.0 | 13.0 | 18.4 | 14.0 | os_cpu | 0.1 |
| needle_search | 1 | warm | 9.0 | 10.0 | 10.0 | 20.7 | 7.8 | os_cpu | 0.0 |
| rollup_services_24h | 1 | warm | 8.0 | 9.0 | 9.0 | 19.3 | 10.9 | os_cpu | 0.1 |
| service_tail_1h | 1 | warm | 11.0 | 14.0 | 14.0 | 19.8 | 15.9 | os_cpu | 0.1 |
| severity_by_service_24h | 1 | warm | 17.0 | 20.0 | 20.0 | 16.6 | 25.9 | os_cpu | 9.2 |
| top_error_patterns_24h | 1 | warm | 26.0 | 37.0 | 37.0 | 14.0 | 37.1 | os_cpu | 0.1 |
| user_logs | 1 | warm | 14.0 | 20.0 | 20.0 | 17.0 | 18.6 | os_cpu | 0.1 |
| user_logs_items | 1 | warm | 14.0 | 16.0 | 16.0 | 17.9 | 20.1 | os_cpu | 0.1 |
| volume_by_hour | 1 | warm | 24.0 | 57.0 | 57.0 | 14.0 | 49.0 | os_cpu | 17.1 |

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

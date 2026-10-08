> Target `upstream-oss/local-cluster` is a local tier: this run backs parity and correctness only, never performance.

# Benchmark: exporter-default

| Field | Value |
| --- | --- |
| Harness commit | `d75d88e` |
| ClickHouse | 26.3.33.24 (multi-node) |
| Clusters | ch_eval |
| Sweep id | `20261008T073704Z` |
| Concurrency | 1 |
| Started | 2026-10-08T07:37:30Z |
| Finished | 2026-10-08T07:37:51Z |

## Ingest

Source `otel-direct`: 500000 rows in 17.57s (28456 events/sec), batch=10000.

## Queries

| Query | c | cache | p50 ms | p95 ms | p99 ms | QPS | CPU p50 ms | CPU src | Mem MB |
| --- | ---: | --- | ---: | ---: | ---: | ---: | ---: | --- | ---: |
| body_contains_24h | 1 | warm | 28.0 | 38.0 | 38.0 | 14.0 | 36.0 | os_cpu_cluster | 7.4 |
| logs_for_trace | 1 | warm | 14.0 | 15.0 | 15.0 | 18.3 | 11.9 | os_cpu_cluster | 0.1 |
| needle_search | 1 | warm | 12.0 | 16.0 | 16.0 | 18.8 | 4.2 | os_cpu_cluster | 0.0 |
| service_tail_1h | 1 | warm | 13.0 | 21.0 | 21.0 | 17.1 | 18.9 | os_cpu_cluster | 0.1 |
| severity_by_service_24h | 1 | warm | 16.0 | 27.0 | 27.0 | 16.6 | 22.8 | os_cpu_cluster | 9.8 |
| top_error_patterns_24h | 1 | warm | 22.0 | 28.0 | 28.0 | 15.3 | 33.3 | os_cpu_cluster | 0.1 |
| user_logs | 1 | warm | 20.0 | 21.0 | 21.0 | 16.1 | 18.2 | os_cpu_cluster | 0.1 |
| volume_by_hour | 1 | warm | 29.0 | 49.0 | 49.0 | 13.7 | 56.1 | os_cpu_cluster | 10.9 |

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
<summary><code>volume_by_hour</code></summary>

```sql
SELECT toStartOfHour(Timestamp) AS hour, SeverityText, count() AS logs
FROM otel_logs
WHERE Timestamp >= toDateTime({from:UInt32}) AND Timestamp < toDateTime({to:UInt32})
GROUP BY hour, SeverityText
ORDER BY hour, SeverityText
```

</details>

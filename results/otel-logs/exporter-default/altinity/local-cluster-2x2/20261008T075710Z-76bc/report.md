> Target `altinity/local-cluster-2x2` is a local tier: this run backs parity and correctness only, never performance.

# Benchmark: exporter-default

| Field | Value |
| --- | --- |
| Harness commit | `3701df2` |
| ClickHouse | 26.3.33.10001.altinitystable (multi-node) |
| Clusters | ch_eval |
| Sweep id | `20261008T075628Z` |
| Concurrency | 1 |
| Started | 2026-10-08T07:57:10Z |
| Finished | 2026-10-08T07:57:38Z |

## Ingest

Source `otel-direct`: 500000 rows in 23.33s (21428 events/sec), batch=10000.

## Queries

| Query | c | cache | p50 ms | p95 ms | p99 ms | QPS | CPU p50 ms | CPU src | Mem MB |
| --- | ---: | --- | ---: | ---: | ---: | ---: | ---: | --- | ---: |
| body_contains_24h | 1 | warm | 38.0 | 39.0 | 39.0 | 12.3 | 48.8 | os_cpu_cluster | 6.9 |
| logs_for_trace | 1 | warm | 16.0 | 23.0 | 23.0 | 10.9 | 17.8 | os_cpu_cluster | 0.5 |
| needle_search | 1 | warm | 16.0 | 21.0 | 21.0 | 18.2 | 15.6 | os_cpu_cluster | 0.5 |
| service_tail_1h | 1 | warm | 24.0 | 55.0 | 55.0 | 14.1 | 25.4 | os_cpu_cluster | 0.9 |
| severity_by_service_24h | 1 | warm | 34.0 | 37.0 | 37.0 | 12.9 | 38.4 | os_cpu_cluster | 9.7 |
| top_error_patterns_24h | 1 | warm | 35.0 | 35.0 | 35.0 | 13.1 | 47.8 | os_cpu_cluster | 0.5 |
| user_logs | 1 | warm | 25.0 | 33.0 | 33.0 | 14.4 | 27.9 | os_cpu_cluster | 1.5 |
| volume_by_hour | 1 | warm | 37.0 | 47.0 | 47.0 | 13.6 | 57.0 | os_cpu_cluster | 18.0 |

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

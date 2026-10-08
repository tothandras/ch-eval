> Target `altinity/local-cluster-s3-cow` is a local tier: this run backs parity and correctness only, never performance.

# Benchmark: exporter-default

| Field | Value |
| --- | --- |
| Harness commit | `a9d9670` |
| ClickHouse | 26.3.33.10001.altinitystable (multi-node) |
| Clusters | ch_eval |
| Sweep id | `20261008T083919Z` |
| Concurrency | 1 |
| Started | 2026-10-08T08:39:59Z |
| Finished | 2026-10-08T08:40:29Z |

## Ingest

Source `otel-direct`: 500000 rows in 25.66s (19488 events/sec), batch=10000.

## Queries

| Query | c | cache | p50 ms | p95 ms | p99 ms | QPS | CPU p50 ms | CPU src | Mem MB |
| --- | ---: | --- | ---: | ---: | ---: | ---: | ---: | --- | ---: |
| body_contains_24h | 1 | warm | 25.0 | 31.0 | 31.0 | 14.7 | 36.9 | os_cpu_cluster | 12.7 |
| logs_for_trace | 1 | warm | 8.0 | 8.0 | 8.0 | 20.9 | 8.8 | os_cpu_cluster | 2.3 |
| needle_search | 1 | warm | 7.0 | 7.0 | 7.0 | 21.2 | 6.3 | os_cpu_cluster | 1.1 |
| service_tail_1h | 1 | warm | 13.0 | 17.0 | 17.0 | 18.7 | 21.3 | os_cpu_cluster | 1.9 |
| severity_by_service_24h | 1 | warm | 16.0 | 17.0 | 17.0 | 16.7 | 25.4 | os_cpu_cluster | 14.9 |
| top_error_patterns_24h | 1 | warm | 18.0 | 27.0 | 27.0 | 15.8 | 27.7 | os_cpu_cluster | 6.0 |
| user_logs | 1 | warm | 14.0 | 23.0 | 23.0 | 15.9 | 19.1 | os_cpu_cluster | 1.0 |
| volume_by_hour | 1 | warm | 25.0 | 25.0 | 25.0 | 16.3 | 45.6 | os_cpu_cluster | 11.6 |

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

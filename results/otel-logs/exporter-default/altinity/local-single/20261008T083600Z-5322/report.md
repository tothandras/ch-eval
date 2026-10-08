> Target `altinity/local-single` is a local tier: this run backs parity and correctness only, never performance.

# Benchmark: exporter-default

| Field | Value |
| --- | --- |
| Harness commit | `a9d9670` |
| ClickHouse | 26.3.33.10001.altinitystable (single-node) |
| Clusters | default |
| Sweep id | `20261008T083540Z` |
| Concurrency | 1 |
| Started | 2026-10-08T08:36:00Z |
| Finished | 2026-10-08T08:36:16Z |

## Ingest

Source `otel-direct`: 500000 rows in 12.58s (39756 events/sec), batch=10000.

## Queries

| Query | c | cache | p50 ms | p95 ms | p99 ms | QPS | CPU p50 ms | CPU src | Mem MB |
| --- | ---: | --- | ---: | ---: | ---: | ---: | ---: | --- | ---: |
| body_contains_24h | 1 | warm | 23.0 | 23.0 | 23.0 | 15.5 | 38.3 | os_cpu | 10.1 |
| logs_for_trace | 1 | warm | 10.0 | 15.0 | 15.0 | 19.3 | 9.0 | os_cpu | 0.1 |
| needle_search | 1 | warm | 7.0 | 7.0 | 7.0 | 20.1 | 4.4 | os_cpu | 0.0 |
| service_tail_1h | 1 | warm | 11.0 | 13.0 | 13.0 | 19.9 | 16.7 | os_cpu | 0.2 |
| severity_by_service_24h | 1 | warm | 18.0 | 20.0 | 20.0 | 15.7 | 32.0 | os_cpu | 8.4 |
| top_error_patterns_24h | 1 | warm | 23.0 | 25.0 | 25.0 | 15.2 | 37.8 | os_cpu | 0.1 |
| user_logs | 1 | warm | 14.0 | 15.0 | 15.0 | 17.8 | 19.6 | os_cpu | 0.1 |
| volume_by_hour | 1 | warm | 23.0 | 25.0 | 25.0 | 16.7 | 49.7 | os_cpu | 13.7 |

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

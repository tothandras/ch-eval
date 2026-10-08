> Target `upstream-oss/kind` is a local tier: this run backs parity and correctness only, never performance.

# Benchmark: exporter-default

| Field | Value |
| --- | --- |
| Harness commit | `effdf02` |
| ClickHouse | 26.3.33.24 (multi-node) |
| Clusters | all-clusters, all-replicated, all-sharded, ch-eval, default |
| Sweep id | `20261008T082329Z` |
| Concurrency | 1 |
| Started | 2026-10-08T08:23:59Z |
| Finished | 2026-10-08T08:24:26Z |

## Ingest

Source `otel-direct`: 500018 rows in 22.06s (22664 events/sec), batch=10000.

## Queries

| Query | c | cache | p50 ms | p95 ms | p99 ms | QPS | CPU p50 ms | CPU src | Mem MB |
| --- | ---: | --- | ---: | ---: | ---: | ---: | ---: | --- | ---: |
| error_rate_24h | 1 | warm | 15.0 | 16.0 | 16.0 | 16.4 | 17.8 | os_cpu_cluster | 0.1 |
| latency_by_service_1h | 1 | warm | 17.0 | 17.0 | 17.0 | 17.3 | 20.5 | os_cpu_cluster | 0.2 |
| route_status_24h | 1 | warm | 25.0 | 42.0 | 42.0 | 12.7 | 42.5 | os_cpu_cluster | 4.1 |
| service_dependencies_1h | 1 | warm | 55.0 | 81.0 | 81.0 | 9.2 | 82.5 | os_cpu_cluster | 1.9 |
| slowest_traces_1h | 1 | warm | 20.0 | 22.0 | 22.0 | 14.5 | 28.5 | os_cpu_cluster | 0.2 |
| spans_by_hour | 1 | warm | 28.0 | 36.0 | 36.0 | 14.5 | 51.0 | os_cpu_cluster | 14.5 |
| top_operations_24h | 1 | warm | 13.0 | 17.0 | 17.0 | 16.9 | 16.8 | os_cpu_cluster | 0.1 |
| trace_by_id | 1 | warm | 16.0 | 17.0 | 17.0 | 15.7 | 16.4 | os_cpu_cluster | 0.0 |
| trace_by_id_lookup | 1 | warm | 32.0 | 77.0 | 77.0 | 11.7 | 28.7 | os_cpu_cluster | 0.2 |
| user_activity | 1 | warm | 18.0 | 20.0 | 20.0 | 16.7 | 23.2 | os_cpu_cluster | 0.1 |

<details>
<summary><code>error_rate_24h</code></summary>

```sql
SELECT ServiceName, count() AS spans, countIf(StatusCode = 'Error') AS errors, round(errors / spans, 6) AS error_rate
FROM otel_traces
WHERE Timestamp >= toDateTime({from_24h:UInt32}) AND Timestamp < toDateTime({to:UInt32}) AND SpanKind = 'Server'
GROUP BY ServiceName
ORDER BY ServiceName
```

</details>

<details>
<summary><code>latency_by_service_1h</code></summary>

```sql
SELECT ServiceName, toStartOfFiveMinutes(Timestamp) AS t, count() AS spans,
       quantileExact(0.5)(Duration) AS p50, quantileExact(0.95)(Duration) AS p95, quantileExact(0.99)(Duration) AS p99
FROM otel_traces
WHERE Timestamp >= toDateTime({from_1h:UInt32}) AND Timestamp < toDateTime({to:UInt32}) AND SpanKind = 'Server'
GROUP BY ServiceName, t
ORDER BY ServiceName, t
```

</details>

<details>
<summary><code>route_status_24h</code></summary>

```sql
SELECT SpanAttributes['http.response.status_code'] AS status, count() AS spans, quantileExact(0.95)(Duration) AS p95
FROM otel_traces
WHERE Timestamp >= toDateTime({from_24h:UInt32}) AND Timestamp < toDateTime({to:UInt32})
  AND ServiceName = {service:String} AND SpanKind = 'Server' AND SpanAttributes['http.route'] = {route:String}
GROUP BY status
ORDER BY status
```

</details>

<details>
<summary><code>service_dependencies_1h</code></summary>

```sql
SELECT p.ServiceName AS parent, c.ServiceName AS child, count() AS calls
FROM
(
    SELECT TraceId, ParentSpanId, ServiceName FROM otel_traces
    WHERE Timestamp >= toDateTime({from_1h:UInt32}) AND Timestamp < toDateTime({to:UInt32}) AND ParentSpanId != ''
) AS c
INNER JOIN
(
    SELECT TraceId, SpanId, ServiceName FROM otel_traces
    WHERE Timestamp >= toDateTime({from_1h:UInt32}) - 600 AND Timestamp < toDateTime({to:UInt32})
) AS p ON c.TraceId = p.TraceId AND c.ParentSpanId = p.SpanId
WHERE p.ServiceName != c.ServiceName
GROUP BY parent, child
ORDER BY parent, child
SETTINGS distributed_product_mode = 'local'
```

</details>

<details>
<summary><code>slowest_traces_1h</code></summary>

```sql
SELECT TraceId, ServiceName, SpanName, Duration, StatusCode
FROM otel_traces
WHERE Timestamp >= toDateTime({from_1h:UInt32}) AND Timestamp < toDateTime({to:UInt32}) AND ParentSpanId = ''
ORDER BY Duration DESC, TraceId
LIMIT 20
```

</details>

<details>
<summary><code>spans_by_hour</code></summary>

```sql
SELECT toStartOfHour(Timestamp) AS hour, ServiceName, count() AS spans, countIf(StatusCode = 'Error') AS errors
FROM otel_traces
WHERE Timestamp >= toDateTime({from:UInt32}) AND Timestamp < toDateTime({to:UInt32})
GROUP BY hour, ServiceName
ORDER BY hour, ServiceName
```

</details>

<details>
<summary><code>top_operations_24h</code></summary>

```sql
SELECT SpanName, count() AS spans, avg(Duration) AS avg_duration
FROM otel_traces
WHERE ServiceName = {service:String} AND Timestamp >= toDateTime({from_24h:UInt32}) AND Timestamp < toDateTime({to:UInt32})
GROUP BY SpanName
ORDER BY spans DESC, SpanName
LIMIT 10
```

</details>

<details>
<summary><code>trace_by_id</code></summary>

```sql
SELECT Timestamp, SpanId, ParentSpanId, ServiceName, SpanName, SpanKind, Duration, StatusCode
FROM otel_traces
WHERE TraceId = {trace_id:String}
ORDER BY Timestamp, SpanId
```

</details>

<details>
<summary><code>trace_by_id_lookup</code></summary>

```sql
WITH
    (SELECT min(Start) FROM otel_traces_trace_id_ts WHERE TraceId = {trace_id:String}) AS trace_start,
    (SELECT max(End) + 1 FROM otel_traces_trace_id_ts WHERE TraceId = {trace_id:String}) AS trace_end
SELECT Timestamp, SpanId, ParentSpanId, ServiceName, SpanName, SpanKind, Duration, StatusCode
FROM otel_traces
WHERE TraceId = {trace_id:String} AND Timestamp >= trace_start AND Timestamp <= trace_end
ORDER BY Timestamp, SpanId
```

</details>

<details>
<summary><code>user_activity</code></summary>

```sql
SELECT count() AS spans, uniqExact(TraceId) AS traces, min(Timestamp) AS first_seen, max(Timestamp) AS last_seen
FROM otel_traces
WHERE SpanAttributes['user.id'] = {user_id:String}
```

</details>

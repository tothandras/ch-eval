> Target `altinity/local-single` is a local tier: this run backs parity and correctness only, never performance.

# Benchmark: exporter-default

| Field | Value |
| --- | --- |
| Harness commit | `a9d9670` |
| ClickHouse | 26.3.33.10001.altinitystable (single-node) |
| Clusters | default |
| Sweep id | `20261008T083508Z` |
| Concurrency | 1 |
| Started | 2026-10-08T08:35:25Z |
| Finished | 2026-10-08T08:35:39Z |

## Ingest

Source `otel-direct`: 500018 rows in 10.93s (45727 events/sec), batch=10000.

## Queries

| Query | c | cache | p50 ms | p95 ms | p99 ms | QPS | CPU p50 ms | CPU src | Mem MB |
| --- | ---: | --- | ---: | ---: | ---: | ---: | ---: | --- | ---: |
| error_rate_24h | 1 | warm | 17.0 | 18.0 | 18.0 | 17.1 | 23.4 | os_cpu | 0.1 |
| latency_by_service_1h | 1 | warm | 17.0 | 17.0 | 17.0 | 17.8 | 22.1 | os_cpu | 0.2 |
| route_status_24h | 1 | warm | 33.0 | 45.0 | 45.0 | 13.0 | 47.8 | os_cpu | 4.1 |
| service_dependencies_1h | 1 | warm | 39.0 | 49.0 | 49.0 | 11.7 | 67.1 | os_cpu | 2.3 |
| slowest_traces_1h | 1 | warm | 15.0 | 16.0 | 16.0 | 17.4 | 23.0 | os_cpu | 0.2 |
| spans_by_hour | 1 | warm | 26.0 | 33.0 | 33.0 | 15.4 | 47.4 | os_cpu | 15.4 |
| top_operations_24h | 1 | warm | 10.0 | 10.0 | 10.0 | 19.4 | 14.0 | os_cpu | 0.1 |
| trace_by_id | 1 | warm | 13.0 | 14.0 | 14.0 | 19.1 | 14.2 | os_cpu | 0.0 |
| trace_by_id_lookup | 1 | warm | 28.0 | 30.0 | 30.0 | 14.8 | 27.3 | os_cpu | 0.2 |
| user_activity | 1 | warm | 26.0 | 37.0 | 37.0 | 15.5 | 26.2 | os_cpu | 0.1 |

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

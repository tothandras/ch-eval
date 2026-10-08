> Target `upstream-oss/local-single` is a local tier: this run backs parity and correctness only, never performance.

# Benchmark: exporter-default

| Field | Value |
| --- | --- |
| Harness commit | `effdf02` |
| ClickHouse | 26.3.33.24 (single-node) |
| Clusters | default |
| Sweep id | `20261008T082030Z` |
| Concurrency | 1 |
| Started | 2026-10-08T08:20:49Z |
| Finished | 2026-10-08T08:21:04Z |

## Ingest

Source `otel-direct`: 500018 rows in 11.33s (44129 events/sec), batch=10000.

## Queries

| Query | c | cache | p50 ms | p95 ms | p99 ms | QPS | CPU p50 ms | CPU src | Mem MB |
| --- | ---: | --- | ---: | ---: | ---: | ---: | ---: | --- | ---: |
| error_rate_24h | 1 | warm | 15.0 | 17.0 | 17.0 | 17.3 | 20.3 | os_cpu | 0.1 |
| latency_by_service_1h | 1 | warm | 17.0 | 18.0 | 18.0 | 18.5 | 25.1 | os_cpu | 0.1 |
| route_status_24h | 1 | warm | 27.0 | 33.0 | 33.0 | 13.6 | 40.2 | os_cpu | 4.1 |
| service_dependencies_1h | 1 | warm | 51.0 | 88.0 | 88.0 | 9.4 | 82.5 | os_cpu | 1.7 |
| slowest_traces_1h | 1 | warm | 15.0 | 31.0 | 31.0 | 15.7 | 22.1 | os_cpu | 0.2 |
| spans_by_hour | 1 | warm | 31.0 | 46.0 | 46.0 | 14.3 | 46.9 | os_cpu | 18.4 |
| top_operations_24h | 1 | warm | 12.0 | 17.0 | 17.0 | 17.7 | 16.6 | os_cpu | 0.1 |
| trace_by_id | 1 | warm | 18.0 | 19.0 | 19.0 | 15.5 | 14.1 | os_cpu | 0.0 |
| trace_by_id_lookup | 1 | warm | 29.0 | 40.0 | 40.0 | 13.7 | 28.9 | os_cpu | 0.2 |
| user_activity | 1 | warm | 19.0 | 20.0 | 20.0 | 16.8 | 25.5 | os_cpu | 0.1 |

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

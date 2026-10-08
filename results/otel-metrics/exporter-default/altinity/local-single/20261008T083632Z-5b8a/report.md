> Target `altinity/local-single` is a local tier: this run backs parity and correctness only, never performance.

# Benchmark: exporter-default

| Field | Value |
| --- | --- |
| Harness commit | `a9d9670` |
| ClickHouse | 26.3.33.10001.altinitystable (single-node) |
| Clusters | default |
| Sweep id | `20261008T083616Z` |
| Concurrency | 1 |
| Started | 2026-10-08T08:36:32Z |
| Finished | 2026-10-08T08:36:47Z |

## Ingest

Source `otel-direct`: 496716 rows in 9.12s (54450 events/sec), batch=10000.

## Queries

| Query | c | cache | p50 ms | p95 ms | p99 ms | QPS | CPU p50 ms | CPU src | Mem MB |
| --- | ---: | --- | ---: | ---: | ---: | ---: | ---: | --- | ---: |
| cpu_by_service_1h | 1 | warm | 12.0 | 15.0 | 15.0 | 19.3 | 15.1 | os_cpu | 0.1 |
| error_ratio_1h | 1 | warm | 150.0 | 160.0 | 160.0 | 5.3 | 196.3 | os_cpu | 22.1 |
| gc_pause_median_1h | 1 | warm | 8.0 | 12.0 | 12.0 | 19.8 | 9.7 | os_cpu | 0.1 |
| latency_p95_by_route_1h | 1 | warm | 132.0 | 180.0 | 180.0 | 5.7 | 130.9 | os_cpu | 15.2 |
| memory_by_pod_1h | 1 | warm | 27.0 | 28.0 | 28.0 | 14.6 | 31.2 | os_cpu | 4.1 |
| request_rate_by_route_1h | 1 | warm | 172.0 | 201.0 | 201.0 | 4.7 | 361.2 | os_cpu | 49.5 |
| rpc_count_by_method_1h | 1 | warm | 80.0 | 81.0 | 81.0 | 8.4 | 81.4 | os_cpu | 4.0 |
| series_by_metric | 1 | warm | 154.0 | 160.0 | 160.0 | 5.5 | 314.0 | os_cpu | 23.7 |
| user_requests | 1 | warm | 8.0 | 9.0 | 9.0 | 19.4 | 8.2 | os_cpu | 0.0 |

<details>
<summary><code>cpu_by_service_1h</code></summary>

```sql
SELECT ServiceName, toStartOfMinute(TimeUnix) AS t, round(avg(Value), 6) AS cpu
FROM otel_metrics_gauge
WHERE MetricName = {gauge:String} AND TimeUnix >= toDateTime({from_1h:UInt32}) AND TimeUnix < toDateTime({to:UInt32})
GROUP BY ServiceName, t
ORDER BY ServiceName, t
```

</details>

<details>
<summary><code>error_ratio_1h</code></summary>

```sql
SELECT ServiceName,
       sumIf(inc, Attributes['http.response.status_code'] = '500') AS errors, sum(inc) AS requests,
       round(errors / greatest(requests, 1), 6) AS error_ratio
FROM
(
    SELECT ServiceName, Attributes, max(Value) - min(Value) AS inc
    FROM otel_metrics_sum
    WHERE MetricName = {counter:String} AND TimeUnix >= toDateTime({from_1h:UInt32}) AND TimeUnix < toDateTime({to:UInt32})
    GROUP BY ServiceName, ResourceAttributes, Attributes, StartTimeUnix
)
GROUP BY ServiceName
ORDER BY ServiceName
```

</details>

<details>
<summary><code>gc_pause_median_1h</code></summary>

```sql
SELECT ServiceName, round(avg(arrayElement(ValueAtQuantiles.Value, indexOf(ValueAtQuantiles.Quantile, 0.5))), 6) AS median_pause
FROM otel_metrics_summary
WHERE MetricName = {summary:String} AND TimeUnix >= toDateTime({from_1h:UInt32}) AND TimeUnix < toDateTime({to:UInt32})
GROUP BY ServiceName
ORDER BY ServiceName
```

</details>

<details>
<summary><code>latency_p95_by_route_1h</code></summary>

```sql
SELECT route, total,
       if(idx > length(bounds), inf, bounds[idx]) AS p95_upper_bound
FROM
(
    SELECT route, any(bounds) AS bounds, sumForEach(delta) AS counts, arraySum(counts) AS total,
           arrayFirstIndex(c -> c >= 0.95 * total, arrayCumSum(counts)) AS idx
    FROM
    (
        SELECT Attributes['http.route'] AS route, any(ExplicitBounds) AS bounds,
               arrayMap((last, first) -> last - first, argMax(BucketCounts, TimeUnix), argMin(BucketCounts, TimeUnix)) AS delta
        FROM otel_metrics_histogram
        WHERE MetricName = {histo:String} AND TimeUnix >= toDateTime({from_1h:UInt32}) AND TimeUnix < toDateTime({to:UInt32})
        GROUP BY ServiceName, ResourceAttributes, Attributes, StartTimeUnix
    )
    GROUP BY route
)
WHERE total > 0
ORDER BY route
```

</details>

<details>
<summary><code>memory_by_pod_1h</code></summary>

```sql
SELECT ResourceAttributes['k8s.pod.name'] AS pod, max(Value) AS peak_bytes
FROM otel_metrics_gauge
WHERE MetricName = 'process.memory.usage' AND TimeUnix >= toDateTime({from_1h:UInt32}) AND TimeUnix < toDateTime({to:UInt32})
GROUP BY pod
ORDER BY pod
```

</details>

<details>
<summary><code>request_rate_by_route_1h</code></summary>

```sql
SELECT ServiceName, route, toStartOfFiveMinutes(TimeUnix) AS t, sum(delta) AS requests
FROM
(
    SELECT ServiceName, Attributes['http.route'] AS route, TimeUnix,
           Value - lagInFrame(Value, 1, Value) OVER (
               PARTITION BY cityHash64(ResourceAttributes, Attributes), StartTimeUnix ORDER BY TimeUnix
               ROWS BETWEEN UNBOUNDED PRECEDING AND CURRENT ROW) AS delta
    FROM otel_metrics_sum
    WHERE MetricName = {counter:String} AND TimeUnix >= toDateTime({from_1h:UInt32}) AND TimeUnix < toDateTime({to:UInt32})
)
GROUP BY ServiceName, route, t
ORDER BY ServiceName, route, t
```

</details>

<details>
<summary><code>rpc_count_by_method_1h</code></summary>

```sql
SELECT Attributes['rpc.method'] AS method, sum(inc) AS calls, round(sum(sum_inc) / greatest(sum(inc), 1), 6) AS avg_ms
FROM
(
    SELECT Attributes, max(Count) - min(Count) AS inc, max(Sum) - min(Sum) AS sum_inc
    FROM otel_metrics_exponential_histogram
    WHERE MetricName = {exphisto:String} AND TimeUnix >= toDateTime({from_1h:UInt32}) AND TimeUnix < toDateTime({to:UInt32})
    GROUP BY ServiceName, ResourceAttributes, Attributes, StartTimeUnix
)
GROUP BY method
ORDER BY method
```

</details>

<details>
<summary><code>series_by_metric</code></summary>

```sql
SELECT MetricName, uniqExact(cityHash64(ResourceAttributes, Attributes)) AS series, count() AS points
FROM
(
    SELECT MetricName, ResourceAttributes, Attributes FROM otel_metrics_gauge WHERE TimeUnix >= toDateTime({from:UInt32}) AND TimeUnix < toDateTime({to:UInt32})
    UNION ALL
    SELECT MetricName, ResourceAttributes, Attributes FROM otel_metrics_sum WHERE TimeUnix >= toDateTime({from:UInt32}) AND TimeUnix < toDateTime({to:UInt32})
    UNION ALL
    SELECT MetricName, ResourceAttributes, Attributes FROM otel_metrics_histogram WHERE TimeUnix >= toDateTime({from:UInt32}) AND TimeUnix < toDateTime({to:UInt32})
    UNION ALL
    SELECT MetricName, ResourceAttributes, Attributes FROM otel_metrics_exponential_histogram WHERE TimeUnix >= toDateTime({from:UInt32}) AND TimeUnix < toDateTime({to:UInt32})
    UNION ALL
    SELECT MetricName, ResourceAttributes, Attributes FROM otel_metrics_summary WHERE TimeUnix >= toDateTime({from:UInt32}) AND TimeUnix < toDateTime({to:UInt32})
)
GROUP BY MetricName
ORDER BY MetricName
```

</details>

<details>
<summary><code>user_requests</code></summary>

```sql
SELECT ServiceName, max(Value) - min(Value) AS requests
FROM otel_metrics_sum
WHERE MetricName = 'app.user.requests' AND Attributes['user.id'] = {user_id:String}
GROUP BY ServiceName
ORDER BY ServiceName
```

</details>

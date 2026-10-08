> Target `upstream-oss/kind` is a local tier: this run backs parity and correctness only, never performance.

# Benchmark: baseline-openmeter

| Field | Value |
| --- | --- |
| Harness commit | `effdf02` |
| ClickHouse | 26.3.33.24 (multi-node) |
| Clusters | all-clusters, all-replicated, all-sharded, ch-eval, default |
| Sweep id | `20261008T082217Z` |
| Concurrency | 1 |
| Started | 2026-10-08T08:22:17Z |
| Finished | 2026-10-08T08:22:48Z |

## Ingest

Source `go-seeder`: 1000000 rows in 16.49s (60629 events/sec), batch=10000.

## Queries

| Query | c | cache | p50 ms | p95 ms | p99 ms | QPS | CPU p50 ms | CPU src | Mem MB |
| --- | ---: | --- | ---: | ---: | ---: | ---: | ---: | --- | ---: |
| agent_runs_by_name | 1 | warm | 11.0 | 15.0 | 15.0 | 18.1 | 14.0 | os_cpu_cluster | 6.9 |
| avg_hour | 1 | warm | 47.0 | 66.0 | 66.0 | 11.0 | 46.9 | os_cpu_cluster | 8.6 |
| count_hour | 1 | warm | 16.0 | 17.0 | 17.0 | 17.1 | 16.8 | os_cpu_cluster | 4.0 |
| count_total | 1 | warm | 13.0 | 34.0 | 34.0 | 15.8 | 14.4 | os_cpu_cluster | 4.1 |
| distinct_subjects | 1 | warm | 22.0 | 25.0 | 25.0 | 15.0 | 29.8 | os_cpu_cluster | 5.5 |
| kong_api_request_total | 1 | warm | 14.0 | 19.0 | 19.0 | 16.2 | 13.0 | os_cpu_cluster | 0.1 |
| kong_llm_tokens_total | 1 | warm | 80.0 | 89.0 | 89.0 | 7.7 | 83.3 | os_cpu_cluster | 24.8 |
| kong_status_by_route | 1 | warm | 46.0 | 56.0 | 56.0 | 11.4 | 52.0 | os_cpu_cluster | 23.2 |
| latest_hour | 1 | warm | 49.0 | 52.0 | 52.0 | 11.3 | 50.1 | os_cpu_cluster | 13.7 |
| llm_tokens_by_model | 1 | warm | 38.0 | 44.0 | 44.0 | 13.0 | 36.8 | os_cpu_cluster | 19.0 |
| max_hour | 1 | warm | 33.0 | 64.0 | 64.0 | 12.3 | 37.9 | os_cpu_cluster | 8.6 |
| min_hour | 1 | warm | 37.0 | 64.0 | 64.0 | 11.9 | 40.1 | os_cpu_cluster | 8.6 |
| sum_day | 1 | warm | 38.0 | 63.0 | 63.0 | 11.9 | 41.7 | os_cpu_cluster | 8.6 |
| sum_day_by_dim | 1 | warm | 27.0 | 33.0 | 33.0 | 14.1 | 30.6 | os_cpu_cluster | 26.1 |
| sum_day_by_subject | 1 | warm | 37.0 | 38.0 | 38.0 | 13.3 | 40.4 | os_cpu_cluster | 7.2 |
| sum_hour | 1 | warm | 43.0 | 51.0 | 51.0 | 12.2 | 46.0 | os_cpu_cluster | 7.2 |
| sum_hour_group1 | 1 | warm | 44.0 | 50.0 | 50.0 | 11.9 | 49.0 | os_cpu_cluster | 5.9 |
| sum_hour_group1_group2 | 1 | warm | 42.0 | 46.0 | 46.0 | 12.2 | 46.7 | os_cpu_cluster | 5.9 |
| sum_hour_group1_group2_no_prewhere | 1 | warm | 45.0 | 61.0 | 61.0 | 11.4 | 50.1 | os_cpu_cluster | 5.9 |
| sum_hour_group1_no_prewhere | 1 | warm | 47.0 | 50.0 | 50.0 | 11.7 | 49.7 | os_cpu_cluster | 5.9 |
| sum_hour_tz | 1 | warm | 28.0 | 29.0 | 29.0 | 14.9 | 32.2 | os_cpu_cluster | 7.2 |
| sum_minute | 1 | warm | 47.0 | 47.0 | 47.0 | 11.1 | 50.5 | os_cpu_cluster | 11.2 |
| sum_month | 1 | warm | 45.0 | 45.0 | 45.0 | 12.5 | 46.1 | os_cpu_cluster | 9.9 |
| sum_no_window | 1 | warm | 35.0 | 49.0 | 49.0 | 12.5 | 38.9 | os_cpu_cluster | 7.2 |
| sum_total | 1 | warm | 35.0 | 37.0 | 37.0 | 13.5 | 38.8 | os_cpu_cluster | 8.3 |
| sum_total_by_subject | 1 | warm | 31.0 | 33.0 | 33.0 | 14.5 | 33.9 | os_cpu_cluster | 8.6 |
| sum_total_filter1 | 1 | warm | 36.0 | 37.0 | 37.0 | 13.4 | 42.4 | os_cpu_cluster | 5.9 |
| sum_total_filter2 | 1 | warm | 34.0 | 36.0 | 36.0 | 13.5 | 40.8 | os_cpu_cluster | 10.0 |
| unique_count_hour | 1 | warm | 39.0 | 54.0 | 54.0 | 11.6 | 41.5 | os_cpu_cluster | 7.1 |
| workload_seconds_by_region | 1 | warm | 19.0 | 30.0 | 30.0 | 16.6 | 21.4 | os_cpu_cluster | 14.6 |

<details>
<summary><code>agent_runs_by_name</code></summary>

```sql
SELECT
  tumbleStart(baseline_openmeter_events.time, toIntervalHour(1), 'UTC') AS windowstart,
  tumbleEnd(baseline_openmeter_events.time, toIntervalHour(1), 'UTC') AS windowend,
  count(*) AS value,
  JSON_VALUE(baseline_openmeter_events.data, '$.agent_name') AS agent_name
FROM baseline_openmeter_events
WHERE baseline_openmeter_events.namespace = {namespace:String}
  AND baseline_openmeter_events.type = 'agent_run'
  AND baseline_openmeter_events.subject IN {subjects:Array(String)}
  AND baseline_openmeter_events.time >= {from:DateTime}
  AND baseline_openmeter_events.time < {to:DateTime}
GROUP BY windowstart, windowend, agent_name
ORDER BY windowstart
SETTINGS optimize_move_to_prewhere = 1, allow_reorder_prewhere_conditions = 1
```

</details>

<details>
<summary><code>avg_hour</code></summary>

```sql
SELECT
  tumbleStart(baseline_openmeter_events.time, toIntervalHour(1), 'UTC') AS windowstart,
  tumbleEnd(baseline_openmeter_events.time, toIntervalHour(1), 'UTC') AS windowend,
  avg(toDecimal128OrNull(JSON_VALUE(baseline_openmeter_events.data, '$.value'), 19)) AS value,
  baseline_openmeter_events.subject AS subject
FROM baseline_openmeter_events
WHERE baseline_openmeter_events.namespace = {namespace:String}
  AND baseline_openmeter_events.type = {type:String}
  AND baseline_openmeter_events.subject IN {subjects:Array(String)}
  AND baseline_openmeter_events.time >= {from:DateTime}
  AND baseline_openmeter_events.time < {to:DateTime}
GROUP BY windowstart, windowend, subject
ORDER BY windowstart
SETTINGS optimize_move_to_prewhere = 1, allow_reorder_prewhere_conditions = 1
```

</details>

<details>
<summary><code>count_hour</code></summary>

```sql
SELECT
  tumbleStart(baseline_openmeter_events.time, toIntervalHour(1), 'UTC') AS windowstart,
  tumbleEnd(baseline_openmeter_events.time, toIntervalHour(1), 'UTC') AS windowend,
  count(*) AS value,
  baseline_openmeter_events.subject AS subject
FROM baseline_openmeter_events
WHERE baseline_openmeter_events.namespace = {namespace:String}
  AND baseline_openmeter_events.type = {type:String}
  AND baseline_openmeter_events.subject IN {subjects:Array(String)}
  AND baseline_openmeter_events.time >= {from:DateTime}
  AND baseline_openmeter_events.time < {to:DateTime}
GROUP BY windowstart, windowend, subject
ORDER BY windowstart
SETTINGS optimize_move_to_prewhere = 1, allow_reorder_prewhere_conditions = 1
```

</details>

<details>
<summary><code>count_total</code></summary>

```sql
SELECT
  tumbleStart(min(baseline_openmeter_events.time), toIntervalMinute(1)) AS windowstart,
  tumbleEnd(max(baseline_openmeter_events.time), toIntervalMinute(1)) AS windowend,
  count(*) AS value
FROM baseline_openmeter_events
WHERE baseline_openmeter_events.namespace = {namespace:String}
  AND baseline_openmeter_events.type = {type:String}
  AND baseline_openmeter_events.subject IN {subjects:Array(String)}
  AND baseline_openmeter_events.time >= {from:DateTime}
  AND baseline_openmeter_events.time < {to:DateTime}
SETTINGS optimize_move_to_prewhere = 1, allow_reorder_prewhere_conditions = 1
```

</details>

<details>
<summary><code>distinct_subjects</code></summary>

```sql
SELECT DISTINCT baseline_openmeter_events.subject AS subject
FROM baseline_openmeter_events
WHERE baseline_openmeter_events.namespace = {namespace:String}
  AND baseline_openmeter_events.type = {type:String}
  AND baseline_openmeter_events.time >= {from:DateTime}
  AND baseline_openmeter_events.time < {to:DateTime}
ORDER BY subject
```

</details>

<details>
<summary><code>kong_api_request_total</code></summary>

```sql
SELECT count(*) AS value
FROM baseline_openmeter_events
WHERE baseline_openmeter_events.namespace = {namespace:String}
  AND baseline_openmeter_events.type = 'kong.api_request'
  AND baseline_openmeter_events.time >= {from:DateTime}
  AND baseline_openmeter_events.time < {to:DateTime}
SETTINGS optimize_move_to_prewhere = 1, allow_reorder_prewhere_conditions = 1
```

</details>

<details>
<summary><code>kong_llm_tokens_total</code></summary>

```sql
SELECT sum(toDecimal128OrNull(JSON_VALUE(baseline_openmeter_events.data, '$.tokens'), 19)) AS value
FROM baseline_openmeter_events
WHERE baseline_openmeter_events.namespace = {namespace:String}
  AND baseline_openmeter_events.type = 'kong.llm_request'
  AND baseline_openmeter_events.time >= {from:DateTime}
  AND baseline_openmeter_events.time < {to:DateTime}
SETTINGS optimize_move_to_prewhere = 1, allow_reorder_prewhere_conditions = 1
```

</details>

<details>
<summary><code>kong_status_by_route</code></summary>

```sql
SELECT
  tumbleStart(baseline_openmeter_events.time, toIntervalHour(1), 'UTC') AS windowstart,
  tumbleEnd(baseline_openmeter_events.time, toIntervalHour(1), 'UTC') AS windowend,
  count(*) AS value,
  JSON_VALUE(baseline_openmeter_events.data, '$.response_http_status') AS status,
  JSON_VALUE(baseline_openmeter_events.data, '$.route_name') AS route
FROM baseline_openmeter_events
WHERE baseline_openmeter_events.namespace = {namespace:String}
  AND baseline_openmeter_events.type = 'kong.api_request'
  AND baseline_openmeter_events.subject IN {subjects:Array(String)}
  AND baseline_openmeter_events.time >= {from:DateTime}
  AND baseline_openmeter_events.time < {to:DateTime}
GROUP BY windowstart, windowend, status, route
ORDER BY windowstart
SETTINGS optimize_move_to_prewhere = 1, allow_reorder_prewhere_conditions = 1
```

</details>

<details>
<summary><code>latest_hour</code></summary>

```sql
SELECT
  tumbleStart(baseline_openmeter_events.time, toIntervalHour(1), 'UTC') AS windowstart,
  tumbleEnd(baseline_openmeter_events.time, toIntervalHour(1), 'UTC') AS windowend,
  argMax(toDecimal128OrNull(JSON_VALUE(baseline_openmeter_events.data, '$.value'), 19), (baseline_openmeter_events.time, baseline_openmeter_events.store_row_id)) AS value,
  baseline_openmeter_events.subject AS subject
FROM baseline_openmeter_events
WHERE baseline_openmeter_events.namespace = {namespace:String}
  AND baseline_openmeter_events.type = {type:String}
  AND baseline_openmeter_events.subject IN {subjects:Array(String)}
  AND baseline_openmeter_events.time >= {from:DateTime}
  AND baseline_openmeter_events.time < {to:DateTime}
GROUP BY windowstart, windowend, subject
ORDER BY windowstart
SETTINGS optimize_move_to_prewhere = 1, allow_reorder_prewhere_conditions = 1
```

</details>

<details>
<summary><code>llm_tokens_by_model</code></summary>

```sql
SELECT
  tumbleStart(baseline_openmeter_events.time, toIntervalHour(1), 'UTC') AS windowstart,
  tumbleEnd(baseline_openmeter_events.time, toIntervalHour(1), 'UTC') AS windowend,
  sum(toDecimal128OrNull(JSON_VALUE(baseline_openmeter_events.data, '$.tokens'), 19)) AS value,
  JSON_VALUE(baseline_openmeter_events.data, '$.model') AS model,
  JSON_VALUE(baseline_openmeter_events.data, '$.provider') AS provider
FROM baseline_openmeter_events
WHERE baseline_openmeter_events.namespace = {namespace:String}
  AND baseline_openmeter_events.type = 'kong.llm_request'
  AND baseline_openmeter_events.subject IN {subjects:Array(String)}
  AND baseline_openmeter_events.time >= {from:DateTime}
  AND baseline_openmeter_events.time < {to:DateTime}
GROUP BY windowstart, windowend, model, provider
ORDER BY windowstart
SETTINGS optimize_move_to_prewhere = 1, allow_reorder_prewhere_conditions = 1
```

</details>

<details>
<summary><code>max_hour</code></summary>

```sql
SELECT
  tumbleStart(baseline_openmeter_events.time, toIntervalHour(1), 'UTC') AS windowstart,
  tumbleEnd(baseline_openmeter_events.time, toIntervalHour(1), 'UTC') AS windowend,
  max(toDecimal128OrNull(JSON_VALUE(baseline_openmeter_events.data, '$.value'), 19)) AS value,
  baseline_openmeter_events.subject AS subject
FROM baseline_openmeter_events
WHERE baseline_openmeter_events.namespace = {namespace:String}
  AND baseline_openmeter_events.type = {type:String}
  AND baseline_openmeter_events.subject IN {subjects:Array(String)}
  AND baseline_openmeter_events.time >= {from:DateTime}
  AND baseline_openmeter_events.time < {to:DateTime}
GROUP BY windowstart, windowend, subject
ORDER BY windowstart
SETTINGS optimize_move_to_prewhere = 1, allow_reorder_prewhere_conditions = 1
```

</details>

<details>
<summary><code>min_hour</code></summary>

```sql
SELECT
  tumbleStart(baseline_openmeter_events.time, toIntervalHour(1), 'UTC') AS windowstart,
  tumbleEnd(baseline_openmeter_events.time, toIntervalHour(1), 'UTC') AS windowend,
  min(toDecimal128OrNull(JSON_VALUE(baseline_openmeter_events.data, '$.value'), 19)) AS value,
  baseline_openmeter_events.subject AS subject
FROM baseline_openmeter_events
WHERE baseline_openmeter_events.namespace = {namespace:String}
  AND baseline_openmeter_events.type = {type:String}
  AND baseline_openmeter_events.subject IN {subjects:Array(String)}
  AND baseline_openmeter_events.time >= {from:DateTime}
  AND baseline_openmeter_events.time < {to:DateTime}
GROUP BY windowstart, windowend, subject
ORDER BY windowstart
SETTINGS optimize_move_to_prewhere = 1, allow_reorder_prewhere_conditions = 1
```

</details>

<details>
<summary><code>sum_day</code></summary>

```sql
SELECT
  tumbleStart(baseline_openmeter_events.time, toIntervalDay(1), 'UTC') AS windowstart,
  tumbleEnd(baseline_openmeter_events.time, toIntervalDay(1), 'UTC') AS windowend,
  sum(toDecimal128OrNull(JSON_VALUE(baseline_openmeter_events.data, '$.value'), 19)) AS value,
  baseline_openmeter_events.subject AS subject
FROM baseline_openmeter_events
WHERE baseline_openmeter_events.namespace = {namespace:String}
  AND baseline_openmeter_events.type = {type:String}
  AND baseline_openmeter_events.subject IN {subjects:Array(String)}
  AND baseline_openmeter_events.time >= {from:DateTime}
  AND baseline_openmeter_events.time < {to:DateTime}
GROUP BY windowstart, windowend, subject
ORDER BY windowstart
SETTINGS optimize_move_to_prewhere = 1, allow_reorder_prewhere_conditions = 1
```

</details>

<details>
<summary><code>sum_day_by_dim</code></summary>

```sql
SELECT
  tumbleStart(baseline_openmeter_events.time, toIntervalDay(1), 'UTC') AS windowstart,
  windowstart + toIntervalDay(1) AS windowend,
  sum(toDecimal128OrNull(JSON_VALUE(baseline_openmeter_events.data, '$.tokens'), 19)) AS value,
  JSON_VALUE(baseline_openmeter_events.data, '$.model') AS model_id,
  baseline_openmeter_events.subject AS subject
FROM baseline_openmeter_events
WHERE baseline_openmeter_events.namespace = {namespace:String}
  AND baseline_openmeter_events.type = 'kong.llm_request'
  AND baseline_openmeter_events.subject IN {subjects:Array(String)}
  AND baseline_openmeter_events.time >= {from:DateTime}
  AND baseline_openmeter_events.time < {to:DateTime}
GROUP BY windowstart, windowend, model_id, subject
ORDER BY windowstart
SETTINGS optimize_move_to_prewhere = 1, allow_reorder_prewhere_conditions = 1
```

</details>

<details>
<summary><code>sum_day_by_subject</code></summary>

```sql
SELECT
  tumbleStart(baseline_openmeter_events.time, toIntervalDay(1), 'UTC') AS windowstart,
  windowstart + toIntervalDay(1) AS windowend,
  sum(toDecimal128OrNull(JSON_VALUE(baseline_openmeter_events.data, '$.value'), 19)) AS value,
  baseline_openmeter_events.subject AS subject
FROM baseline_openmeter_events
WHERE baseline_openmeter_events.namespace = {namespace:String}
  AND baseline_openmeter_events.type = {type:String}
  AND baseline_openmeter_events.subject IN {subjects:Array(String)}
  AND baseline_openmeter_events.time >= {from:DateTime}
  AND baseline_openmeter_events.time < {to:DateTime}
GROUP BY windowstart, windowend, subject
ORDER BY windowstart
SETTINGS optimize_move_to_prewhere = 1, allow_reorder_prewhere_conditions = 1
```

</details>

<details>
<summary><code>sum_hour</code></summary>

```sql
SELECT
  tumbleStart(baseline_openmeter_events.time, toIntervalHour(1), 'UTC') AS windowstart,
  tumbleEnd(baseline_openmeter_events.time, toIntervalHour(1), 'UTC') AS windowend,
  sum(toDecimal128OrNull(JSON_VALUE(baseline_openmeter_events.data, '$.value'), 19)) AS value,
  baseline_openmeter_events.subject AS subject
FROM baseline_openmeter_events
WHERE baseline_openmeter_events.namespace = {namespace:String}
  AND baseline_openmeter_events.type = {type:String}
  AND baseline_openmeter_events.subject IN {subjects:Array(String)}
  AND baseline_openmeter_events.time >= {from:DateTime}
  AND baseline_openmeter_events.time < {to:DateTime}
GROUP BY windowstart, windowend, subject
ORDER BY windowstart
SETTINGS optimize_move_to_prewhere = 1, allow_reorder_prewhere_conditions = 1
```

</details>

<details>
<summary><code>sum_hour_group1</code></summary>

```sql
SELECT
  tumbleStart(baseline_openmeter_events.time, toIntervalHour(1), 'UTC') AS windowstart,
  tumbleEnd(baseline_openmeter_events.time, toIntervalHour(1), 'UTC') AS windowend,
  sum(toDecimal128OrNull(JSON_VALUE(baseline_openmeter_events.data, '$.value'), 19)) AS value,
  baseline_openmeter_events.subject AS subject,
  JSON_VALUE(baseline_openmeter_events.data, '$.group1') AS group1
FROM baseline_openmeter_events
WHERE baseline_openmeter_events.namespace = {namespace:String}
  AND baseline_openmeter_events.type = {type:String}
  AND baseline_openmeter_events.subject IN {subjects:Array(String)}
  AND baseline_openmeter_events.time >= {from:DateTime}
  AND baseline_openmeter_events.time < {to:DateTime}
  AND JSON_VALUE(baseline_openmeter_events.data, '$.group1') = {group1:String}
GROUP BY windowstart, windowend, subject, group1
ORDER BY windowstart
SETTINGS optimize_move_to_prewhere = 1, allow_reorder_prewhere_conditions = 1
```

</details>

<details>
<summary><code>sum_hour_group1_group2</code></summary>

```sql
SELECT
  tumbleStart(baseline_openmeter_events.time, toIntervalHour(1), 'UTC') AS windowstart,
  tumbleEnd(baseline_openmeter_events.time, toIntervalHour(1), 'UTC') AS windowend,
  sum(toDecimal128OrNull(JSON_VALUE(baseline_openmeter_events.data, '$.value'), 19)) AS value,
  baseline_openmeter_events.subject AS subject,
  JSON_VALUE(baseline_openmeter_events.data, '$.group1') AS group1,
  JSON_VALUE(baseline_openmeter_events.data, '$.group2') AS group2
FROM baseline_openmeter_events
WHERE baseline_openmeter_events.namespace = {namespace:String}
  AND baseline_openmeter_events.type = {type:String}
  AND baseline_openmeter_events.subject IN {subjects:Array(String)}
  AND baseline_openmeter_events.time >= {from:DateTime}
  AND baseline_openmeter_events.time < {to:DateTime}
  AND JSON_VALUE(baseline_openmeter_events.data, '$.group1') = {group1:String}
  AND JSON_VALUE(baseline_openmeter_events.data, '$.group2') = {group2:String}
GROUP BY windowstart, windowend, subject, group1, group2
ORDER BY windowstart
SETTINGS optimize_move_to_prewhere = 1, allow_reorder_prewhere_conditions = 1
```

</details>

<details>
<summary><code>sum_hour_group1_group2_no_prewhere</code></summary>

```sql
SELECT
  tumbleStart(baseline_openmeter_events.time, toIntervalHour(1), 'UTC') AS windowstart,
  tumbleEnd(baseline_openmeter_events.time, toIntervalHour(1), 'UTC') AS windowend,
  sum(toDecimal128OrNull(JSON_VALUE(baseline_openmeter_events.data, '$.value'), 19)) AS value,
  baseline_openmeter_events.subject AS subject,
  JSON_VALUE(baseline_openmeter_events.data, '$.group1') AS group1,
  JSON_VALUE(baseline_openmeter_events.data, '$.group2') AS group2
FROM baseline_openmeter_events
WHERE baseline_openmeter_events.namespace = {namespace:String}
  AND baseline_openmeter_events.type = {type:String}
  AND baseline_openmeter_events.subject IN {subjects:Array(String)}
  AND baseline_openmeter_events.time >= {from:DateTime}
  AND baseline_openmeter_events.time < {to:DateTime}
  AND JSON_VALUE(baseline_openmeter_events.data, '$.group1') = {group1:String}
  AND JSON_VALUE(baseline_openmeter_events.data, '$.group2') = {group2:String}
GROUP BY windowstart, windowend, subject, group1, group2
ORDER BY windowstart
```

</details>

<details>
<summary><code>sum_hour_group1_no_prewhere</code></summary>

```sql
SELECT
  tumbleStart(baseline_openmeter_events.time, toIntervalHour(1), 'UTC') AS windowstart,
  tumbleEnd(baseline_openmeter_events.time, toIntervalHour(1), 'UTC') AS windowend,
  sum(toDecimal128OrNull(JSON_VALUE(baseline_openmeter_events.data, '$.value'), 19)) AS value,
  baseline_openmeter_events.subject AS subject,
  JSON_VALUE(baseline_openmeter_events.data, '$.group1') AS group1
FROM baseline_openmeter_events
WHERE baseline_openmeter_events.namespace = {namespace:String}
  AND baseline_openmeter_events.type = {type:String}
  AND baseline_openmeter_events.subject IN {subjects:Array(String)}
  AND baseline_openmeter_events.time >= {from:DateTime}
  AND baseline_openmeter_events.time < {to:DateTime}
  AND JSON_VALUE(baseline_openmeter_events.data, '$.group1') = {group1:String}
GROUP BY windowstart, windowend, subject, group1
ORDER BY windowstart
```

</details>

<details>
<summary><code>sum_hour_tz</code></summary>

```sql
SELECT
  tumbleStart(baseline_openmeter_events.time, toIntervalHour(1), 'America/New_York') AS windowstart,
  tumbleEnd(baseline_openmeter_events.time, toIntervalHour(1), 'America/New_York') AS windowend,
  sum(toDecimal128OrNull(JSON_VALUE(baseline_openmeter_events.data, '$.value'), 19)) AS value,
  baseline_openmeter_events.subject AS subject
FROM baseline_openmeter_events
WHERE baseline_openmeter_events.namespace = {namespace:String}
  AND baseline_openmeter_events.type = {type:String}
  AND baseline_openmeter_events.subject IN {subjects:Array(String)}
  AND baseline_openmeter_events.time >= {from:DateTime}
  AND baseline_openmeter_events.time < {to:DateTime}
GROUP BY windowstart, windowend, subject
ORDER BY windowstart
SETTINGS optimize_move_to_prewhere = 1, allow_reorder_prewhere_conditions = 1
```

</details>

<details>
<summary><code>sum_minute</code></summary>

```sql
SELECT
  tumbleStart(baseline_openmeter_events.time, toIntervalMinute(1), 'UTC') AS windowstart,
  tumbleEnd(baseline_openmeter_events.time, toIntervalMinute(1), 'UTC') AS windowend,
  sum(toDecimal128OrNull(JSON_VALUE(baseline_openmeter_events.data, '$.value'), 19)) AS value,
  baseline_openmeter_events.subject AS subject
FROM baseline_openmeter_events
WHERE baseline_openmeter_events.namespace = {namespace:String}
  AND baseline_openmeter_events.type = {type:String}
  AND baseline_openmeter_events.subject IN {subjects:Array(String)}
  AND baseline_openmeter_events.time >= {from:DateTime}
  AND baseline_openmeter_events.time < {to:DateTime}
GROUP BY windowstart, windowend, subject
ORDER BY windowstart
SETTINGS optimize_move_to_prewhere = 1, allow_reorder_prewhere_conditions = 1
```

</details>

<details>
<summary><code>sum_month</code></summary>

```sql
SELECT
  toDateTime(tumbleStart(baseline_openmeter_events.time, toIntervalMonth(1), 'UTC'), 'UTC') AS windowstart,
  toDateTime(tumbleEnd(baseline_openmeter_events.time, toIntervalMonth(1), 'UTC'), 'UTC') AS windowend,
  sum(toDecimal128OrNull(JSON_VALUE(baseline_openmeter_events.data, '$.value'), 19)) AS value,
  baseline_openmeter_events.subject AS subject
FROM baseline_openmeter_events
WHERE baseline_openmeter_events.namespace = {namespace:String}
  AND baseline_openmeter_events.type = {type:String}
  AND baseline_openmeter_events.subject IN {subjects:Array(String)}
  AND baseline_openmeter_events.time >= {from:DateTime}
  AND baseline_openmeter_events.time < {to:DateTime}
GROUP BY windowstart, windowend, subject
ORDER BY windowstart
SETTINGS optimize_move_to_prewhere = 1, allow_reorder_prewhere_conditions = 1
```

</details>

<details>
<summary><code>sum_no_window</code></summary>

```sql
SELECT
  tumbleStart(min(baseline_openmeter_events.time), toIntervalMinute(1), 'UTC') AS windowstart,
  tumbleEnd(max(baseline_openmeter_events.time), toIntervalMinute(1), 'UTC') AS windowend,
  sum(toDecimal128OrNull(JSON_VALUE(baseline_openmeter_events.data, '$.value'), 19)) AS value,
  baseline_openmeter_events.subject AS subject
FROM baseline_openmeter_events
WHERE baseline_openmeter_events.namespace = {namespace:String}
  AND baseline_openmeter_events.type = {type:String}
  AND baseline_openmeter_events.subject IN {subjects:Array(String)}
  AND baseline_openmeter_events.time >= {from:DateTime}
  AND baseline_openmeter_events.time < {to:DateTime}
GROUP BY subject
SETTINGS optimize_move_to_prewhere = 1, allow_reorder_prewhere_conditions = 1
```

</details>

<details>
<summary><code>sum_total</code></summary>

```sql
SELECT
  tumbleStart(min(baseline_openmeter_events.time), toIntervalMinute(1)) AS windowstart,
  tumbleEnd(max(baseline_openmeter_events.time), toIntervalMinute(1)) AS windowend,
  sum(toDecimal128OrNull(JSON_VALUE(baseline_openmeter_events.data, '$.value'), 19)) AS value
FROM baseline_openmeter_events
WHERE baseline_openmeter_events.namespace = {namespace:String}
  AND baseline_openmeter_events.type = {type:String}
  AND baseline_openmeter_events.subject IN {subjects:Array(String)}
  AND baseline_openmeter_events.time >= {from:DateTime}
  AND baseline_openmeter_events.time < {to:DateTime}
SETTINGS optimize_move_to_prewhere = 1, allow_reorder_prewhere_conditions = 1
```

</details>

<details>
<summary><code>sum_total_by_subject</code></summary>

```sql
SELECT
  tumbleStart(min(baseline_openmeter_events.time), toIntervalMinute(1)) AS windowstart,
  tumbleEnd(max(baseline_openmeter_events.time), toIntervalMinute(1)) AS windowend,
  sum(toDecimal128OrNull(JSON_VALUE(baseline_openmeter_events.data, '$.value'), 19)) AS value,
  baseline_openmeter_events.subject AS subject
FROM baseline_openmeter_events
WHERE baseline_openmeter_events.namespace = {namespace:String}
  AND baseline_openmeter_events.type = {type:String}
  AND baseline_openmeter_events.subject IN {subjects:Array(String)}
  AND baseline_openmeter_events.time >= {from:DateTime}
  AND baseline_openmeter_events.time < {to:DateTime}
GROUP BY subject
SETTINGS optimize_move_to_prewhere = 1, allow_reorder_prewhere_conditions = 1
```

</details>

<details>
<summary><code>sum_total_filter1</code></summary>

```sql
SELECT
  tumbleStart(min(baseline_openmeter_events.time), toIntervalMinute(1)) AS windowstart,
  tumbleEnd(max(baseline_openmeter_events.time), toIntervalMinute(1)) AS windowend,
  sum(toDecimal128OrNull(JSON_VALUE(baseline_openmeter_events.data, '$.value'), 19)) AS value
FROM baseline_openmeter_events
WHERE baseline_openmeter_events.namespace = {namespace:String}
  AND baseline_openmeter_events.type = {type:String}
  AND baseline_openmeter_events.subject IN {subjects:Array(String)}
  AND baseline_openmeter_events.time >= {from:DateTime}
  AND baseline_openmeter_events.time < {to:DateTime}
  AND JSON_VALUE(baseline_openmeter_events.data, '$.group1') = {group1:String}
SETTINGS optimize_move_to_prewhere = 1, allow_reorder_prewhere_conditions = 1
```

</details>

<details>
<summary><code>sum_total_filter2</code></summary>

```sql
SELECT
  tumbleStart(min(baseline_openmeter_events.time), toIntervalMinute(1)) AS windowstart,
  tumbleEnd(max(baseline_openmeter_events.time), toIntervalMinute(1)) AS windowend,
  sum(toDecimal128OrNull(JSON_VALUE(baseline_openmeter_events.data, '$.value'), 19)) AS value
FROM baseline_openmeter_events
WHERE baseline_openmeter_events.namespace = {namespace:String}
  AND baseline_openmeter_events.type = {type:String}
  AND baseline_openmeter_events.subject IN {subjects:Array(String)}
  AND baseline_openmeter_events.time >= {from:DateTime}
  AND baseline_openmeter_events.time < {to:DateTime}
  AND JSON_VALUE(baseline_openmeter_events.data, '$.group1') = {group1:String}
  AND JSON_VALUE(baseline_openmeter_events.data, '$.group2') = {group2:String}
SETTINGS optimize_move_to_prewhere = 1, allow_reorder_prewhere_conditions = 1
```

</details>

<details>
<summary><code>unique_count_hour</code></summary>

```sql
SELECT
  tumbleStart(baseline_openmeter_events.time, toIntervalHour(1), 'UTC') AS windowstart,
  tumbleEnd(baseline_openmeter_events.time, toIntervalHour(1), 'UTC') AS windowend,
  uniqExact(nullIf(JSON_VALUE(baseline_openmeter_events.data, '$.value'), 'null')) AS value,
  baseline_openmeter_events.subject AS subject
FROM baseline_openmeter_events
WHERE baseline_openmeter_events.namespace = {namespace:String}
  AND baseline_openmeter_events.type = {type:String}
  AND baseline_openmeter_events.subject IN {subjects:Array(String)}
  AND baseline_openmeter_events.time >= {from:DateTime}
  AND baseline_openmeter_events.time < {to:DateTime}
GROUP BY windowstart, windowend, subject
ORDER BY windowstart
SETTINGS optimize_move_to_prewhere = 1, allow_reorder_prewhere_conditions = 1
```

</details>

<details>
<summary><code>workload_seconds_by_region</code></summary>

```sql
SELECT
  tumbleStart(baseline_openmeter_events.time, toIntervalHour(1), 'UTC') AS windowstart,
  tumbleEnd(baseline_openmeter_events.time, toIntervalHour(1), 'UTC') AS windowend,
  sum(toDecimal128OrNull(JSON_VALUE(baseline_openmeter_events.data, '$.duration_seconds'), 19)) AS value,
  JSON_VALUE(baseline_openmeter_events.data, '$.region') AS region
FROM baseline_openmeter_events
WHERE baseline_openmeter_events.namespace = {namespace:String}
  AND baseline_openmeter_events.type = 'workload'
  AND baseline_openmeter_events.subject IN {subjects:Array(String)}
  AND baseline_openmeter_events.time >= {from:DateTime}
  AND baseline_openmeter_events.time < {to:DateTime}
GROUP BY windowstart, windowend, region
ORDER BY windowstart
SETTINGS optimize_move_to_prewhere = 1, allow_reorder_prewhere_conditions = 1
```

</details>

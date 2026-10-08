> Target `altinity/local-single` is a local tier: this run backs parity and correctness only, never performance.

# Benchmark: baseline-openmeter

| Field | Value |
| --- | --- |
| Harness commit | `a9d9670` |
| ClickHouse | 26.3.33.10001.altinitystable (single-node) |
| Clusters | default |
| Sweep id | `20261008T083418Z` |
| Concurrency | 1 |
| Started | 2026-10-08T08:34:18Z |
| Finished | 2026-10-08T08:34:38Z |

## Ingest

Source `go-seeder`: 1000000 rows in 9.97s (100342 events/sec), batch=10000.

## Queries

| Query | c | cache | p50 ms | p95 ms | p99 ms | QPS | CPU p50 ms | CPU src | Mem MB |
| --- | ---: | --- | ---: | ---: | ---: | ---: | ---: | --- | ---: |
| agent_runs_by_name | 1 | warm | 29.0 | 41.0 | 41.0 | 14.8 | 32.4 | os_cpu | 15.8 |
| avg_hour | 1 | warm | 39.0 | 45.0 | 45.0 | 12.5 | 60.8 | os_cpu | 19.1 |
| count_hour | 1 | warm | 17.0 | 20.0 | 20.0 | 18.0 | 21.8 | os_cpu | 0.4 |
| count_total | 1 | warm | 14.0 | 16.0 | 16.0 | 18.9 | 13.6 | os_cpu | 3.2 |
| distinct_subjects | 1 | warm | 31.0 | 33.0 | 33.0 | 13.1 | 35.1 | os_cpu | 4.2 |
| kong_api_request_total | 1 | warm | 12.0 | 14.0 | 14.0 | 18.5 | 14.0 | os_cpu | 0.1 |
| kong_llm_tokens_total | 1 | warm | 84.0 | 100.0 | 100.0 | 7.7 | 98.2 | os_cpu | 23.1 |
| kong_status_by_route | 1 | warm | 58.0 | 65.0 | 65.0 | 11.0 | 58.8 | os_cpu | 26.9 |
| latest_hour | 1 | warm | 47.0 | 49.0 | 49.0 | 12.1 | 55.0 | os_cpu | 19.5 |
| llm_tokens_by_model | 1 | warm | 29.0 | 33.0 | 33.0 | 14.9 | 36.2 | os_cpu | 23.9 |
| max_hour | 1 | warm | 33.0 | 42.0 | 42.0 | 14.1 | 42.5 | os_cpu | 20.9 |
| min_hour | 1 | warm | 33.0 | 36.0 | 36.0 | 14.6 | 41.7 | os_cpu | 22.3 |
| sum_day | 1 | warm | 36.0 | 36.0 | 36.0 | 13.9 | 43.2 | os_cpu | 18.6 |
| sum_day_by_dim | 1 | warm | 23.0 | 28.0 | 28.0 | 16.5 | 31.4 | os_cpu | 25.1 |
| sum_day_by_subject | 1 | warm | 50.0 | 60.0 | 60.0 | 11.8 | 44.9 | os_cpu | 17.3 |
| sum_hour | 1 | warm | 38.0 | 39.0 | 39.0 | 13.4 | 48.0 | os_cpu | 17.3 |
| sum_hour_group1 | 1 | warm | 34.0 | 48.0 | 48.0 | 13.3 | 52.0 | os_cpu | 17.7 |
| sum_hour_group1_group2 | 1 | warm | 46.0 | 46.0 | 46.0 | 12.1 | 61.0 | os_cpu | 16.4 |
| sum_hour_group1_group2_no_prewhere | 1 | warm | 49.0 | 50.0 | 50.0 | 12.5 | 61.2 | os_cpu | 17.7 |
| sum_hour_group1_no_prewhere | 1 | warm | 32.0 | 42.0 | 42.0 | 13.9 | 47.5 | os_cpu | 17.7 |
| sum_hour_tz | 1 | warm | 35.0 | 47.0 | 47.0 | 13.2 | 43.1 | os_cpu | 15.9 |
| sum_minute | 1 | warm | 59.0 | 62.0 | 62.0 | 10.7 | 54.4 | os_cpu | 18.3 |
| sum_month | 1 | warm | 43.0 | 53.0 | 53.0 | 12.3 | 48.5 | os_cpu | 18.7 |
| sum_no_window | 1 | warm | 42.0 | 43.0 | 43.0 | 12.8 | 46.6 | os_cpu | 17.2 |
| sum_total | 1 | warm | 37.0 | 53.0 | 53.0 | 12.7 | 43.7 | os_cpu | 18.4 |
| sum_total_by_subject | 1 | warm | 36.0 | 41.0 | 41.0 | 13.4 | 45.4 | os_cpu | 21.0 |
| sum_total_filter1 | 1 | warm | 34.0 | 55.0 | 55.0 | 13.0 | 49.5 | os_cpu | 17.7 |
| sum_total_filter2 | 1 | warm | 42.0 | 45.0 | 45.0 | 12.7 | 48.5 | os_cpu | 17.7 |
| unique_count_hour | 1 | warm | 34.0 | 36.0 | 36.0 | 14.1 | 43.5 | os_cpu | 18.1 |
| workload_seconds_by_region | 1 | warm | 15.0 | 20.0 | 20.0 | 18.5 | 24.4 | os_cpu | 15.7 |

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

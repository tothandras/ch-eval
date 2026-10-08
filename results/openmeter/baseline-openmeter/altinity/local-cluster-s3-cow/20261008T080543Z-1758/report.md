> Target `altinity/local-cluster-s3-cow` is a local tier: this run backs parity and correctness only, never performance.

# Benchmark: baseline-openmeter

| Field | Value |
| --- | --- |
| Harness commit | `baac5b0` |
| ClickHouse | 26.3.33.10001.altinitystable (multi-node) |
| Clusters | ch_eval |
| Sweep id | `20261008T080543Z` |
| Concurrency | 1 |
| Started | 2026-10-08T08:05:43Z |
| Finished | 2026-10-08T08:06:03Z |

## Ingest

Source `go-seeder`: 500000 rows in 10.88s (45938 events/sec), batch=10000.

## Queries

| Query | c | cache | p50 ms | p95 ms | p99 ms | QPS | CPU p50 ms | CPU src | Mem MB |
| --- | ---: | --- | ---: | ---: | ---: | ---: | ---: | --- | ---: |
| agent_runs_by_name | 1 | warm | 16.0 | 16.0 | 16.0 | 13.2 | n/a | n/a | n/a |
| avg_hour | 1 | warm | 29.0 | 29.0 | 29.0 | 12.0 | n/a | n/a | n/a |
| count_hour | 1 | warm | 21.0 | 21.0 | 21.0 | 12.1 | n/a | n/a | n/a |
| count_total | 1 | warm | 23.0 | 23.0 | 23.0 | 12.0 | n/a | n/a | n/a |
| distinct_subjects | 1 | warm | 24.0 | 24.0 | 24.0 | 11.7 | n/a | n/a | n/a |
| kong_api_request_total | 1 | warm | 14.0 | 14.0 | 14.0 | 10.5 | n/a | n/a | n/a |
| kong_llm_tokens_total | 1 | warm | 48.0 | 48.0 | 48.0 | 8.8 | n/a | n/a | n/a |
| kong_status_by_route | 1 | warm | 30.0 | 30.0 | 30.0 | 12.8 | n/a | n/a | n/a |
| latest_hour | 1 | warm | 27.0 | 27.0 | 27.0 | 11.3 | n/a | n/a | n/a |
| llm_tokens_by_model | 1 | warm | 25.0 | 25.0 | 25.0 | 12.3 | n/a | n/a | n/a |
| max_hour | 1 | warm | 34.0 | 34.0 | 34.0 | 11.9 | n/a | n/a | n/a |
| min_hour | 1 | warm | 26.0 | 26.0 | 26.0 | 12.6 | n/a | n/a | n/a |
| sum_day | 1 | warm | 39.0 | 39.0 | 39.0 | 11.8 | n/a | n/a | n/a |
| sum_day_by_dim | 1 | warm | 55.0 | 55.0 | 55.0 | 10.2 | n/a | n/a | n/a |
| sum_day_by_subject | 1 | warm | 33.0 | 33.0 | 33.0 | 12.1 | n/a | n/a | n/a |
| sum_hour | 1 | warm | 27.0 | 27.0 | 27.0 | 11.7 | n/a | n/a | n/a |
| sum_hour_group1 | 1 | warm | 33.0 | 33.0 | 33.0 | 12.2 | n/a | n/a | n/a |
| sum_hour_group1_group2 | 1 | warm | 39.0 | 39.0 | 39.0 | 11.3 | n/a | n/a | n/a |
| sum_hour_group1_group2_no_prewhere | 1 | warm | 32.0 | 32.0 | 32.0 | 12.1 | n/a | n/a | n/a |
| sum_hour_group1_no_prewhere | 1 | warm | 43.0 | 43.0 | 43.0 | 10.8 | n/a | n/a | n/a |
| sum_hour_tz | 1 | warm | 31.0 | 31.0 | 31.0 | 11.3 | n/a | n/a | n/a |
| sum_minute | 1 | warm | 36.0 | 36.0 | 36.0 | 10.9 | n/a | n/a | n/a |
| sum_month | 1 | warm | 36.0 | 36.0 | 36.0 | 11.2 | n/a | n/a | n/a |
| sum_no_window | 1 | warm | 39.0 | 39.0 | 39.0 | 10.6 | n/a | n/a | n/a |
| sum_total | 1 | warm | 53.0 | 53.0 | 53.0 | 10.3 | n/a | n/a | n/a |
| sum_total_by_subject | 1 | warm | 22.0 | 22.0 | 22.0 | 12.9 | n/a | n/a | n/a |
| sum_total_filter1 | 1 | warm | 31.0 | 31.0 | 31.0 | 12.1 | n/a | n/a | n/a |
| sum_total_filter2 | 1 | warm | 32.0 | 32.0 | 32.0 | 12.0 | n/a | n/a | n/a |
| unique_count_hour | 1 | warm | 46.0 | 46.0 | 46.0 | 11.0 | n/a | n/a | n/a |
| workload_seconds_by_region | 1 | warm | 27.0 | 27.0 | 27.0 | 11.6 | n/a | n/a | n/a |

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

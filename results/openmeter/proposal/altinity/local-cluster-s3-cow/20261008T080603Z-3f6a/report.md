> Target `altinity/local-cluster-s3-cow` is a local tier: this run backs parity and correctness only, never performance.

# Benchmark: proposal

| Field | Value |
| --- | --- |
| Harness commit | `baac5b0` |
| ClickHouse | 26.3.33.10001.altinitystable (multi-node) |
| Clusters | ch_eval |
| Sweep id | `20261008T080543Z` |
| Concurrency | 1 |
| Started | 2026-10-08T08:06:03Z |
| Finished | 2026-10-08T08:06:36Z |

## Ingest

Source `go-seeder`: 500000 rows in 19.25s (25976 events/sec), batch=10000.

## Queries

| Query | c | cache | p50 ms | p95 ms | p99 ms | QPS | CPU p50 ms | CPU src | Mem MB |
| --- | ---: | --- | ---: | ---: | ---: | ---: | ---: | --- | ---: |
| agent_runs_by_name | 1 | warm | 32.0 | 32.0 | 32.0 | 11.9 | n/a | n/a | n/a |
| avg_hour | 1 | warm | 108.0 | 108.0 | 108.0 | 6.9 | n/a | n/a | n/a |
| count_hour | 1 | warm | 65.0 | 65.0 | 65.0 | 6.7 | n/a | n/a | n/a |
| count_total | 1 | warm | 35.0 | 35.0 | 35.0 | 11.4 | n/a | n/a | n/a |
| distinct_subjects | 1 | warm | 54.0 | 54.0 | 54.0 | 8.3 | n/a | n/a | n/a |
| kong_api_request_by_all_dims | 1 | warm | 163.0 | 163.0 | 163.0 | 4.6 | n/a | n/a | n/a |
| kong_api_request_by_method | 1 | warm | 35.0 | 35.0 | 35.0 | 11.6 | n/a | n/a | n/a |
| kong_api_request_by_service | 1 | warm | 34.0 | 34.0 | 34.0 | 11.4 | n/a | n/a | n/a |
| kong_api_request_total | 1 | warm | 15.0 | 15.0 | 15.0 | 12.8 | n/a | n/a | n/a |
| kong_llm_tokens_total | 1 | warm | 31.0 | 31.0 | 31.0 | 10.2 | n/a | n/a | n/a |
| kong_status_by_route | 1 | warm | 36.0 | 36.0 | 36.0 | 11.3 | n/a | n/a | n/a |
| latest_hour | 1 | warm | 41.0 | 41.0 | 41.0 | 10.9 | n/a | n/a | n/a |
| llm_tokens_by_model | 1 | warm | 41.0 | 41.0 | 41.0 | 11.1 | n/a | n/a | n/a |
| lookup_by_id | 1 | warm | 118.0 | 118.0 | 118.0 | 6.0 | n/a | n/a | n/a |
| max_hour | 1 | warm | 61.0 | 61.0 | 61.0 | 9.5 | n/a | n/a | n/a |
| min_hour | 1 | warm | 32.0 | 32.0 | 32.0 | 11.8 | n/a | n/a | n/a |
| sum_day | 1 | warm | 34.0 | 34.0 | 34.0 | 12.0 | n/a | n/a | n/a |
| sum_day_by_dim | 1 | warm | 30.0 | 30.0 | 30.0 | 11.5 | n/a | n/a | n/a |
| sum_day_by_subject | 1 | warm | 41.0 | 41.0 | 41.0 | 10.9 | n/a | n/a | n/a |
| sum_hour | 1 | warm | 33.0 | 33.0 | 33.0 | 11.8 | n/a | n/a | n/a |
| sum_hour_group1 | 1 | warm | 52.0 | 52.0 | 52.0 | 9.5 | n/a | n/a | n/a |
| sum_hour_group1_group2 | 1 | warm | 65.0 | 65.0 | 65.0 | 8.5 | n/a | n/a | n/a |
| sum_hour_group1_group2_no_prewhere | 1 | warm | 43.0 | 43.0 | 43.0 | 10.5 | n/a | n/a | n/a |
| sum_hour_group1_no_prewhere | 1 | warm | 79.0 | 79.0 | 79.0 | 8.8 | n/a | n/a | n/a |
| sum_hour_tz | 1 | warm | 66.0 | 66.0 | 66.0 | 9.5 | n/a | n/a | n/a |
| sum_minute | 1 | warm | 48.0 | 48.0 | 48.0 | 9.6 | n/a | n/a | n/a |
| sum_month | 1 | warm | 44.0 | 44.0 | 44.0 | 11.3 | n/a | n/a | n/a |
| sum_no_window | 1 | warm | 45.0 | 45.0 | 45.0 | 10.1 | n/a | n/a | n/a |
| sum_total | 1 | warm | 77.0 | 77.0 | 77.0 | 9.3 | n/a | n/a | n/a |
| sum_total_by_subject | 1 | warm | 45.0 | 45.0 | 45.0 | 10.1 | n/a | n/a | n/a |
| sum_total_filter1 | 1 | warm | 57.0 | 57.0 | 57.0 | 9.4 | n/a | n/a | n/a |
| sum_total_filter2 | 1 | warm | 81.0 | 81.0 | 81.0 | 8.5 | n/a | n/a | n/a |
| unique_count_hour | 1 | warm | 36.0 | 36.0 | 36.0 | 11.5 | n/a | n/a | n/a |
| workload_seconds_by_region | 1 | warm | 36.0 | 36.0 | 36.0 | 11.2 | n/a | n/a | n/a |

### Index pruning (`EXPLAIN indexes=1`, literal id)

| Query | granules (no index → index) | parts (no index → index) |
| --- | ---: | ---: |
| lookup_by_id | 62 → 2 | 8 → 2 |

<details>
<summary><code>agent_runs_by_name</code></summary>

```sql
SELECT
  tumbleStart(proposal_events.time, toIntervalHour(1), 'UTC') AS windowstart,
  tumbleEnd(proposal_events.time, toIntervalHour(1), 'UTC') AS windowend,
  count(*) AS value,
  toString(proposal_events.data.agent_name.:String) AS agent_name
FROM proposal_events
WHERE proposal_events.namespace = {namespace:String}
  AND proposal_events.type = 'agent_run'
  AND proposal_events.subject IN {subjects:Array(String)}
  AND proposal_events.time >= {from:DateTime}
  AND proposal_events.time < {to:DateTime}
GROUP BY windowstart, windowend, agent_name
ORDER BY windowstart
SETTINGS optimize_move_to_prewhere = 1, allow_reorder_prewhere_conditions = 1
```

</details>

<details>
<summary><code>avg_hour</code></summary>

```sql
SELECT
  tumbleStart(proposal_events.time, toIntervalHour(1), 'UTC') AS windowstart,
  tumbleEnd(proposal_events.time, toIntervalHour(1), 'UTC') AS windowend,
  avg(toDecimal128OrNull(toString(proposal_events.data.value), 19)) AS value,
  proposal_events.subject AS subject
FROM proposal_events
WHERE proposal_events.namespace = {namespace:String}
  AND proposal_events.type = {type:String}
  AND proposal_events.subject IN {subjects:Array(String)}
  AND proposal_events.time >= {from:DateTime}
  AND proposal_events.time < {to:DateTime}
GROUP BY windowstart, windowend, subject
ORDER BY windowstart
SETTINGS optimize_move_to_prewhere = 1, allow_reorder_prewhere_conditions = 1
```

</details>

<details>
<summary><code>count_hour</code></summary>

```sql
SELECT
  tumbleStart(proposal_events.time, toIntervalHour(1), 'UTC') AS windowstart,
  tumbleEnd(proposal_events.time, toIntervalHour(1), 'UTC') AS windowend,
  count(*) AS value,
  proposal_events.subject AS subject
FROM proposal_events
WHERE proposal_events.namespace = {namespace:String}
  AND proposal_events.type = {type:String}
  AND proposal_events.subject IN {subjects:Array(String)}
  AND proposal_events.time >= {from:DateTime}
  AND proposal_events.time < {to:DateTime}
GROUP BY windowstart, windowend, subject
ORDER BY windowstart
SETTINGS optimize_move_to_prewhere = 1, allow_reorder_prewhere_conditions = 1
```

</details>

<details>
<summary><code>count_total</code></summary>

```sql
SELECT
  tumbleStart(min(proposal_events.time), toIntervalMinute(1)) AS windowstart,
  tumbleEnd(max(proposal_events.time), toIntervalMinute(1)) AS windowend,
  count(*) AS value
FROM proposal_events
WHERE proposal_events.namespace = {namespace:String}
  AND proposal_events.type = {type:String}
  AND proposal_events.subject IN {subjects:Array(String)}
  AND proposal_events.time >= {from:DateTime}
  AND proposal_events.time < {to:DateTime}
SETTINGS optimize_move_to_prewhere = 1, allow_reorder_prewhere_conditions = 1
```

</details>

<details>
<summary><code>distinct_subjects</code></summary>

```sql
SELECT DISTINCT proposal_events.subject AS subject
FROM proposal_events
WHERE proposal_events.namespace = {namespace:String}
  AND proposal_events.type = {type:String}
  AND proposal_events.time >= {from:DateTime}
  AND proposal_events.time < {to:DateTime}
ORDER BY subject
```

</details>

<details>
<summary><code>kong_api_request_by_all_dims</code></summary>

```sql
SELECT
  tumbleStart(proposal_events.time, toIntervalHour(1), 'UTC') AS windowstart,
  tumbleEnd(proposal_events.time, toIntervalHour(1), 'UTC') AS windowend,
  count(*) AS value,
  toString(proposal_events.data.api_id.:String)                 AS api_id,
  toString(proposal_events.data.api_product_id.:String)         AS api_product_id,
  toString(proposal_events.data.api_product_version_id.:String) AS api_product_version_id,
  toString(proposal_events.data.application_id.:String)         AS application_id,
  toString(proposal_events.data.client_ip.:String)             AS client_ip,
  toString(proposal_events.data.control_plane_id.:String)       AS control_plane_id,
  toString(proposal_events.data.portal_id.:String)             AS portal_id,
  toString(proposal_events.data.request_host.:String)           AS request_host,
  toString(proposal_events.data.request_method.:String)         AS request_method,
  toString(proposal_events.data.request_uri.:String)           AS request_uri,
  toString(proposal_events.data.request_user_agent.:String)     AS request_user_agent,
  toString(proposal_events.data.response_http_status.:String)   AS response_http_status,
  toString(proposal_events.data.route_id.:String)               AS route_id,
  toString(proposal_events.data.route_name.:String)             AS route_name,
  toString(proposal_events.data.service_id.:String)             AS service_id,
  toString(proposal_events.data.service_name.:String)           AS service_name,
  toString(proposal_events.data.service_port.:String)           AS service_port,
  toString(proposal_events.data.service_protocol.:String)       AS service_protocol,
  toString(proposal_events.data.upstream_status.:String)        AS upstream_status
FROM proposal_events
WHERE proposal_events.namespace = {namespace:String}
  AND proposal_events.type = 'kong.api_request'
  AND proposal_events.subject IN {subjects:Array(String)}
  AND proposal_events.time >= {from:DateTime}
  AND proposal_events.time < {to:DateTime}
GROUP BY
  windowstart, windowend,
  api_id, api_product_id, api_product_version_id, application_id, client_ip,
  control_plane_id, portal_id, request_host, request_method, request_uri,
  request_user_agent, response_http_status, route_id, route_name, service_id,
  service_name, service_port, service_protocol, upstream_status
ORDER BY windowstart
SETTINGS optimize_move_to_prewhere = 1, allow_reorder_prewhere_conditions = 1
```

</details>

<details>
<summary><code>kong_api_request_by_method</code></summary>

```sql
SELECT
  tumbleStart(proposal_events.time, toIntervalHour(1), 'UTC') AS windowstart,
  tumbleEnd(proposal_events.time, toIntervalHour(1), 'UTC') AS windowend,
  count(*) AS value,
  toString(proposal_events.data.request_method.:String) AS request_method
FROM proposal_events
WHERE proposal_events.namespace = {namespace:String}
  AND proposal_events.type = 'kong.api_request'
  AND proposal_events.subject IN {subjects:Array(String)}
  AND proposal_events.time >= {from:DateTime}
  AND proposal_events.time < {to:DateTime}
GROUP BY windowstart, windowend, request_method
ORDER BY windowstart
SETTINGS optimize_move_to_prewhere = 1, allow_reorder_prewhere_conditions = 1
```

</details>

<details>
<summary><code>kong_api_request_by_service</code></summary>

```sql
SELECT
  tumbleStart(proposal_events.time, toIntervalHour(1), 'UTC') AS windowstart,
  tumbleEnd(proposal_events.time, toIntervalHour(1), 'UTC') AS windowend,
  count(*) AS value,
  toString(proposal_events.data.service_name.:String) AS service_name,
  toString(proposal_events.data.request_method.:String) AS request_method
FROM proposal_events
WHERE proposal_events.namespace = {namespace:String}
  AND proposal_events.type = 'kong.api_request'
  AND proposal_events.subject IN {subjects:Array(String)}
  AND proposal_events.time >= {from:DateTime}
  AND proposal_events.time < {to:DateTime}
GROUP BY windowstart, windowend, service_name, request_method
ORDER BY windowstart
SETTINGS optimize_move_to_prewhere = 1, allow_reorder_prewhere_conditions = 1
```

</details>

<details>
<summary><code>kong_api_request_total</code></summary>

```sql
SELECT count(*) AS value
FROM proposal_events
WHERE proposal_events.namespace = {namespace:String}
  AND proposal_events.type = 'kong.api_request'
  AND proposal_events.time >= {from:DateTime}
  AND proposal_events.time < {to:DateTime}
SETTINGS optimize_move_to_prewhere = 1, allow_reorder_prewhere_conditions = 1
```

</details>

<details>
<summary><code>kong_llm_tokens_total</code></summary>

```sql
SELECT sum(toDecimal128OrNull(toString(proposal_events.data.tokens), 19)) AS value
FROM proposal_events
WHERE proposal_events.namespace = {namespace:String}
  AND proposal_events.type = 'kong.llm_request'
  AND proposal_events.time >= {from:DateTime}
  AND proposal_events.time < {to:DateTime}
SETTINGS optimize_move_to_prewhere = 1, allow_reorder_prewhere_conditions = 1
```

</details>

<details>
<summary><code>kong_status_by_route</code></summary>

```sql
SELECT
  tumbleStart(proposal_events.time, toIntervalHour(1), 'UTC') AS windowstart,
  tumbleEnd(proposal_events.time, toIntervalHour(1), 'UTC') AS windowend,
  count(*) AS value,
  toString(proposal_events.data.response_http_status.:String) AS status,
  toString(proposal_events.data.route_name.:String) AS route
FROM proposal_events
WHERE proposal_events.namespace = {namespace:String}
  AND proposal_events.type = 'kong.api_request'
  AND proposal_events.subject IN {subjects:Array(String)}
  AND proposal_events.time >= {from:DateTime}
  AND proposal_events.time < {to:DateTime}
GROUP BY windowstart, windowend, status, route
ORDER BY windowstart
SETTINGS optimize_move_to_prewhere = 1, allow_reorder_prewhere_conditions = 1
```

</details>

<details>
<summary><code>latest_hour</code></summary>

```sql
SELECT
  tumbleStart(proposal_events.time, toIntervalHour(1), 'UTC') AS windowstart,
  tumbleEnd(proposal_events.time, toIntervalHour(1), 'UTC') AS windowend,
  argMax(toDecimal128OrNull(toString(proposal_events.data.value), 19), (proposal_events.time, proposal_events.store_row_id)) AS value,
  proposal_events.subject AS subject
FROM proposal_events
WHERE proposal_events.namespace = {namespace:String}
  AND proposal_events.type = {type:String}
  AND proposal_events.subject IN {subjects:Array(String)}
  AND proposal_events.time >= {from:DateTime}
  AND proposal_events.time < {to:DateTime}
GROUP BY windowstart, windowend, subject
ORDER BY windowstart
SETTINGS optimize_move_to_prewhere = 1, allow_reorder_prewhere_conditions = 1
```

</details>

<details>
<summary><code>llm_tokens_by_model</code></summary>

```sql
SELECT
  tumbleStart(proposal_events.time, toIntervalHour(1), 'UTC') AS windowstart,
  tumbleEnd(proposal_events.time, toIntervalHour(1), 'UTC') AS windowend,
  sum(toDecimal128OrNull(toString(proposal_events.data.tokens), 19)) AS value,
  toString(proposal_events.data.model.:String) AS model,
  toString(proposal_events.data.provider.:String) AS provider
FROM proposal_events
WHERE proposal_events.namespace = {namespace:String}
  AND proposal_events.type = 'kong.llm_request'
  AND proposal_events.subject IN {subjects:Array(String)}
  AND proposal_events.time >= {from:DateTime}
  AND proposal_events.time < {to:DateTime}
GROUP BY windowstart, windowend, model, provider
ORDER BY windowstart
SETTINGS optimize_move_to_prewhere = 1, allow_reorder_prewhere_conditions = 1
```

</details>

<details>
<summary><code>lookup_by_id</code></summary>

```sql
SELECT
  proposal_events.id AS id,
  proposal_events.type AS type,
  proposal_events.subject AS subject,
  proposal_events.source AS source,
  proposal_events.time AS time,
  proposal_events.data AS data,
  proposal_events.ingested_at AS ingested_at,
  proposal_events.stored_at AS stored_at,
  proposal_events.store_row_id AS store_row_id
FROM proposal_events
WHERE proposal_events.namespace = {namespace:String}
  AND proposal_events.id = (
    SELECT id FROM proposal_events
    WHERE namespace = {namespace:String}
    ORDER BY namespace, type, subject, time
    LIMIT 1 OFFSET 100000
  )
SETTINGS optimize_move_to_prewhere = 1, allow_reorder_prewhere_conditions = 1
```

</details>

<details>
<summary><code>max_hour</code></summary>

```sql
SELECT
  tumbleStart(proposal_events.time, toIntervalHour(1), 'UTC') AS windowstart,
  tumbleEnd(proposal_events.time, toIntervalHour(1), 'UTC') AS windowend,
  max(toDecimal128OrNull(toString(proposal_events.data.value), 19)) AS value,
  proposal_events.subject AS subject
FROM proposal_events
WHERE proposal_events.namespace = {namespace:String}
  AND proposal_events.type = {type:String}
  AND proposal_events.subject IN {subjects:Array(String)}
  AND proposal_events.time >= {from:DateTime}
  AND proposal_events.time < {to:DateTime}
GROUP BY windowstart, windowend, subject
ORDER BY windowstart
SETTINGS optimize_move_to_prewhere = 1, allow_reorder_prewhere_conditions = 1
```

</details>

<details>
<summary><code>min_hour</code></summary>

```sql
SELECT
  tumbleStart(proposal_events.time, toIntervalHour(1), 'UTC') AS windowstart,
  tumbleEnd(proposal_events.time, toIntervalHour(1), 'UTC') AS windowend,
  min(toDecimal128OrNull(toString(proposal_events.data.value), 19)) AS value,
  proposal_events.subject AS subject
FROM proposal_events
WHERE proposal_events.namespace = {namespace:String}
  AND proposal_events.type = {type:String}
  AND proposal_events.subject IN {subjects:Array(String)}
  AND proposal_events.time >= {from:DateTime}
  AND proposal_events.time < {to:DateTime}
GROUP BY windowstart, windowend, subject
ORDER BY windowstart
SETTINGS optimize_move_to_prewhere = 1, allow_reorder_prewhere_conditions = 1
```

</details>

<details>
<summary><code>sum_day</code></summary>

```sql
SELECT
  tumbleStart(proposal_events.time, toIntervalDay(1), 'UTC') AS windowstart,
  tumbleEnd(proposal_events.time, toIntervalDay(1), 'UTC') AS windowend,
  sum(toDecimal128OrNull(toString(proposal_events.data.value), 19)) AS value,
  proposal_events.subject AS subject
FROM proposal_events
WHERE proposal_events.namespace = {namespace:String}
  AND proposal_events.type = {type:String}
  AND proposal_events.subject IN {subjects:Array(String)}
  AND proposal_events.time >= {from:DateTime}
  AND proposal_events.time < {to:DateTime}
GROUP BY windowstart, windowend, subject
ORDER BY windowstart
SETTINGS optimize_move_to_prewhere = 1, allow_reorder_prewhere_conditions = 1
```

</details>

<details>
<summary><code>sum_day_by_dim</code></summary>

```sql
SELECT
  tumbleStart(proposal_events.time, toIntervalDay(1), 'UTC') AS windowstart,
  windowstart + toIntervalDay(1) AS windowend,
  sum(toDecimal128OrNull(toString(proposal_events.data.tokens), 19)) AS value,
  toString(proposal_events.data.model) AS model_id,
  proposal_events.subject AS subject
FROM proposal_events
WHERE proposal_events.namespace = {namespace:String}
  AND proposal_events.type = 'kong.llm_request'
  AND proposal_events.subject IN {subjects:Array(String)}
  AND proposal_events.time >= {from:DateTime}
  AND proposal_events.time < {to:DateTime}
GROUP BY windowstart, windowend, model_id, subject
ORDER BY windowstart
SETTINGS optimize_move_to_prewhere = 1, allow_reorder_prewhere_conditions = 1
```

</details>

<details>
<summary><code>sum_day_by_subject</code></summary>

```sql
SELECT
  tumbleStart(proposal_events.time, toIntervalDay(1), 'UTC') AS windowstart,
  windowstart + toIntervalDay(1) AS windowend,
  sum(toDecimal128OrNull(toString(proposal_events.data.value), 19)) AS value,
  proposal_events.subject AS subject
FROM proposal_events
WHERE proposal_events.namespace = {namespace:String}
  AND proposal_events.type = {type:String}
  AND proposal_events.subject IN {subjects:Array(String)}
  AND proposal_events.time >= {from:DateTime}
  AND proposal_events.time < {to:DateTime}
GROUP BY windowstart, windowend, subject
ORDER BY windowstart
SETTINGS optimize_move_to_prewhere = 1, allow_reorder_prewhere_conditions = 1
```

</details>

<details>
<summary><code>sum_hour</code></summary>

```sql
SELECT
  tumbleStart(proposal_events.time, toIntervalHour(1), 'UTC') AS windowstart,
  tumbleEnd(proposal_events.time, toIntervalHour(1), 'UTC') AS windowend,
  sum(toDecimal128OrNull(toString(proposal_events.data.value), 19)) AS value,
  proposal_events.subject AS subject
FROM proposal_events
WHERE proposal_events.namespace = {namespace:String}
  AND proposal_events.type = {type:String}
  AND proposal_events.subject IN {subjects:Array(String)}
  AND proposal_events.time >= {from:DateTime}
  AND proposal_events.time < {to:DateTime}
GROUP BY windowstart, windowend, subject
ORDER BY windowstart
SETTINGS optimize_move_to_prewhere = 1, allow_reorder_prewhere_conditions = 1
```

</details>

<details>
<summary><code>sum_hour_group1</code></summary>

```sql
SELECT
  tumbleStart(proposal_events.time, toIntervalHour(1), 'UTC') AS windowstart,
  tumbleEnd(proposal_events.time, toIntervalHour(1), 'UTC') AS windowend,
  sum(toDecimal128OrNull(toString(proposal_events.data.value), 19)) AS value,
  proposal_events.subject AS subject,
  toString(proposal_events.data.group1.:String) AS group1
FROM proposal_events
WHERE proposal_events.namespace = {namespace:String}
  AND proposal_events.type = {type:String}
  AND proposal_events.subject IN {subjects:Array(String)}
  AND proposal_events.time >= {from:DateTime}
  AND proposal_events.time < {to:DateTime}
  AND toString(proposal_events.data.group1.:String) = {group1:String}
GROUP BY windowstart, windowend, subject, group1
ORDER BY windowstart
SETTINGS optimize_move_to_prewhere = 1, allow_reorder_prewhere_conditions = 1
```

</details>

<details>
<summary><code>sum_hour_group1_group2</code></summary>

```sql
SELECT
  tumbleStart(proposal_events.time, toIntervalHour(1), 'UTC') AS windowstart,
  tumbleEnd(proposal_events.time, toIntervalHour(1), 'UTC') AS windowend,
  sum(toDecimal128OrNull(toString(proposal_events.data.value), 19)) AS value,
  proposal_events.subject AS subject,
  toString(proposal_events.data.group1.:String) AS group1,
  toString(proposal_events.data.group2.:String) AS group2
FROM proposal_events
WHERE proposal_events.namespace = {namespace:String}
  AND proposal_events.type = {type:String}
  AND proposal_events.subject IN {subjects:Array(String)}
  AND proposal_events.time >= {from:DateTime}
  AND proposal_events.time < {to:DateTime}
  AND toString(proposal_events.data.group1.:String) = {group1:String}
  AND toString(proposal_events.data.group2.:String) = {group2:String}
GROUP BY windowstart, windowend, subject, group1, group2
ORDER BY windowstart
SETTINGS optimize_move_to_prewhere = 1, allow_reorder_prewhere_conditions = 1
```

</details>

<details>
<summary><code>sum_hour_group1_group2_no_prewhere</code></summary>

```sql
SELECT
  tumbleStart(proposal_events.time, toIntervalHour(1), 'UTC') AS windowstart,
  tumbleEnd(proposal_events.time, toIntervalHour(1), 'UTC') AS windowend,
  sum(toDecimal128OrNull(toString(proposal_events.data.value), 19)) AS value,
  proposal_events.subject AS subject,
  toString(proposal_events.data.group1.:String) AS group1,
  toString(proposal_events.data.group2.:String) AS group2
FROM proposal_events
WHERE proposal_events.namespace = {namespace:String}
  AND proposal_events.type = {type:String}
  AND proposal_events.subject IN {subjects:Array(String)}
  AND proposal_events.time >= {from:DateTime}
  AND proposal_events.time < {to:DateTime}
  AND toString(proposal_events.data.group1.:String) = {group1:String}
  AND toString(proposal_events.data.group2.:String) = {group2:String}
GROUP BY windowstart, windowend, subject, group1, group2
ORDER BY windowstart
```

</details>

<details>
<summary><code>sum_hour_group1_no_prewhere</code></summary>

```sql
SELECT
  tumbleStart(proposal_events.time, toIntervalHour(1), 'UTC') AS windowstart,
  tumbleEnd(proposal_events.time, toIntervalHour(1), 'UTC') AS windowend,
  sum(toDecimal128OrNull(toString(proposal_events.data.value), 19)) AS value,
  proposal_events.subject AS subject,
  toString(proposal_events.data.group1.:String) AS group1
FROM proposal_events
WHERE proposal_events.namespace = {namespace:String}
  AND proposal_events.type = {type:String}
  AND proposal_events.subject IN {subjects:Array(String)}
  AND proposal_events.time >= {from:DateTime}
  AND proposal_events.time < {to:DateTime}
  AND toString(proposal_events.data.group1.:String) = {group1:String}
GROUP BY windowstart, windowend, subject, group1
ORDER BY windowstart
```

</details>

<details>
<summary><code>sum_hour_tz</code></summary>

```sql
SELECT
  tumbleStart(proposal_events.time, toIntervalHour(1), 'America/New_York') AS windowstart,
  tumbleEnd(proposal_events.time, toIntervalHour(1), 'America/New_York') AS windowend,
  sum(toDecimal128OrNull(toString(proposal_events.data.value), 19)) AS value,
  proposal_events.subject AS subject
FROM proposal_events
WHERE proposal_events.namespace = {namespace:String}
  AND proposal_events.type = {type:String}
  AND proposal_events.subject IN {subjects:Array(String)}
  AND proposal_events.time >= {from:DateTime}
  AND proposal_events.time < {to:DateTime}
GROUP BY windowstart, windowend, subject
ORDER BY windowstart
SETTINGS optimize_move_to_prewhere = 1, allow_reorder_prewhere_conditions = 1
```

</details>

<details>
<summary><code>sum_minute</code></summary>

```sql
SELECT
  tumbleStart(proposal_events.time, toIntervalMinute(1), 'UTC') AS windowstart,
  tumbleEnd(proposal_events.time, toIntervalMinute(1), 'UTC') AS windowend,
  sum(toDecimal128OrNull(toString(proposal_events.data.value), 19)) AS value,
  proposal_events.subject AS subject
FROM proposal_events
WHERE proposal_events.namespace = {namespace:String}
  AND proposal_events.type = {type:String}
  AND proposal_events.subject IN {subjects:Array(String)}
  AND proposal_events.time >= {from:DateTime}
  AND proposal_events.time < {to:DateTime}
GROUP BY windowstart, windowend, subject
ORDER BY windowstart
SETTINGS optimize_move_to_prewhere = 1, allow_reorder_prewhere_conditions = 1
```

</details>

<details>
<summary><code>sum_month</code></summary>

```sql
SELECT
  toDateTime(tumbleStart(proposal_events.time, toIntervalMonth(1), 'UTC'), 'UTC') AS windowstart,
  toDateTime(tumbleEnd(proposal_events.time, toIntervalMonth(1), 'UTC'), 'UTC') AS windowend,
  sum(toDecimal128OrNull(toString(proposal_events.data.value), 19)) AS value,
  proposal_events.subject AS subject
FROM proposal_events
WHERE proposal_events.namespace = {namespace:String}
  AND proposal_events.type = {type:String}
  AND proposal_events.subject IN {subjects:Array(String)}
  AND proposal_events.time >= {from:DateTime}
  AND proposal_events.time < {to:DateTime}
GROUP BY windowstart, windowend, subject
ORDER BY windowstart
SETTINGS optimize_move_to_prewhere = 1, allow_reorder_prewhere_conditions = 1
```

</details>

<details>
<summary><code>sum_no_window</code></summary>

```sql
SELECT
  tumbleStart(min(proposal_events.time), toIntervalMinute(1), 'UTC') AS windowstart,
  tumbleEnd(max(proposal_events.time), toIntervalMinute(1), 'UTC') AS windowend,
  sum(toDecimal128OrNull(toString(proposal_events.data.value), 19)) AS value,
  proposal_events.subject AS subject
FROM proposal_events
WHERE proposal_events.namespace = {namespace:String}
  AND proposal_events.type = {type:String}
  AND proposal_events.subject IN {subjects:Array(String)}
  AND proposal_events.time >= {from:DateTime}
  AND proposal_events.time < {to:DateTime}
GROUP BY subject
SETTINGS optimize_move_to_prewhere = 1, allow_reorder_prewhere_conditions = 1
```

</details>

<details>
<summary><code>sum_total</code></summary>

```sql
SELECT
  tumbleStart(min(proposal_events.time), toIntervalMinute(1)) AS windowstart,
  tumbleEnd(max(proposal_events.time), toIntervalMinute(1)) AS windowend,
  sum(toDecimal128OrNull(toString(proposal_events.data.value), 19)) AS value
FROM proposal_events
WHERE proposal_events.namespace = {namespace:String}
  AND proposal_events.type = {type:String}
  AND proposal_events.subject IN {subjects:Array(String)}
  AND proposal_events.time >= {from:DateTime}
  AND proposal_events.time < {to:DateTime}
SETTINGS optimize_move_to_prewhere = 1, allow_reorder_prewhere_conditions = 1
```

</details>

<details>
<summary><code>sum_total_by_subject</code></summary>

```sql
SELECT
  tumbleStart(min(proposal_events.time), toIntervalMinute(1)) AS windowstart,
  tumbleEnd(max(proposal_events.time), toIntervalMinute(1)) AS windowend,
  sum(toDecimal128OrNull(toString(proposal_events.data.value), 19)) AS value,
  proposal_events.subject AS subject
FROM proposal_events
WHERE proposal_events.namespace = {namespace:String}
  AND proposal_events.type = {type:String}
  AND proposal_events.subject IN {subjects:Array(String)}
  AND proposal_events.time >= {from:DateTime}
  AND proposal_events.time < {to:DateTime}
GROUP BY subject
SETTINGS optimize_move_to_prewhere = 1, allow_reorder_prewhere_conditions = 1
```

</details>

<details>
<summary><code>sum_total_filter1</code></summary>

```sql
SELECT
  tumbleStart(min(proposal_events.time), toIntervalMinute(1)) AS windowstart,
  tumbleEnd(max(proposal_events.time), toIntervalMinute(1)) AS windowend,
  sum(toDecimal128OrNull(toString(proposal_events.data.value), 19)) AS value
FROM proposal_events
WHERE proposal_events.namespace = {namespace:String}
  AND proposal_events.type = {type:String}
  AND proposal_events.subject IN {subjects:Array(String)}
  AND proposal_events.time >= {from:DateTime}
  AND proposal_events.time < {to:DateTime}
  AND toString(proposal_events.data.group1) = {group1:String}
SETTINGS optimize_move_to_prewhere = 1, allow_reorder_prewhere_conditions = 1
```

</details>

<details>
<summary><code>sum_total_filter2</code></summary>

```sql
SELECT
  tumbleStart(min(proposal_events.time), toIntervalMinute(1)) AS windowstart,
  tumbleEnd(max(proposal_events.time), toIntervalMinute(1)) AS windowend,
  sum(toDecimal128OrNull(toString(proposal_events.data.value), 19)) AS value
FROM proposal_events
WHERE proposal_events.namespace = {namespace:String}
  AND proposal_events.type = {type:String}
  AND proposal_events.subject IN {subjects:Array(String)}
  AND proposal_events.time >= {from:DateTime}
  AND proposal_events.time < {to:DateTime}
  AND toString(proposal_events.data.group1) = {group1:String}
  AND toString(proposal_events.data.group2) = {group2:String}
SETTINGS optimize_move_to_prewhere = 1, allow_reorder_prewhere_conditions = 1
```

</details>

<details>
<summary><code>unique_count_hour</code></summary>

```sql
SELECT
  tumbleStart(proposal_events.time, toIntervalHour(1), 'UTC') AS windowstart,
  tumbleEnd(proposal_events.time, toIntervalHour(1), 'UTC') AS windowend,
  uniqExact(nullIf(toString(proposal_events.data.value), 'null')) AS value,
  proposal_events.subject AS subject
FROM proposal_events
WHERE proposal_events.namespace = {namespace:String}
  AND proposal_events.type = {type:String}
  AND proposal_events.subject IN {subjects:Array(String)}
  AND proposal_events.time >= {from:DateTime}
  AND proposal_events.time < {to:DateTime}
GROUP BY windowstart, windowend, subject
ORDER BY windowstart
SETTINGS optimize_move_to_prewhere = 1, allow_reorder_prewhere_conditions = 1
```

</details>

<details>
<summary><code>workload_seconds_by_region</code></summary>

```sql
SELECT
  tumbleStart(proposal_events.time, toIntervalHour(1), 'UTC') AS windowstart,
  tumbleEnd(proposal_events.time, toIntervalHour(1), 'UTC') AS windowend,
  sum(toDecimal128OrNull(toString(proposal_events.data.duration_seconds), 19)) AS value,
  toString(proposal_events.data.region.:String) AS region
FROM proposal_events
WHERE proposal_events.namespace = {namespace:String}
  AND proposal_events.type = 'workload'
  AND proposal_events.subject IN {subjects:Array(String)}
  AND proposal_events.time >= {from:DateTime}
  AND proposal_events.time < {to:DateTime}
GROUP BY windowstart, windowend, region
ORDER BY windowstart
SETTINGS optimize_move_to_prewhere = 1, allow_reorder_prewhere_conditions = 1
```

</details>

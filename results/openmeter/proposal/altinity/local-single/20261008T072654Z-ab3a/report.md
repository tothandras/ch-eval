> Target `altinity/local-single` is a local tier: this run backs parity and correctness only, never performance.

# Benchmark: proposal

| Field | Value |
| --- | --- |
| Harness commit | `d75d88e` |
| ClickHouse | 26.3.33.10001.altinitystable (single-node) |
| Clusters | default |
| Sweep id | `20261008T072635Z` |
| Concurrency | 1 |
| Started | 2026-10-08T07:26:54Z |
| Finished | 2026-10-08T07:27:20Z |

## Ingest

Source `go-seeder`: 1000000 rows in 14.37s (69575 events/sec), batch=10000.

## Queries

| Query | c | cache | p50 ms | p95 ms | p99 ms | QPS | CPU p50 ms | CPU src | Mem MB |
| --- | ---: | --- | ---: | ---: | ---: | ---: | ---: | --- | ---: |
| agent_runs_by_name | 1 | warm | 12.0 | 15.0 | 15.0 | 16.2 | 19.5 | os_cpu | 0.4 |
| avg_hour | 1 | warm | 23.0 | 26.0 | 26.0 | 16.9 | 41.1 | os_cpu | 4.2 |
| count_hour | 1 | warm | 12.0 | 14.0 | 14.0 | 18.7 | 20.0 | os_cpu | 0.4 |
| count_total | 1 | warm | 11.0 | 12.0 | 12.0 | 19.3 | 15.2 | os_cpu | 0.1 |
| distinct_subjects | 1 | warm | 25.0 | 26.0 | 26.0 | 14.9 | 34.7 | os_cpu | 3.0 |
| kong_api_request_by_all_dims | 1 | warm | 82.0 | 117.0 | 117.0 | 7.4 | 128.8 | os_cpu | 46.0 |
| kong_api_request_by_method | 1 | warm | 18.0 | 25.0 | 25.0 | 17.8 | 31.4 | os_cpu | 0.1 |
| kong_api_request_by_service | 1 | warm | 16.0 | 20.0 | 20.0 | 18.9 | 30.1 | os_cpu | 4.3 |
| kong_api_request_total | 1 | warm | 8.0 | 9.0 | 9.0 | 20.4 | 13.0 | os_cpu | 0.1 |
| kong_llm_tokens_total | 1 | warm | 23.0 | 24.0 | 24.0 | 15.9 | 37.5 | os_cpu | 4.1 |
| kong_status_by_route | 1 | warm | 19.0 | 21.0 | 21.0 | 17.6 | 34.8 | os_cpu | 4.1 |
| latest_hour | 1 | warm | 39.0 | 45.0 | 45.0 | 13.5 | 58.1 | os_cpu | 4.0 |
| llm_tokens_by_model | 1 | warm | 25.0 | 31.0 | 31.0 | 14.8 | 33.0 | os_cpu | 4.1 |
| lookup_by_id | 1 | warm | 118.0 | 124.0 | 124.0 | 7.1 | 119.2 | os_cpu | 54.4 |
| max_hour | 1 | warm | 40.0 | 43.0 | 43.0 | 12.9 | 40.2 | os_cpu | 4.2 |
| min_hour | 1 | warm | 35.0 | 40.0 | 40.0 | 13.9 | 42.4 | os_cpu | 4.2 |
| sum_day | 1 | warm | 37.0 | 45.0 | 45.0 | 13.3 | 47.0 | os_cpu | 4.2 |
| sum_day_by_dim | 1 | warm | 27.0 | 27.0 | 27.0 | 15.8 | 35.6 | os_cpu | 4.1 |
| sum_day_by_subject | 1 | warm | 30.0 | 34.0 | 34.0 | 14.7 | 44.7 | os_cpu | 4.2 |
| sum_hour | 1 | warm | 27.0 | 28.0 | 28.0 | 15.8 | 35.9 | os_cpu | 4.2 |
| sum_hour_group1 | 1 | warm | 44.0 | 47.0 | 47.0 | 12.1 | 45.8 | os_cpu | 13.5 |
| sum_hour_group1_group2 | 1 | warm | 51.0 | 58.0 | 58.0 | 11.4 | 50.9 | os_cpu | 15.2 |
| sum_hour_group1_group2_no_prewhere | 1 | warm | 37.0 | 49.0 | 49.0 | 13.2 | 49.8 | os_cpu | 16.6 |
| sum_hour_group1_no_prewhere | 1 | warm | 43.0 | 62.0 | 62.0 | 11.8 | 56.5 | os_cpu | 16.2 |
| sum_hour_tz | 1 | warm | 28.0 | 48.0 | 48.0 | 14.2 | 38.0 | os_cpu | 4.2 |
| sum_minute | 1 | warm | 42.0 | 46.0 | 46.0 | 12.4 | 58.6 | os_cpu | 9.9 |
| sum_month | 1 | warm | 40.0 | 53.0 | 53.0 | 13.2 | 41.9 | os_cpu | 4.2 |
| sum_no_window | 1 | warm | 37.0 | 58.0 | 58.0 | 12.9 | 48.4 | os_cpu | 4.2 |
| sum_total | 1 | warm | 35.0 | 50.0 | 50.0 | 13.9 | 43.5 | os_cpu | 0.1 |
| sum_total_by_subject | 1 | warm | 29.0 | 53.0 | 53.0 | 14.3 | 39.9 | os_cpu | 4.2 |
| sum_total_filter1 | 1 | warm | 28.0 | 54.0 | 54.0 | 13.7 | 36.7 | os_cpu | 0.1 |
| sum_total_filter2 | 1 | warm | 36.0 | 41.0 | 41.0 | 13.9 | 53.4 | os_cpu | 4.0 |
| unique_count_hour | 1 | warm | 32.0 | 32.0 | 32.0 | 15.3 | 45.0 | os_cpu | 8.1 |
| workload_seconds_by_region | 1 | warm | 22.0 | 24.0 | 24.0 | 16.9 | 32.0 | os_cpu | 0.2 |

### Index pruning (`EXPLAIN indexes=1`, literal id)

| Query | granules (no index → index) | parts (no index → index) |
| --- | ---: | ---: |
| lookup_by_id | 122 → 5 | 13 → 4 |

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

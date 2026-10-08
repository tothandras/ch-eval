# Local cross-target parity (2026-10-08)

Smoke and parity only, never performance. Targets: altinity/local-single altinity/local-cluster-2x2 altinity/kind. Window ends 2026-06-01T00:00:00Z; seed 42;
openmeter 1000000 rows, OTel 500000 rows per scenario. Each section is `bench compare` output.

---

## Compare openmeter/baseline-openmeter

> **Smoke, not performance.** At least one side is a local tier, or the sides are on different tiers. Latency and CPU deltas below are not evidence about the platforms.

| Side | Target | Version | Topology | Run | Commit |
| --- | --- | --- | --- | --- | --- |
| altinity/local-single | altinity/local-single (local-single) | 26.3.33.10001.altinitystable | single 1x1 local | 20261008T075113Z-09ac | 3701df2 |
| altinity/local-cluster-2x2 | altinity/local-cluster-2x2 (local-cluster) | 26.3.33.10001.altinitystable | replicated 2x2 local | 20261008T075359Z-96ae | 3701df2 |
| altinity/kind | altinity/kind (kind) | 26.3.33.10001.altinitystable | replicated 1x2 local | 20261008T075856Z-79a8 | 3701df2 |

## altinity/local-single → altinity/local-cluster-2x2

| Query | p50 Δ | CPU Δ | base p50 ms | cand p50 ms |
| --- | ---: | ---: | ---: | ---: |
| agent_runs_by_name | +41% | +23% | 17.0 | 24.0 |
| avg_hour | +95% | +33% | 41.0 | 80.0 |
| count_hour | +50% | +70% | 16.0 | 24.0 |
| count_total | +47% | +94% | 15.0 | 22.0 |
| distinct_subjects | +19% | +38% | 26.0 | 31.0 |
| kong_api_request_total | +19% | +66% | 16.0 | 19.0 |
| kong_llm_tokens_total | +5% | +7% | 82.0 | 86.0 |
| kong_status_by_route | +24% | +20% | 34.0 | 42.0 |
| latest_hour | +4% | +23% | 49.0 | 51.0 |
| llm_tokens_by_model | +42% | +52% | 26.0 | 37.0 |
| max_hour | +13% | +6% | 39.0 | 44.0 |
| min_hour | +52% | +25% | 46.0 | 70.0 |
| sum_day | +19% | +42% | 36.0 | 43.0 |
| sum_day_by_dim | +3% | +20% | 32.0 | 33.0 |
| sum_day_by_subject | -2% | +29% | 51.0 | 50.0 |
| sum_hour | +2% | +5% | 49.0 | 50.0 |
| sum_hour_group1 | -4% | +5% | 46.0 | 44.0 |
| sum_hour_group1_group2 | +18% | +30% | 40.0 | 47.0 |
| sum_hour_group1_group2_no_prewhere | +2% | +24% | 48.0 | 49.0 |
| sum_hour_group1_no_prewhere | +5% | +38% | 40.0 | 42.0 |
| sum_hour_tz | +16% | +24% | 37.0 | 43.0 |
| sum_minute | +9% | +22% | 44.0 | 48.0 |
| sum_month | +11% | +15% | 35.0 | 39.0 |
| sum_no_window | +53% | +124% | 36.0 | 55.0 |
| sum_total | +10% | +18% | 39.0 | 43.0 |
| sum_total_by_subject | +47% | +29% | 40.0 | 59.0 |
| sum_total_filter1 | +3% | +28% | 38.0 | 39.0 |
| sum_total_filter2 | +3% | +41% | 40.0 | 41.0 |
| unique_count_hour | +0% | +36% | 53.0 | 53.0 |
| workload_seconds_by_region | +79% | +50% | 19.0 | 34.0 |

Ingest: base 98128 events/s, candidate 49938 events/s (Δ -49%)

## Value parity (window 2026-05-29 00:00:00 → 2026-06-01 00:00:00)

| Query | result | rows |
| --- | --- | ---: |
| agent_runs_by_name | MATCH | 288 |
| avg_hour | MATCH | 720 |
| count_hour | MATCH | 720 |
| count_total | MATCH | 1 |
| distinct_subjects | MATCH | 100 |
| kong_api_request_total | MATCH | 1 |
| kong_llm_tokens_total | MATCH | 1 |
| kong_status_by_route | MATCH | 6703 |
| latest_hour | MATCH | 720 |
| llm_tokens_by_model | MATCH | 1080 |
| max_hour | MATCH | 720 |
| min_hour | MATCH | 720 |
| sum_day | MATCH | 30 |
| sum_day_by_dim | MATCH | 150 |
| sum_day_by_subject | MATCH | 30 |
| sum_hour | MATCH | 720 |
| sum_hour_group1 | MATCH | 720 |
| sum_hour_group1_group2 | MATCH | 720 |
| sum_hour_group1_group2_no_prewhere | MATCH | 720 |
| sum_hour_group1_no_prewhere | MATCH | 720 |
| sum_hour_tz | MATCH | 720 |
| sum_minute | MATCH | 29632 |
| sum_month | MATCH | 10 |
| sum_no_window | MATCH | 10 |
| sum_total | MATCH | 1 |
| sum_total_by_subject | MATCH | 10 |
| sum_total_filter1 | MATCH | 1 |
| sum_total_filter2 | MATCH | 1 |
| unique_count_hour | MATCH | 720 |
| workload_seconds_by_region | MATCH | 288 |

MATCH=30 DIFFERS=0 (one side only=0)

## altinity/local-single → altinity/kind

| Query | p50 Δ | CPU Δ | base p50 ms | cand p50 ms |
| --- | ---: | ---: | ---: | ---: |
| agent_runs_by_name | +6% | -4% | 17.0 | 18.0 |
| avg_hour | +24% | +10% | 41.0 | 51.0 |
| count_hour | +25% | +11% | 16.0 | 20.0 |
| count_total | -20% | +17% | 15.0 | 12.0 |
| distinct_subjects | +42% | +50% | 26.0 | 37.0 |
| kong_api_request_total | -6% | +23% | 16.0 | 15.0 |
| kong_llm_tokens_total | +28% | +32% | 82.0 | 105.0 |
| kong_status_by_route | +59% | +21% | 34.0 | 54.0 |
| latest_hour | -4% | +14% | 49.0 | 47.0 |
| llm_tokens_by_model | +104% | +55% | 26.0 | 53.0 |
| max_hour | +26% | +9% | 39.0 | 49.0 |
| min_hour | +17% | +17% | 46.0 | 54.0 |
| sum_day | +36% | +27% | 36.0 | 49.0 |
| sum_day_by_dim | +50% | +35% | 32.0 | 48.0 |
| sum_day_by_subject | -20% | -4% | 51.0 | 41.0 |
| sum_hour | -22% | -15% | 49.0 | 38.0 |
| sum_hour_group1 | -7% | +2% | 46.0 | 43.0 |
| sum_hour_group1_group2 | +45% | +27% | 40.0 | 58.0 |
| sum_hour_group1_group2_no_prewhere | +0% | +12% | 48.0 | 48.0 |
| sum_hour_group1_no_prewhere | +7% | +11% | 40.0 | 43.0 |
| sum_hour_tz | +24% | +11% | 37.0 | 46.0 |
| sum_minute | +61% | +38% | 44.0 | 71.0 |
| sum_month | +20% | +22% | 35.0 | 42.0 |
| sum_no_window | +31% | +18% | 36.0 | 47.0 |
| sum_total | +21% | +11% | 39.0 | 47.0 |
| sum_total_by_subject | +47% | +18% | 40.0 | 59.0 |
| sum_total_filter1 | +8% | +6% | 38.0 | 41.0 |
| sum_total_filter2 | +0% | +4% | 40.0 | 40.0 |
| unique_count_hour | +51% | +61% | 53.0 | 80.0 |
| workload_seconds_by_region | +42% | +18% | 19.0 | 27.0 |

Ingest: base 98128 events/s, candidate 57336 events/s (Δ -42%)

## Value parity (window 2026-05-29 00:00:00 → 2026-06-01 00:00:00)

| Query | result | rows |
| --- | --- | ---: |
| agent_runs_by_name | MATCH | 288 |
| avg_hour | MATCH | 720 |
| count_hour | MATCH | 720 |
| count_total | MATCH | 1 |
| distinct_subjects | MATCH | 100 |
| kong_api_request_total | MATCH | 1 |
| kong_llm_tokens_total | MATCH | 1 |
| kong_status_by_route | MATCH | 6703 |
| latest_hour | MATCH | 720 |
| llm_tokens_by_model | MATCH | 1080 |
| max_hour | MATCH | 720 |
| min_hour | MATCH | 720 |
| sum_day | MATCH | 30 |
| sum_day_by_dim | MATCH | 150 |
| sum_day_by_subject | MATCH | 30 |
| sum_hour | MATCH | 720 |
| sum_hour_group1 | MATCH | 720 |
| sum_hour_group1_group2 | MATCH | 720 |
| sum_hour_group1_group2_no_prewhere | MATCH | 720 |
| sum_hour_group1_no_prewhere | MATCH | 720 |
| sum_hour_tz | MATCH | 720 |
| sum_minute | MATCH | 29632 |
| sum_month | MATCH | 10 |
| sum_no_window | MATCH | 10 |
| sum_total | MATCH | 1 |
| sum_total_by_subject | MATCH | 10 |
| sum_total_filter1 | MATCH | 1 |
| sum_total_filter2 | MATCH | 1 |
| unique_count_hour | MATCH | 720 |
| workload_seconds_by_region | MATCH | 288 |

MATCH=30 DIFFERS=0 (one side only=0)

compare exit status: 0 (non-zero = parity failed or not established)

---

## Compare openmeter/proposal

> **Smoke, not performance.** At least one side is a local tier, or the sides are on different tiers. Latency and CPU deltas below are not evidence about the platforms.

| Side | Target | Version | Topology | Run | Commit |
| --- | --- | --- | --- | --- | --- |
| altinity/local-single | altinity/local-single (local-single) | 26.3.33.10001.altinitystable | single 1x1 local | 20261008T075134Z-ad96 | 3701df2 |
| altinity/local-cluster-2x2 | altinity/local-cluster-2x2 (local-cluster) | 26.3.33.10001.altinitystable | replicated 2x2 local | 20261008T075436Z-ed35 | 3701df2 |
| altinity/kind | altinity/kind (kind) | 26.3.33.10001.altinitystable | replicated 1x2 local | 20261008T075930Z-b24c | 3701df2 |

## altinity/local-single → altinity/local-cluster-2x2

| Query | p50 Δ | CPU Δ | base p50 ms | cand p50 ms |
| --- | ---: | ---: | ---: | ---: |
| agent_runs_by_name | +71% | +45% | 14.0 | 24.0 |
| avg_hour | +34% | +31% | 38.0 | 51.0 |
| count_hour | -4% | -2% | 27.0 | 26.0 |
| count_total | +57% | +38% | 14.0 | 22.0 |
| distinct_subjects | +21% | +16% | 24.0 | 29.0 |
| kong_api_request_by_all_dims | +25% | +17% | 128.0 | 160.0 |
| kong_api_request_by_method | +0% | +27% | 27.0 | 27.0 |
| kong_api_request_by_service | +11% | +51% | 36.0 | 40.0 |
| kong_api_request_total | +100% | +81% | 13.0 | 26.0 |
| kong_llm_tokens_total | +28% | +8% | 25.0 | 32.0 |
| kong_status_by_route | +95% | +76% | 22.0 | 43.0 |
| latest_hour | -13% | +9% | 45.0 | 39.0 |
| llm_tokens_by_model | +53% | +32% | 30.0 | 46.0 |
| lookup_by_id | +33% | +35% | 91.0 | 121.0 |
| max_hour | +66% | +103% | 41.0 | 68.0 |
| min_hour | -21% | +0% | 42.0 | 33.0 |
| sum_day | +49% | +19% | 35.0 | 52.0 |
| sum_day_by_dim | +17% | +10% | 30.0 | 35.0 |
| sum_day_by_subject | +17% | +25% | 36.0 | 42.0 |
| sum_hour | +15% | +34% | 33.0 | 38.0 |
| sum_hour_group1 | +52% | +36% | 33.0 | 50.0 |
| sum_hour_group1_group2 | +16% | +28% | 43.0 | 50.0 |
| sum_hour_group1_group2_no_prewhere | +15% | +22% | 40.0 | 46.0 |
| sum_hour_group1_no_prewhere | +90% | +32% | 29.0 | 55.0 |
| sum_hour_tz | +86% | +67% | 28.0 | 52.0 |
| sum_minute | +68% | +47% | 37.0 | 62.0 |
| sum_month | +3% | -0% | 32.0 | 33.0 |
| sum_no_window | +18% | +16% | 34.0 | 40.0 |
| sum_total | +38% | +20% | 26.0 | 36.0 |
| sum_total_by_subject | +23% | +12% | 26.0 | 32.0 |
| sum_total_filter1 | +87% | +35% | 24.0 | 45.0 |
| sum_total_filter2 | +5% | +21% | 37.0 | 39.0 |
| unique_count_hour | +42% | +43% | 31.0 | 44.0 |
| workload_seconds_by_region | +35% | +28% | 20.0 | 27.0 |

Ingest: base 55897 events/s, candidate 38251 events/s (Δ -32%)

## Value parity (window 2026-05-29 00:00:00 → 2026-06-01 00:00:00)

| Query | result | rows |
| --- | --- | ---: |
| agent_runs_by_name | MATCH | 288 |
| avg_hour | MATCH | 720 |
| count_hour | MATCH | 720 |
| count_total | MATCH | 1 |
| distinct_subjects | MATCH | 100 |
| kong_api_request_by_all_dims | MATCH | 24803 |
| kong_api_request_by_method | MATCH | 360 |
| kong_api_request_by_service | MATCH | 2880 |
| kong_api_request_total | MATCH | 1 |
| kong_llm_tokens_total | MATCH | 1 |
| kong_status_by_route | MATCH | 6703 |
| latest_hour | MATCH | 720 |
| llm_tokens_by_model | MATCH | 1080 |
| lookup_by_id | MATCH | 1 |
| max_hour | MATCH | 720 |
| min_hour | MATCH | 720 |
| sum_day | MATCH | 30 |
| sum_day_by_dim | MATCH | 150 |
| sum_day_by_subject | MATCH | 30 |
| sum_hour | MATCH | 720 |
| sum_hour_group1 | MATCH | 720 |
| sum_hour_group1_group2 | MATCH | 720 |
| sum_hour_group1_group2_no_prewhere | MATCH | 720 |
| sum_hour_group1_no_prewhere | MATCH | 720 |
| sum_hour_tz | MATCH | 720 |
| sum_minute | MATCH | 29632 |
| sum_month | MATCH | 10 |
| sum_no_window | MATCH | 10 |
| sum_total | MATCH | 1 |
| sum_total_by_subject | MATCH | 10 |
| sum_total_filter1 | MATCH | 1 |
| sum_total_filter2 | MATCH | 1 |
| unique_count_hour | MATCH | 720 |
| workload_seconds_by_region | MATCH | 288 |

MATCH=34 DIFFERS=0 (one side only=0)

## altinity/local-single → altinity/kind

| Query | p50 Δ | CPU Δ | base p50 ms | cand p50 ms |
| --- | ---: | ---: | ---: | ---: |
| agent_runs_by_name | +50% | +19% | 14.0 | 21.0 |
| avg_hour | +8% | +11% | 38.0 | 41.0 |
| count_hour | -26% | -16% | 27.0 | 20.0 |
| count_total | +7% | +1% | 14.0 | 15.0 |
| distinct_subjects | +50% | +16% | 24.0 | 36.0 |
| kong_api_request_by_all_dims | +8% | +9% | 128.0 | 138.0 |
| kong_api_request_by_method | +4% | +14% | 27.0 | 28.0 |
| kong_api_request_by_service | +22% | +51% | 36.0 | 44.0 |
| kong_api_request_total | +38% | +41% | 13.0 | 18.0 |
| kong_llm_tokens_total | +20% | -10% | 25.0 | 30.0 |
| kong_status_by_route | +23% | +21% | 22.0 | 27.0 |
| latest_hour | -4% | +11% | 45.0 | 43.0 |
| llm_tokens_by_model | +3% | +6% | 30.0 | 31.0 |
| lookup_by_id | +13% | +6% | 91.0 | 103.0 |
| max_hour | -15% | +28% | 41.0 | 35.0 |
| min_hour | -21% | -30% | 42.0 | 33.0 |
| sum_day | +31% | -1% | 35.0 | 46.0 |
| sum_day_by_dim | +33% | -1% | 30.0 | 40.0 |
| sum_day_by_subject | +11% | +3% | 36.0 | 40.0 |
| sum_hour | +3% | +27% | 33.0 | 34.0 |
| sum_hour_group1 | +15% | +11% | 33.0 | 38.0 |
| sum_hour_group1_group2 | +19% | +24% | 43.0 | 51.0 |
| sum_hour_group1_group2_no_prewhere | -10% | +12% | 40.0 | 36.0 |
| sum_hour_group1_no_prewhere | +31% | +15% | 29.0 | 38.0 |
| sum_hour_tz | +32% | +17% | 28.0 | 37.0 |
| sum_minute | +30% | +21% | 37.0 | 48.0 |
| sum_month | +31% | +13% | 32.0 | 42.0 |
| sum_no_window | -3% | +15% | 34.0 | 33.0 |
| sum_total | +31% | -2% | 26.0 | 34.0 |
| sum_total_by_subject | +35% | +13% | 26.0 | 35.0 |
| sum_total_filter1 | +71% | -1% | 24.0 | 41.0 |
| sum_total_filter2 | +24% | +16% | 37.0 | 46.0 |
| unique_count_hour | +35% | +22% | 31.0 | 42.0 |
| workload_seconds_by_region | +10% | -2% | 20.0 | 22.0 |

Ingest: base 55897 events/s, candidate 33752 events/s (Δ -40%)

## Value parity (window 2026-05-29 00:00:00 → 2026-06-01 00:00:00)

| Query | result | rows |
| --- | --- | ---: |
| agent_runs_by_name | MATCH | 288 |
| avg_hour | MATCH | 720 |
| count_hour | MATCH | 720 |
| count_total | MATCH | 1 |
| distinct_subjects | MATCH | 100 |
| kong_api_request_by_all_dims | MATCH | 24803 |
| kong_api_request_by_method | MATCH | 360 |
| kong_api_request_by_service | MATCH | 2880 |
| kong_api_request_total | MATCH | 1 |
| kong_llm_tokens_total | MATCH | 1 |
| kong_status_by_route | MATCH | 6703 |
| latest_hour | MATCH | 720 |
| llm_tokens_by_model | MATCH | 1080 |
| lookup_by_id | MATCH | 1 |
| max_hour | MATCH | 720 |
| min_hour | MATCH | 720 |
| sum_day | MATCH | 30 |
| sum_day_by_dim | MATCH | 150 |
| sum_day_by_subject | MATCH | 30 |
| sum_hour | MATCH | 720 |
| sum_hour_group1 | MATCH | 720 |
| sum_hour_group1_group2 | MATCH | 720 |
| sum_hour_group1_group2_no_prewhere | MATCH | 720 |
| sum_hour_group1_no_prewhere | MATCH | 720 |
| sum_hour_tz | MATCH | 720 |
| sum_minute | MATCH | 29632 |
| sum_month | MATCH | 10 |
| sum_no_window | MATCH | 10 |
| sum_total | MATCH | 1 |
| sum_total_by_subject | MATCH | 10 |
| sum_total_filter1 | MATCH | 1 |
| sum_total_filter2 | MATCH | 1 |
| unique_count_hour | MATCH | 720 |
| workload_seconds_by_region | MATCH | 288 |

MATCH=34 DIFFERS=0 (one side only=0)

compare exit status: 0 (non-zero = parity failed or not established)

---

## Compare otel-traces/clickstack

> **Smoke, not performance.** At least one side is a local tier, or the sides are on different tiers. Latency and CPU deltas below are not evidence about the platforms.

| Side | Target | Version | Topology | Run | Commit |
| --- | --- | --- | --- | --- | --- |
| altinity/local-single | altinity/local-single (local-single) | 26.3.33.10001.altinitystable | single 1x1 local | 20261008T075204Z-75bd | 3701df2 |
| altinity/local-cluster-2x2 | altinity/local-cluster-2x2 (local-cluster) | 26.3.33.10001.altinitystable | replicated 2x2 local | 20261008T075520Z-8e5e | 3701df2 |
| altinity/kind | altinity/kind (kind) | 26.3.33.10001.altinitystable | replicated 1x2 local | 20261008T080017Z-ba71 | 3701df2 |

## altinity/local-single → altinity/local-cluster-2x2

| Query | p50 Δ | CPU Δ | base p50 ms | cand p50 ms |
| --- | ---: | ---: | ---: | ---: |
| error_rate_24h | +267% | +170% | 12.0 | 44.0 |
| latency_by_service_1h | +53% | +103% | 19.0 | 29.0 |
| rollup_values_24h | +231% | +228% | 13.0 | 43.0 |
| route_status_24h | +106% | +97% | 18.0 | 37.0 |
| route_status_items_24h | +214% | +161% | 14.0 | 44.0 |
| service_dependencies_1h | +3% | +33% | 69.0 | 71.0 |
| slowest_traces_1h | +142% | +155% | 12.0 | 29.0 |
| spans_by_hour | +127% | +97% | 30.0 | 68.0 |
| top_operations_24h | +145% | +153% | 11.0 | 27.0 |
| trace_by_id | +90% | +124% | 10.0 | 19.0 |
| user_activity | +370% | +342% | 10.0 | 47.0 |

Ingest: base 36321 events/s, candidate 16357 events/s (Δ -55%)

## Value parity (window 2026-05-29 00:00:00 → 2026-06-01 00:00:00)

| Query | result | rows |
| --- | --- | ---: |
| error_rate_24h | MATCH | 27 |
| latency_by_service_1h | MATCH | 310 |
| parity/kv_rollup | MATCH | 1 |
| parity/otel_traces_rows | MATCH | 1 |
| rollup_values_24h | MATCH | 20 |
| route_status_24h | MATCH | 4 |
| route_status_items_24h | MATCH | 4 |
| service_dependencies_1h | MATCH | 128 |
| slowest_traces_1h | MATCH | 20 |
| spans_by_hour | MATCH | 1944 |
| top_operations_24h | MATCH | 10 |
| trace_by_id | MATCH | 7 |
| user_activity | MATCH | 1 |

MATCH=13 DIFFERS=0 (one side only=0)

## altinity/local-single → altinity/kind

| Query | p50 Δ | CPU Δ | base p50 ms | cand p50 ms |
| --- | ---: | ---: | ---: | ---: |
| error_rate_24h | +25% | +13% | 12.0 | 15.0 |
| latency_by_service_1h | +21% | +23% | 19.0 | 23.0 |
| rollup_values_24h | +15% | +15% | 13.0 | 15.0 |
| route_status_24h | +0% | -9% | 18.0 | 18.0 |
| route_status_items_24h | +29% | -3% | 14.0 | 18.0 |
| service_dependencies_1h | -19% | -18% | 69.0 | 56.0 |
| slowest_traces_1h | +8% | +0% | 12.0 | 13.0 |
| spans_by_hour | +3% | -1% | 30.0 | 31.0 |
| top_operations_24h | +27% | +20% | 11.0 | 14.0 |
| trace_by_id | +10% | +8% | 10.0 | 11.0 |
| user_activity | +40% | +33% | 10.0 | 14.0 |

Ingest: base 36321 events/s, candidate 19872 events/s (Δ -45%)

## Value parity (window 2026-05-29 00:00:00 → 2026-06-01 00:00:00)

| Query | result | rows |
| --- | --- | ---: |
| error_rate_24h | MATCH | 27 |
| latency_by_service_1h | MATCH | 310 |
| parity/kv_rollup | MATCH | 1 |
| parity/otel_traces_rows | MATCH | 1 |
| rollup_values_24h | MATCH | 20 |
| route_status_24h | MATCH | 4 |
| route_status_items_24h | MATCH | 4 |
| service_dependencies_1h | MATCH | 128 |
| slowest_traces_1h | MATCH | 20 |
| spans_by_hour | MATCH | 1944 |
| top_operations_24h | MATCH | 10 |
| trace_by_id | MATCH | 7 |
| user_activity | MATCH | 1 |

MATCH=13 DIFFERS=0 (one side only=0)

compare exit status: 0 (non-zero = parity failed or not established)

---

## Compare otel-traces/exporter-default

> **Smoke, not performance.** At least one side is a local tier, or the sides are on different tiers. Latency and CPU deltas below are not evidence about the platforms.

| Side | Target | Version | Topology | Run | Commit |
| --- | --- | --- | --- | --- | --- |
| altinity/local-single | altinity/local-single (local-single) | 26.3.33.10001.altinitystable | single 1x1 local | 20261008T075222Z-b56e | 3701df2 |
| altinity/local-cluster-2x2 | altinity/local-cluster-2x2 (local-cluster) | 26.3.33.10001.altinitystable | replicated 2x2 local | 20261008T075558Z-8046 | 3701df2 |
| altinity/kind | altinity/kind (kind) | 26.3.33.10001.altinitystable | replicated 1x2 local | 20261008T080049Z-35a0 | 3701df2 |

## altinity/local-single → altinity/local-cluster-2x2

| Query | p50 Δ | CPU Δ | base p50 ms | cand p50 ms |
| --- | ---: | ---: | ---: | ---: |
| error_rate_24h | +80% | +19% | 20.0 | 36.0 |
| latency_by_service_1h | +20% | +27% | 25.0 | 30.0 |
| route_status_24h | +52% | +11% | 25.0 | 38.0 |
| service_dependencies_1h | -18% | -8% | 68.0 | 56.0 |
| slowest_traces_1h | +5% | +21% | 22.0 | 23.0 |
| spans_by_hour | +34% | +21% | 35.0 | 47.0 |
| top_operations_24h | +92% | +30% | 13.0 | 25.0 |
| trace_by_id | +35% | +55% | 17.0 | 23.0 |
| trace_by_id_lookup | +143% | +130% | 30.0 | 73.0 |
| user_activity | +13% | +19% | 23.0 | 26.0 |

Ingest: base 41950 events/s, candidate 21876 events/s (Δ -48%)

## Value parity (window 2026-05-29 00:00:00 → 2026-06-01 00:00:00)

| Query | result | rows |
| --- | --- | ---: |
| error_rate_24h | MATCH | 27 |
| latency_by_service_1h | MATCH | 310 |
| parity/otel_traces_rows | MATCH | 1 |
| parity/trace_id_ts | MATCH | 1 |
| route_status_24h | MATCH | 4 |
| service_dependencies_1h | MATCH | 128 |
| slowest_traces_1h | MATCH | 20 |
| spans_by_hour | MATCH | 1944 |
| top_operations_24h | MATCH | 10 |
| trace_by_id | MATCH | 7 |
| trace_by_id_lookup | MATCH | 7 |
| user_activity | MATCH | 1 |

MATCH=12 DIFFERS=0 (one side only=0)

## altinity/local-single → altinity/kind

| Query | p50 Δ | CPU Δ | base p50 ms | cand p50 ms |
| --- | ---: | ---: | ---: | ---: |
| error_rate_24h | +30% | -9% | 20.0 | 26.0 |
| latency_by_service_1h | -8% | -7% | 25.0 | 23.0 |
| route_status_24h | +20% | +7% | 25.0 | 30.0 |
| service_dependencies_1h | -15% | -19% | 68.0 | 58.0 |
| slowest_traces_1h | +5% | +15% | 22.0 | 23.0 |
| spans_by_hour | +26% | +8% | 35.0 | 44.0 |
| top_operations_24h | +8% | +6% | 13.0 | 14.0 |
| trace_by_id | +12% | +28% | 17.0 | 19.0 |
| trace_by_id_lookup | +23% | +16% | 30.0 | 37.0 |
| user_activity | +30% | +4% | 23.0 | 30.0 |

Ingest: base 41950 events/s, candidate 24301 events/s (Δ -42%)

## Value parity (window 2026-05-29 00:00:00 → 2026-06-01 00:00:00)

| Query | result | rows |
| --- | --- | ---: |
| error_rate_24h | MATCH | 27 |
| latency_by_service_1h | MATCH | 310 |
| parity/otel_traces_rows | MATCH | 1 |
| parity/trace_id_ts | MATCH | 1 |
| route_status_24h | MATCH | 4 |
| service_dependencies_1h | MATCH | 128 |
| slowest_traces_1h | MATCH | 20 |
| spans_by_hour | MATCH | 1944 |
| top_operations_24h | MATCH | 10 |
| trace_by_id | MATCH | 7 |
| trace_by_id_lookup | MATCH | 7 |
| user_activity | MATCH | 1 |

MATCH=12 DIFFERS=0 (one side only=0)

compare exit status: 0 (non-zero = parity failed or not established)

---

## Compare otel-logs/clickstack

> **Smoke, not performance.** At least one side is a local tier, or the sides are on different tiers. Latency and CPU deltas below are not evidence about the platforms.

| Side | Target | Version | Topology | Run | Commit |
| --- | --- | --- | --- | --- | --- |
| altinity/local-single | altinity/local-single (local-single) | 26.3.33.10001.altinitystable | single 1x1 local | 20261008T075239Z-db4d | 3701df2 |
| altinity/local-cluster-2x2 | altinity/local-cluster-2x2 (local-cluster) | 26.3.33.10001.altinitystable | replicated 2x2 local | 20261008T075628Z-1de3 | 3701df2 |
| altinity/kind | altinity/kind (kind) | 26.3.33.10001.altinitystable | replicated 1x2 local | 20261008T080116Z-983b | 3701df2 |

## altinity/local-single → altinity/local-cluster-2x2

| Query | p50 Δ | CPU Δ | base p50 ms | cand p50 ms |
| --- | ---: | ---: | ---: | ---: |
| body_contains_24h | +79% | +25% | 34.0 | 61.0 |
| logs_for_trace | +160% | +178% | 10.0 | 26.0 |
| needle_search | +67% | +94% | 9.0 | 15.0 |
| rollup_services_24h | +110% | +105% | 10.0 | 21.0 |
| service_tail_1h | +93% | +85% | 14.0 | 27.0 |
| severity_by_service_24h | +72% | +69% | 25.0 | 43.0 |
| top_error_patterns_24h | +105% | +72% | 22.0 | 45.0 |
| user_logs | +71% | +54% | 21.0 | 36.0 |
| user_logs_items | +43% | +58% | 23.0 | 33.0 |
| volume_by_hour | +33% | +43% | 36.0 | 48.0 |

Ingest: base 23507 events/s, candidate 14050 events/s (Δ -40%)

## Value parity (window 2026-05-29 00:00:00 → 2026-06-01 00:00:00)

| Query | result | rows |
| --- | --- | ---: |
| body_contains_24h | MATCH | 1 |
| logs_for_trace | MATCH | 4 |
| needle_search | MATCH | 1 |
| parity/kv_rollup | MATCH | 1 |
| parity/otel_logs_rows | MATCH | 1 |
| rollup_services_24h | MATCH | 20 |
| service_tail_1h | MATCH | 100 |
| severity_by_service_24h | MATCH | 180 |
| top_error_patterns_24h | MATCH | 10 |
| user_logs | MATCH | 3 |
| user_logs_items | MATCH | 3 |
| volume_by_hour | MATCH | 432 |

MATCH=12 DIFFERS=0 (one side only=0)

## altinity/local-single → altinity/kind

| Query | p50 Δ | CPU Δ | base p50 ms | cand p50 ms |
| --- | ---: | ---: | ---: | ---: |
| body_contains_24h | -6% | +0% | 34.0 | 32.0 |
| logs_for_trace | +140% | +101% | 10.0 | 24.0 |
| needle_search | +0% | +3% | 9.0 | 9.0 |
| rollup_services_24h | +0% | -4% | 10.0 | 10.0 |
| service_tail_1h | +64% | +17% | 14.0 | 23.0 |
| severity_by_service_24h | -4% | +6% | 25.0 | 24.0 |
| top_error_patterns_24h | +45% | +15% | 22.0 | 32.0 |
| user_logs | -33% | -21% | 21.0 | 14.0 |
| user_logs_items | -4% | -7% | 23.0 | 22.0 |
| volume_by_hour | -22% | -6% | 36.0 | 28.0 |

Ingest: base 23507 events/s, candidate 18118 events/s (Δ -23%)

## Value parity (window 2026-05-29 00:00:00 → 2026-06-01 00:00:00)

| Query | result | rows |
| --- | --- | ---: |
| body_contains_24h | MATCH | 1 |
| logs_for_trace | MATCH | 4 |
| needle_search | MATCH | 1 |
| parity/kv_rollup | MATCH | 1 |
| parity/otel_logs_rows | MATCH | 1 |
| rollup_services_24h | MATCH | 20 |
| service_tail_1h | MATCH | 100 |
| severity_by_service_24h | MATCH | 180 |
| top_error_patterns_24h | MATCH | 10 |
| user_logs | MATCH | 3 |
| user_logs_items | MATCH | 3 |
| volume_by_hour | MATCH | 432 |

MATCH=12 DIFFERS=0 (one side only=0)

compare exit status: 0 (non-zero = parity failed or not established)

---

## Compare otel-logs/exporter-default

> **Smoke, not performance.** At least one side is a local tier, or the sides are on different tiers. Latency and CPU deltas below are not evidence about the platforms.

| Side | Target | Version | Topology | Run | Commit |
| --- | --- | --- | --- | --- | --- |
| altinity/local-single | altinity/local-single (local-single) | 26.3.33.10001.altinitystable | single 1x1 local | 20261008T075304Z-0dff | 3701df2 |
| altinity/local-cluster-2x2 | altinity/local-cluster-2x2 (local-cluster) | 26.3.33.10001.altinitystable | replicated 2x2 local | 20261008T075710Z-76bc | 3701df2 |
| altinity/kind | altinity/kind (kind) | 26.3.33.10001.altinitystable | replicated 1x2 local | 20261008T080150Z-1e75 | 3701df2 |

## altinity/local-single → altinity/local-cluster-2x2

| Query | p50 Δ | CPU Δ | base p50 ms | cand p50 ms |
| --- | ---: | ---: | ---: | ---: |
| body_contains_24h | +27% | +16% | 30.0 | 38.0 |
| logs_for_trace | +78% | +94% | 9.0 | 16.0 |
| needle_search | +78% | +109% | 9.0 | 16.0 |
| service_tail_1h | +26% | +12% | 19.0 | 24.0 |
| severity_by_service_24h | +55% | +34% | 22.0 | 34.0 |
| top_error_patterns_24h | +52% | +31% | 23.0 | 35.0 |
| user_logs | +19% | +52% | 21.0 | 25.0 |
| volume_by_hour | +61% | +26% | 23.0 | 37.0 |

Ingest: base 33805 events/s, candidate 21428 events/s (Δ -37%)

## Value parity (window 2026-05-29 00:00:00 → 2026-06-01 00:00:00)

| Query | result | rows |
| --- | --- | ---: |
| body_contains_24h | MATCH | 1 |
| logs_for_trace | MATCH | 4 |
| needle_search | MATCH | 1 |
| parity/otel_logs_rows | MATCH | 1 |
| service_tail_1h | MATCH | 100 |
| severity_by_service_24h | MATCH | 180 |
| top_error_patterns_24h | MATCH | 10 |
| user_logs | MATCH | 3 |
| volume_by_hour | MATCH | 432 |

MATCH=9 DIFFERS=0 (one side only=0)

## altinity/local-single → altinity/kind

| Query | p50 Δ | CPU Δ | base p50 ms | cand p50 ms |
| --- | ---: | ---: | ---: | ---: |
| body_contains_24h | -23% | -11% | 30.0 | 23.0 |
| logs_for_trace | +44% | +32% | 9.0 | 13.0 |
| needle_search | -22% | -13% | 9.0 | 7.0 |
| service_tail_1h | -16% | -10% | 19.0 | 16.0 |
| severity_by_service_24h | +27% | +12% | 22.0 | 28.0 |
| top_error_patterns_24h | +22% | +8% | 23.0 | 28.0 |
| user_logs | -33% | -9% | 21.0 | 14.0 |
| volume_by_hour | +22% | +5% | 23.0 | 28.0 |

Ingest: base 33805 events/s, candidate 23303 events/s (Δ -31%)

## Value parity (window 2026-05-29 00:00:00 → 2026-06-01 00:00:00)

| Query | result | rows |
| --- | --- | ---: |
| body_contains_24h | MATCH | 1 |
| logs_for_trace | MATCH | 4 |
| needle_search | MATCH | 1 |
| parity/otel_logs_rows | MATCH | 1 |
| service_tail_1h | MATCH | 100 |
| severity_by_service_24h | MATCH | 180 |
| top_error_patterns_24h | MATCH | 10 |
| user_logs | MATCH | 3 |
| volume_by_hour | MATCH | 432 |

MATCH=9 DIFFERS=0 (one side only=0)

compare exit status: 0 (non-zero = parity failed or not established)

---

## Compare otel-metrics/clickstack

> **Smoke, not performance.** At least one side is a local tier, or the sides are on different tiers. Latency and CPU deltas below are not evidence about the platforms.

| Side | Target | Version | Topology | Run | Commit |
| --- | --- | --- | --- | --- | --- |
| altinity/local-single | altinity/local-single (local-single) | 26.3.33.10001.altinitystable | single 1x1 local | 20261008T075322Z-862d | 3701df2 |
| altinity/local-cluster-2x2 | altinity/local-cluster-2x2 (local-cluster) | 26.3.33.10001.altinitystable | replicated 2x2 local | 20261008T075739Z-922b | 3701df2 |
| altinity/kind | altinity/kind (kind) | 26.3.33.10001.altinitystable | replicated 1x2 local | 20261008T080216Z-9772 | 3701df2 |

## altinity/local-single → altinity/local-cluster-2x2

| Query | p50 Δ | CPU Δ | base p50 ms | cand p50 ms |
| --- | ---: | ---: | ---: | ---: |
| cpu_by_service_1h | +114% | +125% | 14.0 | 30.0 |
| error_ratio_1h | -2% | +10% | 166.0 | 163.0 |
| gc_pause_median_1h | +367% | +291% | 12.0 | 56.0 |
| latency_p95_by_route_1h | +19% | +53% | 138.0 | 164.0 |
| memory_by_pod_1h | +19% | +59% | 36.0 | 43.0 |
| request_rate_by_route_1h | +24% | +33% | 225.0 | 279.0 |
| rpc_count_by_method_1h | +0% | +49% | 72.0 | 72.0 |
| series_by_metric | +85% | +98% | 162.0 | 300.0 |
| user_requests | +188% | +225% | 8.0 | 23.0 |

Ingest: base 44944 events/s, candidate 17858 events/s (Δ -60%)

## Value parity (window 2026-05-31 22:15:00 → 2026-06-01 00:00:01)

| Query | result | rows |
| --- | --- | ---: |
| cpu_by_service_1h | MATCH | 1800 |
| error_ratio_1h | MATCH | 19 |
| gc_pause_median_1h | MATCH | 7 |
| latency_p95_by_route_1h | MATCH | 40 |
| memory_by_pod_1h | MATCH | 90 |
| parity/exponential_histogram_rows | MATCH | 1 |
| parity/gauge_rows | MATCH | 1 |
| parity/histogram_rows | MATCH | 1 |
| parity/sum_rows | MATCH | 1 |
| parity/summary_rows | MATCH | 1 |
| request_rate_by_route_1h | MATCH | 4004 |
| rpc_count_by_method_1h | MATCH | 95 |
| series_by_metric | MATCH | 7 |
| user_requests | MATCH | 0 |

MATCH=14 DIFFERS=0 (one side only=0)

## altinity/local-single → altinity/kind

| Query | p50 Δ | CPU Δ | base p50 ms | cand p50 ms |
| --- | ---: | ---: | ---: | ---: |
| cpu_by_service_1h | +21% | +0% | 14.0 | 17.0 |
| error_ratio_1h | -1% | -18% | 166.0 | 164.0 |
| gc_pause_median_1h | +0% | -24% | 12.0 | 12.0 |
| latency_p95_by_route_1h | +4% | -4% | 138.0 | 143.0 |
| memory_by_pod_1h | +8% | +10% | 36.0 | 39.0 |
| request_rate_by_route_1h | -12% | -10% | 225.0 | 199.0 |
| rpc_count_by_method_1h | +8% | +8% | 72.0 | 78.0 |
| series_by_metric | +30% | +9% | 162.0 | 211.0 |
| user_requests | +62% | +53% | 8.0 | 13.0 |

Ingest: base 44944 events/s, candidate 21299 events/s (Δ -53%)

## Value parity (window 2026-05-31 22:15:00 → 2026-06-01 00:00:01)

| Query | result | rows |
| --- | --- | ---: |
| cpu_by_service_1h | MATCH | 1800 |
| error_ratio_1h | MATCH | 19 |
| gc_pause_median_1h | MATCH | 7 |
| latency_p95_by_route_1h | MATCH | 40 |
| memory_by_pod_1h | MATCH | 90 |
| parity/exponential_histogram_rows | MATCH | 1 |
| parity/gauge_rows | MATCH | 1 |
| parity/histogram_rows | MATCH | 1 |
| parity/sum_rows | MATCH | 1 |
| parity/summary_rows | MATCH | 1 |
| request_rate_by_route_1h | MATCH | 4004 |
| rpc_count_by_method_1h | MATCH | 95 |
| series_by_metric | MATCH | 7 |
| user_requests | MATCH | 0 |

MATCH=14 DIFFERS=0 (one side only=0)

compare exit status: 0 (non-zero = parity failed or not established)

---

## Compare otel-metrics/exporter-default

> **Smoke, not performance.** At least one side is a local tier, or the sides are on different tiers. Latency and CPU deltas below are not evidence about the platforms.

| Side | Target | Version | Topology | Run | Commit |
| --- | --- | --- | --- | --- | --- |
| altinity/local-single | altinity/local-single (local-single) | 26.3.33.10001.altinitystable | single 1x1 local | 20261008T075340Z-6b31 | 3701df2 |
| altinity/local-cluster-2x2 | altinity/local-cluster-2x2 (local-cluster) | 26.3.33.10001.altinitystable | replicated 2x2 local | 20261008T075817Z-6027 | 3701df2 |
| altinity/kind | altinity/kind (kind) | 26.3.33.10001.altinitystable | replicated 1x2 local | 20261008T080249Z-e25a | 3701df2 |

## altinity/local-single → altinity/local-cluster-2x2

| Query | p50 Δ | CPU Δ | base p50 ms | cand p50 ms |
| --- | ---: | ---: | ---: | ---: |
| cpu_by_service_1h | +30% | +58% | 23.0 | 30.0 |
| error_ratio_1h | +6% | +12% | 178.0 | 188.0 |
| gc_pause_median_1h | +65% | +76% | 17.0 | 28.0 |
| latency_p95_by_route_1h | -17% | +27% | 144.0 | 119.0 |
| memory_by_pod_1h | +26% | +85% | 34.0 | 43.0 |
| request_rate_by_route_1h | -18% | +3% | 339.0 | 277.0 |
| rpc_count_by_method_1h | -17% | +14% | 155.0 | 128.0 |
| series_by_metric | +10% | +39% | 286.0 | 316.0 |
| user_requests | -23% | +41% | 26.0 | 20.0 |

Ingest: base 46267 events/s, candidate 17912 events/s (Δ -61%)

## Value parity (window 2026-05-31 22:15:00 → 2026-06-01 00:00:01)

| Query | result | rows |
| --- | --- | ---: |
| cpu_by_service_1h | MATCH | 1800 |
| error_ratio_1h | MATCH | 19 |
| gc_pause_median_1h | MATCH | 7 |
| latency_p95_by_route_1h | MATCH | 40 |
| memory_by_pod_1h | MATCH | 90 |
| parity/exponential_histogram_rows | MATCH | 1 |
| parity/gauge_rows | MATCH | 1 |
| parity/histogram_rows | MATCH | 1 |
| parity/sum_rows | MATCH | 1 |
| parity/summary_rows | MATCH | 1 |
| request_rate_by_route_1h | MATCH | 4004 |
| rpc_count_by_method_1h | MATCH | 95 |
| series_by_metric | MATCH | 7 |
| user_requests | MATCH | 0 |

MATCH=14 DIFFERS=0 (one side only=0)

## altinity/local-single → altinity/kind

| Query | p50 Δ | CPU Δ | base p50 ms | cand p50 ms |
| --- | ---: | ---: | ---: | ---: |
| cpu_by_service_1h | +22% | -13% | 23.0 | 28.0 |
| error_ratio_1h | +24% | -6% | 178.0 | 220.0 |
| gc_pause_median_1h | -12% | -14% | 17.0 | 15.0 |
| latency_p95_by_route_1h | +5% | -0% | 144.0 | 151.0 |
| memory_by_pod_1h | +62% | +27% | 34.0 | 55.0 |
| request_rate_by_route_1h | -20% | -14% | 339.0 | 272.0 |
| rpc_count_by_method_1h | -45% | -39% | 155.0 | 86.0 |
| series_by_metric | -14% | -20% | 286.0 | 247.0 |
| user_requests | -62% | -42% | 26.0 | 10.0 |

Ingest: base 46267 events/s, candidate 21932 events/s (Δ -53%)

## Value parity (window 2026-05-31 22:15:00 → 2026-06-01 00:00:01)

| Query | result | rows |
| --- | --- | ---: |
| cpu_by_service_1h | MATCH | 1800 |
| error_ratio_1h | MATCH | 19 |
| gc_pause_median_1h | MATCH | 7 |
| latency_p95_by_route_1h | MATCH | 40 |
| memory_by_pod_1h | MATCH | 90 |
| parity/exponential_histogram_rows | MATCH | 1 |
| parity/gauge_rows | MATCH | 1 |
| parity/histogram_rows | MATCH | 1 |
| parity/sum_rows | MATCH | 1 |
| parity/summary_rows | MATCH | 1 |
| request_rate_by_route_1h | MATCH | 4004 |
| rpc_count_by_method_1h | MATCH | 95 |
| series_by_metric | MATCH | 7 |
| user_requests | MATCH | 0 |

MATCH=14 DIFFERS=0 (one side only=0)

compare exit status: 0 (non-zero = parity failed or not established)

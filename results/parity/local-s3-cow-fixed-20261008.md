# Local cross-target parity (2026-10-08)

Smoke and parity only, never performance. Targets: altinity/local-single altinity/local-cluster-s3-cow. Window ends 2026-06-01T00:00:00Z; seed 42;
openmeter 1000000 rows, OTel 500000 rows per scenario. Each section is `bench compare` output.

---

## Compare openmeter/baseline-openmeter

> **Smoke, not performance.** At least one side is a local tier, or the sides are on different tiers. Latency and CPU deltas below are not evidence about the platforms.

| Side | Target | Version | Topology | Run | Commit |
| --- | --- | --- | --- | --- | --- |
| altinity/local-single | altinity/local-single (local-single) | 26.3.33.10001.altinitystable | single 1x1 local | 20261008T083418Z-018a | a9d9670 |
| altinity/local-cluster-s3-cow | altinity/local-cluster-s3-cow (local-cluster) | 26.3.33.10001.altinitystable | replicated 1x2 s3 | 20261008T083647Z-6f45 | a9d9670 |

## altinity/local-single → altinity/local-cluster-s3-cow

| Query | p50 Δ | CPU Δ | base p50 ms | cand p50 ms |
| --- | ---: | ---: | ---: | ---: |
| agent_runs_by_name | -59% | -59% | 29.0 | 12.0 |
| avg_hour | -8% | -35% | 39.0 | 36.0 |
| count_hour | -24% | -30% | 17.0 | 13.0 |
| count_total | -7% | +13% | 14.0 | 13.0 |
| distinct_subjects | -19% | +20% | 31.0 | 25.0 |
| kong_api_request_total | +0% | +23% | 12.0 | 12.0 |
| kong_llm_tokens_total | -38% | -15% | 84.0 | 52.0 |
| kong_status_by_route | -33% | -28% | 58.0 | 39.0 |
| latest_hour | -17% | -13% | 47.0 | 39.0 |
| llm_tokens_by_model | +3% | -1% | 29.0 | 30.0 |
| max_hour | +3% | -7% | 33.0 | 34.0 |
| min_hour | +39% | -2% | 33.0 | 46.0 |
| sum_day | +0% | -4% | 36.0 | 36.0 |
| sum_day_by_dim | +35% | +11% | 23.0 | 31.0 |
| sum_day_by_subject | -32% | -16% | 50.0 | 34.0 |
| sum_hour | +0% | -13% | 38.0 | 38.0 |
| sum_hour_group1 | +3% | -30% | 34.0 | 35.0 |
| sum_hour_group1_group2 | -35% | -33% | 46.0 | 30.0 |
| sum_hour_group1_group2_no_prewhere | -41% | -37% | 49.0 | 29.0 |
| sum_hour_group1_no_prewhere | +16% | -17% | 32.0 | 37.0 |
| sum_hour_tz | +29% | -11% | 35.0 | 45.0 |
| sum_minute | -24% | -19% | 59.0 | 45.0 |
| sum_month | -21% | -21% | 43.0 | 34.0 |
| sum_no_window | -21% | -21% | 42.0 | 33.0 |
| sum_total | -22% | -23% | 37.0 | 29.0 |
| sum_total_by_subject | -11% | -20% | 36.0 | 32.0 |
| sum_total_filter1 | -9% | -31% | 34.0 | 31.0 |
| sum_total_filter2 | -26% | -21% | 42.0 | 31.0 |
| unique_count_hour | -18% | -26% | 34.0 | 28.0 |
| workload_seconds_by_region | +7% | -16% | 15.0 | 16.0 |

Ingest: base 100342 events/s, candidate 52204 events/s (Δ -48%)

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
| altinity/local-single | altinity/local-single (local-single) | 26.3.33.10001.altinitystable | single 1x1 local | 20261008T083438Z-76d5 | a9d9670 |
| altinity/local-cluster-s3-cow | altinity/local-cluster-s3-cow (local-cluster) | 26.3.33.10001.altinitystable | replicated 1x2 s3 | 20261008T083722Z-897a | a9d9670 |

## altinity/local-single → altinity/local-cluster-s3-cow

| Query | p50 Δ | CPU Δ | base p50 ms | cand p50 ms |
| --- | ---: | ---: | ---: | ---: |
| agent_runs_by_name | -15% | -14% | 20.0 | 17.0 |
| avg_hour | -33% | -15% | 43.0 | 29.0 |
| count_hour | +7% | -3% | 15.0 | 16.0 |
| count_total | -22% | -18% | 18.0 | 14.0 |
| distinct_subjects | +22% | +28% | 23.0 | 28.0 |
| kong_api_request_by_all_dims | -18% | +14% | 130.0 | 107.0 |
| kong_api_request_by_method | +10% | +24% | 20.0 | 22.0 |
| kong_api_request_by_service | -38% | -20% | 37.0 | 23.0 |
| kong_api_request_total | +8% | +11% | 12.0 | 13.0 |
| kong_llm_tokens_total | +29% | +10% | 24.0 | 31.0 |
| kong_status_by_route | +24% | +13% | 21.0 | 26.0 |
| latest_hour | -11% | +4% | 38.0 | 34.0 |
| llm_tokens_by_model | -4% | -1% | 25.0 | 24.0 |
| lookup_by_id | +14% | +16% | 80.0 | 91.0 |
| max_hour | -6% | +1% | 31.0 | 29.0 |
| min_hour | -3% | +19% | 30.0 | 29.0 |
| sum_day | +17% | +11% | 24.0 | 28.0 |
| sum_day_by_dim | +5% | -2% | 21.0 | 22.0 |
| sum_day_by_subject | -24% | -6% | 34.0 | 26.0 |
| sum_hour | +8% | +6% | 26.0 | 28.0 |
| sum_hour_group1 | +21% | +18% | 28.0 | 34.0 |
| sum_hour_group1_group2 | +3% | +8% | 37.0 | 38.0 |
| sum_hour_group1_group2_no_prewhere | -8% | +2% | 36.0 | 33.0 |
| sum_hour_group1_no_prewhere | +4% | -5% | 28.0 | 29.0 |
| sum_hour_tz | +12% | -5% | 32.0 | 36.0 |
| sum_minute | +8% | -8% | 39.0 | 42.0 |
| sum_month | +19% | +6% | 32.0 | 38.0 |
| sum_no_window | +31% | +15% | 26.0 | 34.0 |
| sum_total | +7% | +18% | 27.0 | 29.0 |
| sum_total_by_subject | +0% | +4% | 29.0 | 29.0 |
| sum_total_filter1 | +33% | +11% | 27.0 | 36.0 |
| sum_total_filter2 | +16% | +16% | 31.0 | 36.0 |
| unique_count_hour | +0% | -9% | 30.0 | 30.0 |
| workload_seconds_by_region | +25% | +10% | 20.0 | 25.0 |

Ingest: base 57103 events/s, candidate 33035 events/s (Δ -42%)

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
| altinity/local-single | altinity/local-single (local-single) | 26.3.33.10001.altinitystable | single 1x1 local | 20261008T083508Z-c88f | a9d9670 |
| altinity/local-cluster-s3-cow | altinity/local-cluster-s3-cow (local-cluster) | 26.3.33.10001.altinitystable | replicated 1x2 s3 | 20261008T083808Z-b8d6 | a9d9670 |

## altinity/local-single → altinity/local-cluster-s3-cow

| Query | p50 Δ | CPU Δ | base p50 ms | cand p50 ms |
| --- | ---: | ---: | ---: | ---: |
| error_rate_24h | -7% | +25% | 14.0 | 13.0 |
| latency_by_service_1h | +12% | +30% | 16.0 | 18.0 |
| rollup_values_24h | +18% | -13% | 11.0 | 13.0 |
| route_status_24h | -17% | -21% | 23.0 | 19.0 |
| route_status_items_24h | -18% | -16% | 17.0 | 14.0 |
| service_dependencies_1h | -27% | +22% | 55.0 | 40.0 |
| slowest_traces_1h | +36% | +32% | 11.0 | 15.0 |
| spans_by_hour | -6% | +4% | 32.0 | 30.0 |
| top_operations_24h | +30% | -5% | 10.0 | 13.0 |
| trace_by_id | -21% | -7% | 14.0 | 11.0 |
| user_activity | +77% | +28% | 13.0 | 23.0 |

Ingest: base 38267 events/s, candidate 15377 events/s (Δ -60%)

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
| altinity/local-single | altinity/local-single (local-single) | 26.3.33.10001.altinitystable | single 1x1 local | 20261008T083525Z-8321 | a9d9670 |
| altinity/local-cluster-s3-cow | altinity/local-cluster-s3-cow (local-cluster) | 26.3.33.10001.altinitystable | replicated 1x2 s3 | 20261008T083848Z-da18 | a9d9670 |

## altinity/local-single → altinity/local-cluster-s3-cow

| Query | p50 Δ | CPU Δ | base p50 ms | cand p50 ms |
| --- | ---: | ---: | ---: | ---: |
| error_rate_24h | -24% | -16% | 17.0 | 13.0 |
| latency_by_service_1h | +6% | +22% | 17.0 | 18.0 |
| route_status_24h | -30% | -16% | 33.0 | 23.0 |
| service_dependencies_1h | +18% | +16% | 39.0 | 46.0 |
| slowest_traces_1h | +20% | +11% | 15.0 | 18.0 |
| spans_by_hour | +31% | +15% | 26.0 | 34.0 |
| top_operations_24h | +20% | +21% | 10.0 | 12.0 |
| trace_by_id | +46% | +44% | 13.0 | 19.0 |
| trace_by_id_lookup | +21% | +13% | 28.0 | 34.0 |
| user_activity | -35% | -6% | 26.0 | 17.0 |

Ingest: base 45727 events/s, candidate 19469 events/s (Δ -57%)

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
| altinity/local-single | altinity/local-single (local-single) | 26.3.33.10001.altinitystable | single 1x1 local | 20261008T083540Z-1281 | a9d9670 |
| altinity/local-cluster-s3-cow | altinity/local-cluster-s3-cow (local-cluster) | 26.3.33.10001.altinitystable | replicated 1x2 s3 | 20261008T083919Z-c2df | a9d9670 |

## altinity/local-single → altinity/local-cluster-s3-cow

| Query | p50 Δ | CPU Δ | base p50 ms | cand p50 ms |
| --- | ---: | ---: | ---: | ---: |
| body_contains_24h | +50% | +15% | 18.0 | 27.0 |
| logs_for_trace | +12% | -4% | 8.0 | 9.0 |
| needle_search | +86% | +99% | 7.0 | 13.0 |
| rollup_services_24h | +37% | +19% | 8.0 | 11.0 |
| service_tail_1h | +50% | +28% | 10.0 | 15.0 |
| severity_by_service_24h | +36% | +31% | 14.0 | 19.0 |
| top_error_patterns_24h | +9% | +6% | 22.0 | 24.0 |
| user_logs | +7% | +22% | 14.0 | 15.0 |
| user_logs_items | +14% | +31% | 14.0 | 16.0 |
| volume_by_hour | +41% | +16% | 22.0 | 31.0 |

Ingest: base 29158 events/s, candidate 14577 events/s (Δ -50%)

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
| altinity/local-single | altinity/local-single (local-single) | 26.3.33.10001.altinitystable | single 1x1 local | 20261008T083600Z-5322 | a9d9670 |
| altinity/local-cluster-s3-cow | altinity/local-cluster-s3-cow (local-cluster) | 26.3.33.10001.altinitystable | replicated 1x2 s3 | 20261008T083959Z-03e4 | a9d9670 |

## altinity/local-single → altinity/local-cluster-s3-cow

| Query | p50 Δ | CPU Δ | base p50 ms | cand p50 ms |
| --- | ---: | ---: | ---: | ---: |
| body_contains_24h | +9% | -4% | 23.0 | 25.0 |
| logs_for_trace | -20% | -1% | 10.0 | 8.0 |
| needle_search | +0% | +43% | 7.0 | 7.0 |
| service_tail_1h | +18% | +28% | 11.0 | 13.0 |
| severity_by_service_24h | -11% | -21% | 18.0 | 16.0 |
| top_error_patterns_24h | -22% | -27% | 23.0 | 18.0 |
| user_logs | +0% | -3% | 14.0 | 14.0 |
| volume_by_hour | +9% | -8% | 23.0 | 25.0 |

Ingest: base 39756 events/s, candidate 19488 events/s (Δ -51%)

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
| altinity/local-single | altinity/local-single (local-single) | 26.3.33.10001.altinitystable | single 1x1 local | 20261008T083616Z-59ea | a9d9670 |
| altinity/local-cluster-s3-cow | altinity/local-cluster-s3-cow (local-cluster) | 26.3.33.10001.altinitystable | replicated 1x2 s3 | 20261008T084030Z-816b | a9d9670 |

## altinity/local-single → altinity/local-cluster-s3-cow

| Query | p50 Δ | CPU Δ | base p50 ms | cand p50 ms |
| --- | ---: | ---: | ---: | ---: |
| cpu_by_service_1h | +58% | +44% | 12.0 | 19.0 |
| error_ratio_1h | -5% | +36% | 147.0 | 139.0 |
| gc_pause_median_1h | +22% | +27% | 9.0 | 11.0 |
| latency_p95_by_route_1h | -16% | +33% | 112.0 | 94.0 |
| memory_by_pod_1h | +22% | +19% | 23.0 | 28.0 |
| request_rate_by_route_1h | -28% | +2% | 179.0 | 128.0 |
| rpc_count_by_method_1h | +27% | +17% | 74.0 | 94.0 |
| series_by_metric | +10% | +28% | 143.0 | 157.0 |
| user_requests | +100% | +77% | 5.0 | 10.0 |

Ingest: base 48080 events/s, candidate 15173 events/s (Δ -68%)

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
| altinity/local-single | altinity/local-single (local-single) | 26.3.33.10001.altinitystable | single 1x1 local | 20261008T083632Z-5b8a | a9d9670 |
| altinity/local-cluster-s3-cow | altinity/local-cluster-s3-cow (local-cluster) | 26.3.33.10001.altinitystable | replicated 1x2 s3 | 20261008T084112Z-3eba | a9d9670 |

## altinity/local-single → altinity/local-cluster-s3-cow

| Query | p50 Δ | CPU Δ | base p50 ms | cand p50 ms |
| --- | ---: | ---: | ---: | ---: |
| cpu_by_service_1h | +8% | +11% | 12.0 | 13.0 |
| error_ratio_1h | -30% | +28% | 150.0 | 105.0 |
| gc_pause_median_1h | +50% | +43% | 8.0 | 12.0 |
| latency_p95_by_route_1h | -20% | +21% | 132.0 | 106.0 |
| memory_by_pod_1h | +19% | +19% | 27.0 | 32.0 |
| request_rate_by_route_1h | -28% | -0% | 172.0 | 124.0 |
| rpc_count_by_method_1h | -15% | -15% | 80.0 | 68.0 |
| series_by_metric | +6% | +29% | 154.0 | 164.0 |
| user_requests | +50% | +47% | 8.0 | 12.0 |

Ingest: base 54450 events/s, candidate 15753 events/s (Δ -71%)

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

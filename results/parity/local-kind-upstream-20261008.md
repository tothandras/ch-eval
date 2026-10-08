# Local cross-target parity (2026-10-08)

Smoke and parity only, never performance. Targets: upstream-oss/local-single upstream-oss/kind. Window ends 2026-06-01T00:00:00Z; seed 42;
openmeter 1000000 rows, OTel 500000 rows per scenario. Each section is `bench compare` output.

---

## Compare openmeter/baseline-openmeter

> **Smoke, not performance.** At least one side is a local tier, or the sides are on different tiers. Latency and CPU deltas below are not evidence about the platforms.

| Side | Target | Version | Topology | Run | Commit |
| --- | --- | --- | --- | --- | --- |
| upstream-oss/local-single | upstream-oss/local-single (local-single) | 26.3.33.24 | single 1x1 local | 20261008T081938Z-2caf | effdf02 |
| upstream-oss/kind | upstream-oss/kind (kind) | 26.3.33.24 | replicated 1x2 local | 20261008T082217Z-b152 | effdf02 |

## upstream-oss/local-single → upstream-oss/kind

| Query | p50 Δ | CPU Δ | base p50 ms | cand p50 ms |
| --- | ---: | ---: | ---: | ---: |
| agent_runs_by_name | -27% | -22% | 15.0 | 11.0 |
| avg_hour | +57% | +28% | 30.0 | 47.0 |
| count_hour | -11% | -8% | 18.0 | 16.0 |
| count_total | +0% | +3% | 13.0 | 13.0 |
| distinct_subjects | -27% | -24% | 30.0 | 22.0 |
| kong_api_request_total | +8% | -8% | 13.0 | 14.0 |
| kong_llm_tokens_total | -16% | -17% | 95.0 | 80.0 |
| kong_status_by_route | +10% | +10% | 42.0 | 46.0 |
| latest_hour | -2% | -13% | 50.0 | 49.0 |
| llm_tokens_by_model | -3% | -8% | 39.0 | 38.0 |
| max_hour | -8% | -7% | 36.0 | 33.0 |
| min_hour | -20% | -15% | 46.0 | 37.0 |
| sum_day | -38% | -8% | 61.0 | 38.0 |
| sum_day_by_dim | -23% | -27% | 35.0 | 27.0 |
| sum_day_by_subject | -24% | -20% | 49.0 | 37.0 |
| sum_hour | -22% | -12% | 55.0 | 43.0 |
| sum_hour_group1 | +22% | +16% | 36.0 | 44.0 |
| sum_hour_group1_group2 | +27% | +7% | 33.0 | 42.0 |
| sum_hour_group1_group2_no_prewhere | +5% | +7% | 43.0 | 45.0 |
| sum_hour_group1_no_prewhere | +21% | +7% | 39.0 | 47.0 |
| sum_hour_tz | -7% | -8% | 30.0 | 28.0 |
| sum_minute | +15% | +10% | 41.0 | 47.0 |
| sum_month | +36% | +28% | 33.0 | 45.0 |
| sum_no_window | -5% | +1% | 37.0 | 35.0 |
| sum_total | +13% | +3% | 31.0 | 35.0 |
| sum_total_by_subject | +7% | -6% | 29.0 | 31.0 |
| sum_total_filter1 | -8% | +6% | 39.0 | 36.0 |
| sum_total_filter2 | +13% | +13% | 30.0 | 34.0 |
| unique_count_hour | -15% | +10% | 46.0 | 39.0 |
| workload_seconds_by_region | +19% | +3% | 16.0 | 19.0 |

Ingest: base 89863 events/s, candidate 60629 events/s (Δ -33%)

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
| upstream-oss/local-single | upstream-oss/local-single (local-single) | 26.3.33.24 | single 1x1 local | 20261008T082000Z-2e46 | effdf02 |
| upstream-oss/kind | upstream-oss/kind (kind) | 26.3.33.24 | replicated 1x2 local | 20261008T082248Z-df1a | effdf02 |

## upstream-oss/local-single → upstream-oss/kind

| Query | p50 Δ | CPU Δ | base p50 ms | cand p50 ms |
| --- | ---: | ---: | ---: | ---: |
| agent_runs_by_name | +18% | +14% | 17.0 | 20.0 |
| avg_hour | +46% | +51% | 24.0 | 35.0 |
| count_hour | -14% | -5% | 14.0 | 12.0 |
| count_total | +27% | +30% | 11.0 | 14.0 |
| distinct_subjects | +0% | +4% | 23.0 | 23.0 |
| kong_api_request_by_all_dims | +16% | +13% | 111.0 | 129.0 |
| kong_api_request_by_method | +11% | +9% | 19.0 | 21.0 |
| kong_api_request_by_service | -16% | +3% | 32.0 | 27.0 |
| kong_api_request_total | +0% | -8% | 11.0 | 11.0 |
| kong_llm_tokens_total | +0% | +2% | 24.0 | 24.0 |
| kong_status_by_route | -13% | -3% | 23.0 | 20.0 |
| latest_hour | +7% | -0% | 42.0 | 45.0 |
| llm_tokens_by_model | +22% | -8% | 27.0 | 33.0 |
| lookup_by_id | +10% | +9% | 71.0 | 78.0 |
| max_hour | +3% | +4% | 33.0 | 34.0 |
| min_hour | -12% | -8% | 34.0 | 30.0 |
| sum_day | +7% | +1% | 30.0 | 32.0 |
| sum_day_by_dim | +12% | +8% | 25.0 | 28.0 |
| sum_day_by_subject | +20% | +13% | 25.0 | 30.0 |
| sum_hour | +3% | +2% | 31.0 | 32.0 |
| sum_hour_group1 | +3% | +27% | 35.0 | 36.0 |
| sum_hour_group1_group2 | -8% | -2% | 37.0 | 34.0 |
| sum_hour_group1_group2_no_prewhere | +12% | +9% | 32.0 | 36.0 |
| sum_hour_group1_no_prewhere | -24% | +10% | 38.0 | 29.0 |
| sum_hour_tz | -5% | +36% | 44.0 | 42.0 |
| sum_minute | -8% | +14% | 40.0 | 37.0 |
| sum_month | +14% | +9% | 29.0 | 33.0 |
| sum_no_window | -10% | -2% | 29.0 | 26.0 |
| sum_total | +0% | -0% | 24.0 | 24.0 |
| sum_total_by_subject | -49% | -18% | 37.0 | 19.0 |
| sum_total_filter1 | +15% | +5% | 27.0 | 31.0 |
| sum_total_filter2 | -8% | -6% | 36.0 | 33.0 |
| unique_count_hour | -16% | -7% | 31.0 | 26.0 |
| workload_seconds_by_region | +8% | +19% | 25.0 | 27.0 |

Ingest: base 53828 events/s, candidate 39747 events/s (Δ -26%)

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
| upstream-oss/local-single | upstream-oss/local-single (local-single) | 26.3.33.24 | single 1x1 local | 20261008T082030Z-0ffa | effdf02 |
| upstream-oss/kind | upstream-oss/kind (kind) | 26.3.33.24 | replicated 1x2 local | 20261008T082329Z-b03e | effdf02 |

## upstream-oss/local-single → upstream-oss/kind

| Query | p50 Δ | CPU Δ | base p50 ms | cand p50 ms |
| --- | ---: | ---: | ---: | ---: |
| error_rate_24h | -7% | -6% | 15.0 | 14.0 |
| latency_by_service_1h | -23% | -28% | 22.0 | 17.0 |
| rollup_values_24h | +8% | +8% | 13.0 | 14.0 |
| route_status_24h | -10% | -5% | 21.0 | 19.0 |
| route_status_items_24h | +6% | +23% | 16.0 | 17.0 |
| service_dependencies_1h | +0% | -3% | 51.0 | 51.0 |
| slowest_traces_1h | +133% | +23% | 12.0 | 28.0 |
| spans_by_hour | +8% | -1% | 25.0 | 27.0 |
| top_operations_24h | +37% | +21% | 8.0 | 11.0 |
| trace_by_id | -33% | -23% | 15.0 | 10.0 |
| user_activity | +31% | -7% | 16.0 | 21.0 |

Ingest: base 33628 events/s, candidate 20916 events/s (Δ -38%)

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
| upstream-oss/local-single | upstream-oss/local-single (local-single) | 26.3.33.24 | single 1x1 local | 20261008T082049Z-cd35 | effdf02 |
| upstream-oss/kind | upstream-oss/kind (kind) | 26.3.33.24 | replicated 1x2 local | 20261008T082359Z-481d | effdf02 |

## upstream-oss/local-single → upstream-oss/kind

| Query | p50 Δ | CPU Δ | base p50 ms | cand p50 ms |
| --- | ---: | ---: | ---: | ---: |
| error_rate_24h | +0% | -12% | 15.0 | 15.0 |
| latency_by_service_1h | +0% | -18% | 17.0 | 17.0 |
| route_status_24h | -7% | +6% | 27.0 | 25.0 |
| service_dependencies_1h | +8% | +0% | 51.0 | 55.0 |
| slowest_traces_1h | +33% | +29% | 15.0 | 20.0 |
| spans_by_hour | -10% | +9% | 31.0 | 28.0 |
| top_operations_24h | +8% | +1% | 12.0 | 13.0 |
| trace_by_id | -11% | +16% | 18.0 | 16.0 |
| trace_by_id_lookup | +10% | -1% | 29.0 | 32.0 |
| user_activity | -5% | -9% | 19.0 | 18.0 |

Ingest: base 44129 events/s, candidate 22664 events/s (Δ -49%)

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
| upstream-oss/local-single | upstream-oss/local-single (local-single) | 26.3.33.24 | single 1x1 local | 20261008T082105Z-9d2e | effdf02 |
| upstream-oss/kind | upstream-oss/kind (kind) | 26.3.33.24 | replicated 1x2 local | 20261008T082427Z-43d3 | effdf02 |

## upstream-oss/local-single → upstream-oss/kind

| Query | p50 Δ | CPU Δ | base p50 ms | cand p50 ms |
| --- | ---: | ---: | ---: | ---: |
| body_contains_24h | -3% | +9% | 29.0 | 28.0 |
| logs_for_trace | -23% | -20% | 13.0 | 10.0 |
| needle_search | -22% | -20% | 9.0 | 7.0 |
| rollup_services_24h | +37% | +18% | 8.0 | 11.0 |
| service_tail_1h | +36% | +8% | 11.0 | 15.0 |
| severity_by_service_24h | +59% | +13% | 17.0 | 27.0 |
| top_error_patterns_24h | -15% | -13% | 26.0 | 22.0 |
| user_logs | +36% | +34% | 14.0 | 19.0 |
| user_logs_items | +14% | +6% | 14.0 | 16.0 |
| volume_by_hour | +8% | -3% | 24.0 | 26.0 |

Ingest: base 28231 events/s, candidate 16150 events/s (Δ -43%)

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
| upstream-oss/local-single | upstream-oss/local-single (local-single) | 26.3.33.24 | single 1x1 local | 20261008T082126Z-1450 | effdf02 |
| upstream-oss/kind | upstream-oss/kind (kind) | 26.3.33.24 | replicated 1x2 local | 20261008T082503Z-5717 | effdf02 |

## upstream-oss/local-single → upstream-oss/kind

| Query | p50 Δ | CPU Δ | base p50 ms | cand p50 ms |
| --- | ---: | ---: | ---: | ---: |
| body_contains_24h | +26% | +1% | 23.0 | 29.0 |
| logs_for_trace | +86% | +30% | 7.0 | 13.0 |
| needle_search | +0% | +21% | 10.0 | 10.0 |
| service_tail_1h | +5% | +7% | 19.0 | 20.0 |
| severity_by_service_24h | +20% | +40% | 20.0 | 24.0 |
| top_error_patterns_24h | +62% | +12% | 21.0 | 34.0 |
| user_logs | +13% | +7% | 15.0 | 17.0 |
| volume_by_hour | +14% | +5% | 22.0 | 25.0 |

Ingest: base 38137 events/s, candidate 24749 events/s (Δ -35%)

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
| upstream-oss/local-single | upstream-oss/local-single (local-single) | 26.3.33.24 | single 1x1 local | 20261008T082142Z-703f | effdf02 |
| upstream-oss/kind | upstream-oss/kind (kind) | 26.3.33.24 | replicated 1x2 local | 20261008T082527Z-22b2 | effdf02 |

## upstream-oss/local-single → upstream-oss/kind

| Query | p50 Δ | CPU Δ | base p50 ms | cand p50 ms |
| --- | ---: | ---: | ---: | ---: |
| cpu_by_service_1h | +38% | +21% | 13.0 | 18.0 |
| error_ratio_1h | +23% | +21% | 159.0 | 195.0 |
| gc_pause_median_1h | +0% | -9% | 12.0 | 12.0 |
| latency_p95_by_route_1h | -6% | -11% | 158.0 | 148.0 |
| memory_by_pod_1h | +59% | +5% | 34.0 | 54.0 |
| request_rate_by_route_1h | -9% | +4% | 268.0 | 245.0 |
| rpc_count_by_method_1h | +14% | +15% | 76.0 | 87.0 |
| series_by_metric | +21% | +5% | 189.0 | 229.0 |
| user_requests | -60% | -43% | 15.0 | 6.0 |

Ingest: base 47820 events/s, candidate 22949 events/s (Δ -52%)

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
| upstream-oss/local-single | upstream-oss/local-single (local-single) | 26.3.33.24 | single 1x1 local | 20261008T082159Z-a408 | effdf02 |
| upstream-oss/kind | upstream-oss/kind (kind) | 26.3.33.24 | replicated 1x2 local | 20261008T082558Z-e72d | effdf02 |

## upstream-oss/local-single → upstream-oss/kind

| Query | p50 Δ | CPU Δ | base p50 ms | cand p50 ms |
| --- | ---: | ---: | ---: | ---: |
| cpu_by_service_1h | -21% | -14% | 14.0 | 11.0 |
| error_ratio_1h | -11% | -7% | 164.0 | 146.0 |
| gc_pause_median_1h | +0% | -0% | 8.0 | 8.0 |
| latency_p95_by_route_1h | -13% | -3% | 143.0 | 124.0 |
| memory_by_pod_1h | -24% | -17% | 34.0 | 26.0 |
| request_rate_by_route_1h | -10% | -8% | 218.0 | 197.0 |
| rpc_count_by_method_1h | +1% | +3% | 68.0 | 69.0 |
| series_by_metric | -23% | -18% | 188.0 | 145.0 |
| user_requests | +14% | -9% | 7.0 | 8.0 |

Ingest: base 44716 events/s, candidate 22275 events/s (Δ -50%)

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

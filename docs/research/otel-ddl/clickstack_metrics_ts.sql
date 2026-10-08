-- ClickStack DDL: ClickStack metrics_ts (TimeSeries engine, PromQL)
-- Source: https://github.com/hyperdxio/hyperdx/blob/eeabaebe6e9226f5ca090b29566e44ec52523b01/docker/otel-collector/schema/seed/00008_otel_metrics_timeseries.sql
-- Version: hyperdxio/hyperdx main @ eeabaebe6e9226f5ca090b29566e44ec52523b01 (2026-10-07)
-- Applied by the ClickStack collector image's goose-based `migrate` tool (packages/otel-collector/cmd/migrate/main.go)
--   before the collector starts; the exporter itself runs with create_schema: false unless JSON/legacy mode is enabled.
-- Macros: ${DATABASE} (env HYPERDX_OTEL_EXPORTER_CLICKHOUSE_DATABASE, default 'default');
--   ${LOGS_TTL}/${TRACES_TTL}/${METRICS_TTL} (env HYPERDX_OTEL_EXPORTER_{LOGS,TRACES,METRICS}_TTL, falling back to
--   HYPERDX_OTEL_EXPORTER_TABLES_TTL, default 720h) rendered as toIntervalDay(30) etc.
-- Only `-- +goose ...` annotation lines were removed; SQL text is verbatim.
-- ===== BEGIN VERBATIM =====
CREATE TABLE IF NOT EXISTS ${DATABASE}.metrics_ts
ENGINE = TimeSeries
SETTINGS allow_experimental_time_series_table = 1;
-- ===== END VERBATIM =====

-- Verbatim from the brief (section 8.2).
CREATE TABLE IF NOT EXISTS _usage_event_field_values_backfill (
  partition_id String
)
ENGINE = ReplacingMergeTree
ORDER BY partition_id;

-- Verbatim from the brief (section 8.2).
CREATE TABLE usage_event_field_values
(
    `tenant_id` UUID,
    `field` LowCardinality(String),
    `key` String,
    `value` String
)
ENGINE = ReplacingMergeTree
ORDER BY (tenant_id, field, key, value)
SETTINGS index_granularity = 8192;

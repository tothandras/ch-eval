-- Verbatim from the brief (section 8.1): SHOW CREATE TABLE output after replaying the team's migrations on 26.7.4.
CREATE TABLE usage_events
(
    `tenant_id` UUID,
    `id` String,
    `type` LowCardinality(String),
    `source` String,
    `subject` String DEFAULT '',
    `time` DateTime64(3, 'UTC'),
    `traceparent` String DEFAULT '',
    `tracestate` String DEFAULT '',
    `ingested_at` DateTime64(3, 'UTC'),
    `stored_at` DateTime64(3, 'UTC') DEFAULT now64(3),
    `kafka_topic` LowCardinality(String) DEFAULT '',
    `kafka_partition` UInt64 DEFAULT 0,
    `kafka_offset` UInt64 DEFAULT 0,
    `arrival` LowCardinality(String) MATERIALIZED if(ingested_at > (time + toIntervalHour(1)), 'late', 'on_time'),
    `provider` LowCardinality(String) DEFAULT '',
    `model` LowCardinality(String) DEFAULT '',
    `model_id` LowCardinality(String) DEFAULT '',
    `request_mode` LowCardinality(String) DEFAULT '',
    `region` LowCardinality(String) DEFAULT '',
    `client` LowCardinality(String) DEFAULT '',
    `input_tokens` UInt64 DEFAULT 0,
    `input_text_tokens` UInt64 DEFAULT 0,
    `input_audio_tokens` UInt64 DEFAULT 0,
    `input_image_tokens` UInt64 DEFAULT 0,
    `input_cache_read_tokens` UInt64 DEFAULT 0,
    `input_cache_read_text_tokens` UInt64 DEFAULT 0,
    `input_cache_read_audio_tokens` UInt64 DEFAULT 0,
    `input_cache_read_image_tokens` UInt64 DEFAULT 0,
    `input_cache_read_video_tokens` UInt64 DEFAULT 0,
    `input_cache_write_tokens` UInt64 DEFAULT 0,
    `input_cache_write_text_tokens` UInt64 DEFAULT 0,
    `input_cache_write_audio_tokens` UInt64 DEFAULT 0,
    `input_cache_write_image_tokens` UInt64 DEFAULT 0,
    `input_cache_write_video_tokens` UInt64 DEFAULT 0,
    `output_tokens` UInt64 DEFAULT 0,
    `output_text_tokens` UInt64 DEFAULT 0,
    `output_audio_tokens` UInt64 DEFAULT 0,
    `output_reasoning_tokens` UInt64 DEFAULT 0,
    `accepted_prediction_tokens` UInt64 DEFAULT 0,
    `rejected_prediction_tokens` UInt64 DEFAULT 0,
    `realtime_audio_ms` UInt64 DEFAULT 0,
    `reported_total_tokens` UInt64 DEFAULT 0,
    `semantic_cache_status` LowCardinality(String) DEFAULT '',
    `semantic_cache_embeddings_provider` LowCardinality(String) DEFAULT '',
    `semantic_cache_embeddings_model` LowCardinality(String) DEFAULT '',
    `semantic_cache_embeddings_tokens` UInt64 DEFAULT 0,
    `compression_status` LowCardinality(String) DEFAULT '',
    `compression_input_tokens_before` UInt64 DEFAULT 0,
    `reported_cost_usd` Decimal(18, 9) DEFAULT 0,
    `reported_cache_cost_usd` Decimal(18, 9) DEFAULT 0,
    `teams` Array(String) DEFAULT [],
    `workspace_id` UUID DEFAULT '00000000-0000-0000-0000-000000000000',
    `service_id` UUID DEFAULT '00000000-0000-0000-0000-000000000000',
    `route_id` UUID DEFAULT '00000000-0000-0000-0000-000000000000',
    `plugin_id` UUID DEFAULT '00000000-0000-0000-0000-000000000000',
    `consumer_id` UUID DEFAULT '00000000-0000-0000-0000-000000000000',
    `consumer_group_ids` Array(UUID) DEFAULT [],
    `application_id` UUID DEFAULT '00000000-0000-0000-0000-000000000000',
    `request_id` String DEFAULT '',
    `tags` Map(LowCardinality(String), String) DEFAULT map(),
    `tag_items` Array(String) MATERIALIZED arrayMap(pair -> concat(base64Encode(pair.1), ':', base64Encode(pair.2)), CAST(tags, 'Array(Tuple(String, String))')) CODEC(ZSTD(1)),
    INDEX idx_id id TYPE bloom_filter(0.01) GRANULARITY 4,
    INDEX idx_tag_items tag_items TYPE text(tokenizer = 'array') GRANULARITY 100000000,
    INDEX idx_teams teams TYPE text(tokenizer = 'array') GRANULARITY 100000000,
    INDEX idx_ingested_at ingested_at TYPE minmax GRANULARITY 4,
    PROJECTION by_model
    (
        SELECT _part_offset
        ORDER BY
            tenant_id,
            model,
            toStartOfMinute(time)
    ),
    PROJECTION by_subject
    (
        SELECT _part_offset
        ORDER BY
            tenant_id,
            subject,
            toStartOfMinute(time)
    )
)
ENGINE = ReplacingMergeTree
PARTITION BY toYYYYMM(time)
PRIMARY KEY (tenant_id, toStartOfMinute(time))
ORDER BY (tenant_id, toStartOfMinute(time), time, source, id)
SETTINGS non_replicated_deduplication_window = 1000, deduplicate_merge_projection_mode = 'rebuild', allow_part_offset_column_in_projections = 1, map_serialization_version = 'with_buckets', map_buckets_strategy = 'sqrt', max_buckets_in_map = 32, index_granularity = 8192;

CREATE TABLE IF NOT EXISTS openmeter.proposal_events_local ON CLUSTER 'ch_eval' (
  namespace String,
  id String,
  type LowCardinality(String),
  subject String,
  source String,
  time DateTime,
  data JSON CODEC(ZSTD(3)),                                                -- ← (1) + (2)
  ingested_at DateTime,
  stored_at DateTime,
  INDEX om_events_stored_at stored_at TYPE minmax GRANULARITY 4,
  INDEX om_events_id_bloom  id        TYPE bloom_filter(0.01) GRANULARITY 1,  -- ← (3)
  store_row_id String
) ENGINE = ReplicatedMergeTree('/clickhouse/tables/{shard}/{database}/{table}', '{replica}')
PARTITION BY toYYYYMM(time)
ORDER BY (namespace, type, subject, toStartOfHour(time));

CREATE TABLE IF NOT EXISTS openmeter.proposal_events ON CLUSTER 'ch_eval' AS openmeter.proposal_events_local ENGINE = Distributed('ch_eval', 'openmeter', 'proposal_events_local', cityHash64(namespace));

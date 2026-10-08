select max(kafka_offset) as offset, count() as events
from usage_events
where tenant_id = toUUID('00000000-0000-4000-8000-000000000001')
  and kafka_topic = 'konnect.ai-cost-management.usage-events'
  and kafka_partition = 3
  and ingested_at >= toDateTime64('2026-10-07 11:55:00.000', 3, 'UTC')
  and ingested_at <= toDateTime64('2026-10-07 12:00:00.000', 3, 'UTC')

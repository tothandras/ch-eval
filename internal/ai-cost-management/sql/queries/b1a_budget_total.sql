select sum(toDecimal256(reported_cost_usd, 9)) as total,
  max(ingested_at) as latest_record_at, count() as events
from usage_events final
where tenant_id = toUUID('00000000-0000-4000-8000-000000000001') and time >= toDateTime64('2026-10-01 00:00:00.000', 3, 'UTC') and time < toDateTime64('2026-11-01 00:00:00.000', 3, 'UTC') and has(teams, 'principal/0198c5b6-a000-7000-8000-000000000010')

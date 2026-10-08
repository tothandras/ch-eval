select *
from (
  select sum(b_reported_cost_usd) as t_reported_cost_usd, sum(b_events) as t_events, arraySort(x -> x.1, groupArray(tuple(bucket, b_reported_cost_usd, b_events))) as series
  from (
    select toDateTime64(toTimeZone(toStartOfDay(time, 'UTC'), 'UTC'), 3, 'UTC') as bucket, sum(toDecimal256(reported_cost_usd, 9)) as b_reported_cost_usd, count() as b_events
    from usage_events final
    where tenant_id = toUUID('00000000-0000-4000-8000-000000000001') and time >= toDateTime64('2026-09-01 00:00:00.000', 3, 'UTC') and time < toDateTime64('2026-10-07 00:00:00.000', 3, 'UTC')
    group by bucket
  )
  )
order by t_reported_cost_usd desc
limit 21

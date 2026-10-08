select *, empty(g0_0) as n0
from (
  select g0_0, sum(b_events) as t_events, sum(b_reported_cost_usd) as t_reported_cost_usd, arraySort(x -> x.1, groupArray(tuple(bucket, b_events, b_reported_cost_usd))) as series
  from (
    select toDateTime64(toTimeZone(toStartOfDay(time, 'Europe/Budapest'), 'UTC'), 3, 'UTC') as bucket, subject as g0_0, count() as b_events, sum(toDecimal256(reported_cost_usd, 9)) as b_reported_cost_usd
    from usage_events_t final
    where tenant_id = toUUID('00000000-0000-4000-8000-000000000001') and time >= toDateTime64('2026-10-01 00:00:00.000', 3, 'UTC') and time < toDateTime64('2026-11-01 00:00:00.000', 3, 'UTC')
    group by bucket, g0_0
  )
  group by g0_0)
order by t_reported_cost_usd desc, n0 asc, g0_0 asc
limit 6

select *, empty(g0_0) as n0, empty(g1_0) as n1
from (
  select g0_0, g1_0, sum(b_reported_cost_usd) as t_reported_cost_usd, arraySort(x -> x.1, groupArray(tuple(bucket, b_reported_cost_usd))) as series
  from (
    select toDateTime64(toTimeZone(toDateTime(toStartOfWeek(time, 1, 'UTC'), 'UTC'), 'UTC'), 3, 'UTC') as bucket, arrayJoin(if(empty(teams), [''], teams)) as g0_0, tags['cost_center'] as g1_0, sum(toDecimal256(reported_cost_usd, 9)) as b_reported_cost_usd
    from usage_events final
    where tenant_id = toUUID('00000000-0000-4000-8000-000000000001') and time >= toDateTime64('2026-07-01 00:00:00.000', 3, 'UTC') and time < toDateTime64('2026-10-01 00:00:00.000', 3, 'UTC')
    group by bucket, g0_0, g1_0
  )
  group by g0_0, g1_0)
order by n0 asc, g0_0 asc, n1 asc, g1_0 asc
limit 21

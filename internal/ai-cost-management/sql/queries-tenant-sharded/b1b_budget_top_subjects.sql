select value, total from (
  select value, sum(total) as total from (
    select toString(subject) as value,
      sum(toDecimal256(reported_cost_usd, 9)) as total
    from usage_events_t final
    where tenant_id = toUUID('00000000-0000-4000-8000-000000000001') and time >= toDateTime64('2026-10-01 00:00:00.000', 3, 'UTC') and time < toDateTime64('2026-11-01 00:00:00.000', 3, 'UTC') and has(teams, 'principal/0198c5b6-a000-7000-8000-000000000010')
    group by value
    having notEmpty(value)
    union all
    select arrayJoin(CAST([], 'Array(String)')) as value, toDecimal256(0, 9) as total
  ) group by value
)
order by total desc, value asc
limit 11

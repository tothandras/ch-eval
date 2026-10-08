select toString(subject) as value,
  sum(toDecimal256(reported_cost_usd, 9)) as total
from usage_events final
where tenant_id = toUUID('00000000-0000-4000-8000-000000000001') and time >= toDateTime64('2026-10-01 00:00:00.000', 3, 'UTC') and time < toDateTime64('2026-11-01 00:00:00.000', 3, 'UTC') and provider = 'openai'
group by value
having notEmpty(value) and total >= toDecimal256('800', 9)
order by value

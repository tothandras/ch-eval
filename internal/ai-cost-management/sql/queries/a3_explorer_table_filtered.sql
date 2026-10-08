select *, empty(g0_0) as n0
from (
  select subject as g0_0, count() as t_events, countIf(reported_cost_usd = 0 and notEmpty(arrayFilter(quantity -> quantity > 0, [input_tokens, input_text_tokens, input_audio_tokens, input_image_tokens, input_cache_read_tokens, input_cache_read_text_tokens, input_cache_read_audio_tokens, input_cache_read_image_tokens, input_cache_read_video_tokens, input_cache_write_tokens, input_cache_write_text_tokens, input_cache_write_audio_tokens, input_cache_write_image_tokens, input_cache_write_video_tokens, output_tokens, output_text_tokens, output_audio_tokens, output_reasoning_tokens, accepted_prediction_tokens, rejected_prediction_tokens]))) as t_unpriced_events, sum(toDecimal256(reported_cost_usd, 9)) as t_reported_cost_usd, sum(toUInt128(input_tokens)) as t_input_tokens, sum(toUInt128(output_tokens)) as t_output_tokens, [] as series
  from usage_events final
  where tenant_id = toUUID('00000000-0000-4000-8000-000000000001') and time >= toDateTime64('2026-10-01 00:00:00.000', 3, 'UTC') and time < toDateTime64('2026-11-01 00:00:00.000', 3, 'UTC') and provider = 'anthropic' and (model ilike '%opus%' or model not ilike '%mini%') and hasAny(teams, ['oidc/team-koko']) and hasAny(tag_items, ['ZW52:cHJvZA==', 'ZW52:c3RhZ2luZw=='])
  group by g0_0)
order by t_reported_cost_usd desc, n0 asc, g0_0 asc
limit 21

-- Deterministic synthetic AICM usage events (no real data). Every value is a pure function of
-- `number`, so any target receives byte-identical rows.
--   {N}      rows to generate starting at {OFFSET}
--   {FILTER} extra WHERE on number (e.g. 1 for all rows; number % 20 = 0 for redelivery)
--   {TSHIFT_MS} milliseconds added to `time` (0 normally; 1000 simulates a producer retry with a changed time)
-- Tenants: 20; tenant ...0001 holds ~40% of events (skew). Time: 2026-07-01 .. 2026-10-08 UTC.
INSERT INTO usage_events (tenant_id, id, type, source, subject, time, traceparent, tracestate, ingested_at, stored_at,
  kafka_topic, kafka_partition, kafka_offset, provider, model, model_id, request_mode, region, client,
  input_tokens, input_text_tokens, input_cache_read_tokens, output_tokens, output_reasoning_tokens, reported_total_tokens,
  reported_cost_usd, reported_cache_cost_usd, teams, workspace_id, service_id, route_id, plugin_id, consumer_id,
  consumer_group_ids, application_id, request_id, tags)
SELECT
  toUUID(concat('00000000-0000-4000-8000-', leftPad(toString(t + 1), 12, '0'))) AS tenant_id,
  lower(hex(cityHash64(number, 'id'))) AS id,
  'ai_gateway.usage' AS type,
  if(h % 10 = 0, 'http-ingest', 'konnect/gateway') AS source,
  multiIf(sk % 3 = 0, concat('principal/', toString(u(subj, 'p'))),
          sk % 3 = 1, concat('consumer/', toString(u(subj, 'c'))),
          concat('oidc/', ['alice', 'bob', 'carol', 'dave', 'erin', 'frank'][sk % 6 + 1], '-', toString(sk))) AS subject,
  ts AS time,
  '' AS traceparent, '' AS tracestate,
  ing AS ingested_at, ing AS stored_at,
  'konnect.ai-cost-management.usage-events' AS kafka_topic,
  h % 12 AS kafka_partition,
  number AS kafka_offset,
  prov AS provider,
  mdl AS model, mdl AS model_id,
  ['chat', 'completions', 'embeddings', 'realtime'][h % 4 + 1] AS request_mode,
  ['us-east-1', 'us-west-2', 'eu-west-1'][h % 3 + 1] AS region,
  ['kong-gateway', 'sdk-python', 'sdk-node'][h % 3 + 1] AS client,
  itok AS input_tokens, itok AS input_text_tokens, intDiv(itok, 4) AS input_cache_read_tokens,
  otok AS output_tokens, intDiv(otok, 3) AS output_reasoning_tokens, itok + otok AS reported_total_tokens,
  if(h % 33 = 0, toDecimal64(0, 9), toDecimal64(itok * 0.000003 + otok * 0.000015, 9)) AS reported_cost_usd,
  toDecimal64(intDiv(itok, 4) * 0.0000003, 9) AS reported_cache_cost_usd,
  arrayDistinct(arrayMap(i -> team_pool[(cityHash64(number, 'team', i) % 50) + 1], range(h % 4))) AS teams,
  u(t, 'ws') AS workspace_id,
  u(h % 40, 'svc') AS service_id, u(h % 200, 'route') AS route_id, u(h % 5, 'plugin') AS plugin_id,
  if(sk % 3 = 1, u(subj, 'c'), toUUID('00000000-0000-0000-0000-000000000000')) AS consumer_id,
  arrayMap(i -> u(cityHash64(number, 'cg', i) % 30, 'cg'), range(h % 3)) AS consumer_group_ids,
  u(h % 60, 'app') AS application_id,
  lower(hex(cityHash64(number, 'req'))) AS request_id,
  mapFromArrays(tag_keys, arrayMap(k -> multiIf(k = 'env', ['prod', 'staging', 'dev'][h % 3 + 1],
                                                 k = 'cost_center', concat('cc-', toString(cityHash64(number, 'cc') % 30 + 1)),
                                                 concat('v', toString(cityHash64(number, k) % 200 + 1))), tag_keys)) AS tags
FROM (
  SELECT number,
    cityHash64(number) AS h,
    if(h % 100 < 40, 0, 1 + (cityHash64(number, 't') % 19)) AS t,
    cityHash64(number, 's') % 3000 AS sk,
    cityHash64(t, sk) AS subj,
    toDateTime64('2026-07-01 00:00:00.000', 3, 'UTC') + toIntervalMillisecond(cityHash64(number, 'ts') % (99 * 86400000)) + toIntervalMillisecond({TSHIFT_MS}) AS ts,
    ts + toIntervalMillisecond(if(h % 20 = 7, 7200000 + cityHash64(number, 'late') % 172800000, 1000 + cityHash64(number, 'lag') % 29000)) AS ing,
    ['openai', 'anthropic', 'google', 'mistral'][h % 4 + 1] AS prov,
    multiIf(prov = 'openai', ['gpt-5', 'gpt-5-mini', 'o4'][cityHash64(number, 'm') % 3 + 1],
            prov = 'anthropic', ['claude-opus-4', 'claude-sonnet-4', 'claude-haiku-4'][cityHash64(number, 'm') % 3 + 1],
            prov = 'google', ['gemini-2.5-pro', 'gemini-2.5-flash'][cityHash64(number, 'm') % 2 + 1],
            ['mistral-large', 'mistral-small'][cityHash64(number, 'm') % 2 + 1]) AS mdl,
    cityHash64(number, 'in') % 5000 AS itok,
    cityHash64(number, 'out') % 2000 AS otok,
    arrayMap(i -> if(i = 0, 'principal/0198c5b6-a000-7000-8000-000000000010', if(i = 1, 'oidc/team-koko', concat('principal/0198c5b6-a000-7000-8000-', leftPad(toString(i + 10), 12, '0')))), range(50)) AS team_pool,
    arrayDistinct(arrayConcat(if(h % 10 < 7, ['env', 'cost_center'], []), arrayMap(i -> concat('k', toString(cityHash64(number, 'tk', i) % 40 + 1)), range(cityHash64(number, 'nk') % 6)))) AS tag_keys
  FROM numbers({OFFSET}, {N})
  WHERE {FILTER}
)

-- Verbatim from the brief (section 8.2).
CREATE OR REPLACE FUNCTION usage_event_field_value_observations AS
(source, provider, model, model_id, region, client, subject,
 service_id, route_id, consumer_id, application_id, teams, consumer_group_ids, tags) ->
arrayConcat(
  arrayFilter(item -> notEmpty(item.3), [
    ('source', '', toString(source)),
    ('provider', '', toString(provider)),
    ('model', '', toString(model)),
    ('model_id', '', toString(model_id)),
    ('region', '', toString(region)),
    ('client', '', toString(client)),
    ('subject', '', subject)
  ]),
  arrayFilter(item -> item.3 != '00000000-0000-0000-0000-000000000000', [
    ('service_id', '', toString(service_id)),
    ('route_id', '', toString(route_id)),
    ('consumer_id', '', toString(consumer_id)),
    ('application_id', '', toString(application_id))
  ]),
  arrayMap(id -> ('teams', '', toString(id)), teams),
  arrayMap(id -> ('consumer_group_ids', '', toString(id)), consumer_group_ids),
  arrayMap(k -> ('tag_keys', '', toString(k)), mapKeys(tags)),
  arrayMap((k, v) -> ('tags', toString(k), v), mapKeys(tags), mapValues(tags))
);
CREATE FUNCTION IF NOT EXISTS tag_token AS (key, value) ->
    concat(base64Encode(key), ':', base64Encode(value));

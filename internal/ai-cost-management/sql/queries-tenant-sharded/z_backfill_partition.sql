INSERT INTO usage_event_field_values (tenant_id, field, key, value)
SELECT tenant_id, observation.1 AS field, observation.2 AS key, observation.3 AS value
FROM usage_events
ARRAY JOIN usage_event_field_value_observations(
  source, provider, model, model_id, region, client, subject,
  service_id, route_id, consumer_id, application_id, teams, consumer_group_ids, tags
) AS observation
WHERE _partition_id = '202610'
GROUP BY tenant_id, field, key, value

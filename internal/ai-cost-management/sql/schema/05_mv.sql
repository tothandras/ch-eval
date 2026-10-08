-- Verbatim from the brief (section 8.2).
CREATE MATERIALIZED VIEW IF NOT EXISTS usage_event_field_values_mv
TO usage_event_field_values
DEFINER = ai_cost_management_field_values_definer SQL SECURITY DEFINER
AS SELECT DISTINCT
  tenant_id,
  observation.1 AS field,
  observation.2 AS key,
  observation.3 AS value
FROM usage_events
ARRAY JOIN usage_event_field_value_observations(
  source, provider, model, model_id, region, client, subject,
  service_id, route_id, consumer_id, application_id, teams, consumer_group_ids, tags
) AS observation;

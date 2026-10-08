-- Cluster-only migrations (D-08), run in ai_cost_management. Distributed tables named like the
-- local ones; D-09 sharding key on usage_events. Grants on the Distributed tables are needed
-- because the reader/writer hit these tables first.
CREATE TABLE IF NOT EXISTS ai_cost_management.usage_events AS ai_cost_management_local.usage_events ENGINE = Distributed(prod_us, ai_cost_management_local, usage_events, cityHash64(tenant_id, id));
CREATE TABLE IF NOT EXISTS ai_cost_management.usage_event_field_values AS ai_cost_management_local.usage_event_field_values ENGINE = Distributed(prod_us, ai_cost_management_local, usage_event_field_values, rand());
GRANT SELECT ON ai_cost_management.usage_events TO ai_cost_management_reader;
GRANT SELECT ON ai_cost_management.usage_event_field_values TO ai_cost_management_reader;
GRANT INSERT ON ai_cost_management.usage_events TO ai_cost_management_writer;

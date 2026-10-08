-- Data-team provisioning (D-07, D-14), run once by an admin. ON CLUSTER is used here only
-- because database creation is a provisioning step, not an application migration.
CREATE DATABASE IF NOT EXISTS ai_cost_management_local ON CLUSTER prod_us ENGINE = Replicated('/clickhouse/databases/ai_cost_management_local', '{shard}', '{replica}');
CREATE DATABASE IF NOT EXISTS ai_cost_management ON CLUSTER prod_us ENGINE = Replicated('/clickhouse/databases/ai_cost_management', 'all', '{replica}');

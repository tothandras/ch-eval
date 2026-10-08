-- Verbatim from the brief (section 8.3).
CREATE ROLE IF NOT EXISTS ai_cost_management_reader;
CREATE ROLE IF NOT EXISTS ai_cost_management_writer;
GRANT SELECT ON usage_events TO ai_cost_management_reader;
GRANT INSERT ON usage_events TO ai_cost_management_writer;
CREATE ROW POLICY OR REPLACE usage_events_reader_base ON usage_events
FOR SELECT USING 1
AS PERMISSIVE
TO ai_cost_management_reader;
CREATE ROW POLICY OR REPLACE usage_events_tenant_guard ON usage_events
FOR SELECT USING tenant_id = toUUID(getSetting('SQL_tenant_id'))
AS RESTRICTIVE
TO ai_cost_management_reader;
CREATE ROW POLICY IF NOT EXISTS usage_event_field_values_reader_base
ON usage_event_field_values FOR SELECT USING 1
AS PERMISSIVE TO ai_cost_management_reader;
CREATE ROW POLICY IF NOT EXISTS usage_event_field_values_tenant_guard
ON usage_event_field_values
FOR SELECT USING tenant_id = toUUID(getSetting('SQL_tenant_id'))
AS RESTRICTIVE TO ai_cost_management_reader;
GRANT SELECT ON usage_event_field_values TO ai_cost_management_reader;
CREATE USER IF NOT EXISTS ai_cost_management_field_values_definer IDENTIFIED WITH sha256_hash BY '0000000000000000000000000000000000000000000000000000000000000000' HOST NONE;
GRANT SELECT ON usage_events TO ai_cost_management_field_values_definer;
GRANT INSERT ON usage_event_field_values TO ai_cost_management_field_values_definer;
CREATE SETTINGS PROFILE OR REPLACE ai_cost_management_reader_settings
SETTINGS
  readonly = 1 CONST,
  SQL_tenant_id = '' CHANGEABLE_IN_READONLY,
  serialize_query_plan = 0 CONST,
  date_time_input_format = 'best_effort' CONST,
  date_time_output_format = 'iso' CONST,
  output_format_json_quote_64bit_integers = 1 CONST,
  output_format_json_quote_decimals = 1 CONST,
  max_execution_time MAX 300 CHANGEABLE_IN_READONLY,
  max_memory_usage MAX 536870912 CHANGEABLE_IN_READONLY,
  max_threads MAX 2 CHANGEABLE_IN_READONLY,
  cancel_http_readonly_queries_on_client_close CHANGEABLE_IN_READONLY,
  enable_http_compression = 1 CONST,
  output_format_decimal_trailing_zeros = 0 CONST,
  output_format_json_escape_forward_slashes = 0 CONST,
  format_csv_null_representation = '' CONST,
  output_format_parquet_compression_method = 'zstd' CONST,
  output_format_parquet_string_as_string = 1 CONST,
  output_format_parquet_compliant_nested_types = 1 CONST,
  output_format_parquet_row_group_size_bytes = 67108864 CONST
TO ai_cost_management_reader;

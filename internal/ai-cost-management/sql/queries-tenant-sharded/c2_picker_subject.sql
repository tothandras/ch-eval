select distinct tenant_id, field, key, value as observed_value
      from usage_event_field_values
      where tenant_id = toUUID('00000000-0000-4000-8000-000000000001') and field = 'subject' and key = '' and value ilike '%alice%'
      order by tenant_id asc, field asc, key asc, value asc
      limit 21

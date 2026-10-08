-- probe: Roles and grants (SE-02)
-- cleanup: DROP ROLE IF EXISTS ch_eval_probe_role
-- expect: value:1
CREATE ROLE ch_eval_probe_role;
GRANT SELECT ON {db}.* TO ch_eval_probe_role;
SELECT count() FROM system.grants WHERE role_name = 'ch_eval_probe_role'

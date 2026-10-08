-- probe: GRANT ALL (SE-02; restricted on Cloud)
-- cleanup: DROP ROLE IF EXISTS ch_eval_probe_role_all
-- expect: ok
-- expect.cloud: error
CREATE ROLE ch_eval_probe_role_all;
GRANT ALL ON {db}.* TO ch_eval_probe_role_all

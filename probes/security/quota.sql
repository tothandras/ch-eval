-- probe: Quotas (SE-03)
-- cleanup: DROP QUOTA IF EXISTS ch_eval_probe_quota
-- expect: value:1
CREATE QUOTA ch_eval_probe_quota FOR INTERVAL 1 hour MAX queries = 1000000;
SELECT count() FROM system.quotas WHERE name = 'ch_eval_probe_quota'

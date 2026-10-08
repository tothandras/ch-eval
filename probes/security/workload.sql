-- probe: Workload scheduling DDL (SE-04)
-- cleanup: DROP WORKLOAD IF EXISTS ch_eval_probe_leaf
-- cleanup: DROP WORKLOAD IF EXISTS ch_eval_probe_root
-- expect: value:2
CREATE WORKLOAD ch_eval_probe_root;
CREATE WORKLOAD ch_eval_probe_leaf IN ch_eval_probe_root SETTINGS max_concurrent_threads = 2;
SELECT count() FROM system.workloads WHERE name LIKE 'ch_eval_probe_%'

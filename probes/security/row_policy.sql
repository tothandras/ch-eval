-- probe: Row policy restricts rows by namespace (SE-01)
-- cleanup: DROP ROW POLICY IF EXISTS ch_eval_probe_rp ON {db}.rp_probe
-- expect: value:1
CREATE TABLE {db}.rp_probe (namespace String) ENGINE = MergeTree ORDER BY namespace;
INSERT INTO {db}.rp_probe VALUES ('a'), ('b');
CREATE ROW POLICY ch_eval_probe_rp ON {db}.rp_probe USING namespace = 'a' TO CURRENT_USER;
SELECT count() FROM {db}.rp_probe

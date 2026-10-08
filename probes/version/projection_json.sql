-- probe: Projection over a JSON subcolumn (VG-09; undocumented, the observation is the evidence)
-- expect: value:1 || error
CREATE TABLE {db}.pj_probe (k UInt8, d JSON, PROJECTION p (SELECT d.a.b, count() GROUP BY d.a.b)) ENGINE = MergeTree ORDER BY k;
INSERT INTO {db}.pj_probe VALUES (1, '{"a": {"b": 7}}');
SELECT count() FROM {db}.pj_probe WHERE d.a.b = 7

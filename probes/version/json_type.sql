-- probe: JSON type with typed subcolumn reads (VG-01)
-- expect: value:42
CREATE TABLE {db}.json_probe (d JSON) ENGINE = MergeTree ORDER BY tuple();
INSERT INTO {db}.json_probe VALUES ('{"a": {"b": 42}}');
SELECT d.a.b FROM {db}.json_probe

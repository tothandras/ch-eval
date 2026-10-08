-- probe: Lightweight DELETE on a table with a projection (VG-10); default mode throws
-- expect: error
CREATE TABLE {db}.plw_probe (k UInt8, v UInt8, PROJECTION p (SELECT k, v ORDER BY v)) ENGINE = MergeTree ORDER BY k;
INSERT INTO {db}.plw_probe VALUES (1, 1), (2, 2);
DELETE FROM {db}.plw_probe WHERE k = 2;
SELECT count() FROM {db}.plw_probe

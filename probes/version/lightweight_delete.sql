-- probe: Lightweight DELETE (VG-04)
-- expect: value:1
CREATE TABLE {db}.lwd_probe (k UInt8) ENGINE = MergeTree ORDER BY k;
INSERT INTO {db}.lwd_probe VALUES (1), (2);
DELETE FROM {db}.lwd_probe WHERE k = 2;
SELECT count() FROM {db}.lwd_probe

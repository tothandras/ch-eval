-- probe: Lightweight UPDATE with enable_lightweight_update = 1 (VG-05)
-- expect: value:2
CREATE TABLE {db}.lwu2_probe (k UInt8, v UInt8) ENGINE = MergeTree ORDER BY k SETTINGS enable_block_number_column = 1, enable_block_offset_column = 1;
INSERT INTO {db}.lwu2_probe VALUES (1, 1);
UPDATE {db}.lwu2_probe SET v = 2 WHERE k = 1 SETTINGS enable_lightweight_update = 1;
SELECT v FROM {db}.lwu2_probe

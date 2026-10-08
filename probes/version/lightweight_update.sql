-- probe: Lightweight UPDATE ... SET without extra settings (VG-05, beta in 26.x)
-- expect: value:2 || error
CREATE TABLE {db}.lwu_probe (k UInt8, v UInt8) ENGINE = MergeTree ORDER BY k SETTINGS enable_block_number_column = 1, enable_block_offset_column = 1;
INSERT INTO {db}.lwu_probe VALUES (1, 1);
UPDATE {db}.lwu_probe SET v = 2 WHERE k = 1;
SELECT v FROM {db}.lwu_probe

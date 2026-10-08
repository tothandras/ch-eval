-- probe: _part_offset projection (VG-10)
-- expect: value:1
CREATE TABLE {db}.po_probe (k UInt8, v String, PROJECTION by_v (SELECT _part_offset ORDER BY v)) ENGINE = MergeTree ORDER BY k;
INSERT INTO {db}.po_probe VALUES (1, 'a'), (2, 'b');
SELECT count() FROM {db}.po_probe WHERE v = 'b'

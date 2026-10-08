-- probe: Settings select_sequential_consistency depends on (RS-04)
-- expect: value:0	0	true
-- expect.cloud: ok
SELECT getSetting('select_sequential_consistency'), getSetting('insert_quorum'), getSetting('insert_quorum_parallel')

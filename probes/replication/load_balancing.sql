-- probe: Replica selection defaults (RS-03); the operator's default profile overrides them
-- expect: value:random	true
-- expect.tier-kind: value:nearest_hostname	false
-- expect.cloud: ok
SELECT getSetting('load_balancing'), getSetting('prefer_localhost_replica')

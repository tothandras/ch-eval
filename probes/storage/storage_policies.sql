-- probe: Storage policies defined on the target (CS-01, ST-04); the s3 tier must expose s3_cached
-- expect: ok
-- expect.storage-s3: regex:s3_cached
SELECT arrayStringConcat(groupArray(policy_name), ',') FROM (SELECT DISTINCT policy_name FROM system.storage_policies ORDER BY policy_name)

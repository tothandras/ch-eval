-- probe: Zero-copy replication default (ST-05)
-- expect: value:0
-- expect.cloud: ok
SELECT value FROM system.merge_tree_settings WHERE name = 'allow_remote_fs_zero_copy_replication'

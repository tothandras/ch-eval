-- probe: Default database engine (ST-02)
-- expect: value:Atomic
-- expect.cloud: value:Shared
SELECT engine FROM system.databases WHERE name = '{db}'

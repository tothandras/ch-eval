-- probe: async_insert, wait_for_async_insert and flush-timeout defaults (IN-01, IN-02)
-- expect: value:true	true	200
-- expect.cloud: value:true	true	1000
SELECT getSetting('async_insert'), getSetting('wait_for_async_insert'), getSetting('async_insert_busy_timeout_max_ms')

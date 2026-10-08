-- user_logs through ClickStack's LogAttributeItems ALIAS column (text index). Same rows.
SELECT ServiceName, count() AS logs
FROM otel_logs
WHERE has(LogAttributeItems, concat('user.id=', {user_id:String}))
GROUP BY ServiceName
ORDER BY ServiceName

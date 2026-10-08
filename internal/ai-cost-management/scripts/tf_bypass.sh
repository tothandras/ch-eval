#!/usr/bin/env bash
# Can the reader read other tenants through mergeTreeIndex / mergeTreeProjection? (row-policy fixes #112329, #115304)
set -uo pipefail
HERE="$(cd "$(dirname "$0")/.." && pwd)"; img="$1"; name="tf-$(echo $img | tr '/:.' '___')"
docker rm -f $name >/dev/null 2>&1
docker run -d --name $name -e CLICKHOUSE_SKIP_USER_SETUP=1 -e CLICKHOUSE_DEFAULT_ACCESS_MANAGEMENT=1 -v $HERE/cluster/config/users.xml:/etc/clickhouse-server/users.d/zz-default.xml:ro $img >/dev/null
docker cp $HERE/cluster/single-prefix.xml $name:/etc/clickhouse-server/config.d/prefix.xml >/dev/null; docker restart $name >/dev/null
ip=$(docker inspect -f '{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}' $name)
for i in $(seq 1 60); do $HERE/scripts/chhttp.py $ip "SELECT 1" >/dev/null 2>&1 && break; sleep 1; done
Q="$HERE/scripts/chhttp.py $ip"
$Q "CREATE DATABASE IF NOT EXISTS ai_cost_management" 2>/dev/null
for f in $HERE/sql/schema/0*.sql $HERE/sql/gen/u_function.sql; do $HERE/scripts/run_sql.py $ip ai_cost_management $f --continue >/dev/null; done
$Q "CREATE USER IF NOT EXISTS aicm_reader IDENTIFIED WITH plaintext_password BY 'reader' DEFAULT DATABASE ai_cost_management" 2>/dev/null; $Q "GRANT ai_cost_management_reader TO aicm_reader" 2>/dev/null; $Q "ALTER USER aicm_reader DEFAULT ROLE ai_cost_management_reader" 2>/dev/null
$HERE/scripts/gen.py $ip ai_cost_management 200000 >/dev/null
R="--db ai_cost_management --user aicm_reader --password reader --set SQL_tenant_id=00000000-0000-4000-8000-000000000001 --fmt TSV"
echo "== $img"
echo "  usage_events (policy)                : $($Q "SELECT uniqExact(tenant_id), count() FROM usage_events" $R 2>&1 | head -c 160)"
echo "  mergeTreeIndex(...)                  : $($Q "SELECT uniqExact(tenant_id), count() FROM mergeTreeIndex(ai_cost_management, usage_events, with_marks = 0)" $R 2>&1 | head -c 160)"
echo "  mergeTreeProjection(..., by_subject) : $($Q "SELECT count() FROM mergeTreeProjection(ai_cost_management, usage_events, by_subject)" $R 2>&1 | head -c 160)"
echo "  total rows (admin)                   : $($Q "SELECT count() FROM ai_cost_management.usage_events" --fmt TSV 2>/dev/null)"
docker rm -f $name >/dev/null 2>&1

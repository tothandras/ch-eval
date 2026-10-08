#!/usr/bin/env bash
# Version-sensitive behaviours on a single node with 1M synthetic rows (+ redeliveries):
# FINAL partition-local plan, projection use, C1 read-in-order, map bucket threshold.
set -uo pipefail
HERE="$(cd "$(dirname "$0")/.." && pwd)"; img="$1"; name="vb-$(echo $img | tr '/:.' '___')"
docker rm -f $name >/dev/null 2>&1
docker run -d --name $name -e CLICKHOUSE_SKIP_USER_SETUP=1 -e CLICKHOUSE_DEFAULT_ACCESS_MANAGEMENT=1 -v $HERE/cluster/config/users.xml:/etc/clickhouse-server/users.d/zz-default.xml:ro $img >/dev/null
docker cp $HERE/cluster/single-prefix.xml $name:/etc/clickhouse-server/config.d/prefix.xml >/dev/null; docker restart $name >/dev/null
ip=$(docker inspect -f '{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}' $name)
for i in $(seq 1 60); do $HERE/scripts/chhttp.py $ip "SELECT 1" >/dev/null 2>&1 && break; sleep 1; done
Q="$HERE/scripts/chhttp.py $ip"
$Q "CREATE DATABASE IF NOT EXISTS ai_cost_management" 2>/dev/null
for f in $HERE/sql/schema/0*.sql $HERE/sql/gen/u_function.sql; do $HERE/scripts/run_sql.py $ip ai_cost_management $f --continue >/dev/null; done
$HERE/scripts/gen.py $ip ai_cost_management 1000000 >/dev/null && $HERE/scripts/gen.py $ip ai_cost_management 1000000 --filter 'number % 20 = 0' >/dev/null
echo "== $img ($($Q 'SELECT version()' 2>/dev/null))"
FQ="SELECT sum(toDecimal256(reported_cost_usd, 9)) FROM usage_events FINAL WHERE tenant_id = toUUID('00000000-0000-4000-8000-000000000001') AND time >= toDateTime64('2026-07-01 00:00:00.000', 3, 'UTC') AND time < toDateTime64('2026-10-01 00:00:00.000', 3, 'UTC')"
echo -n "FINAL ReplacingSorted streams (default settings): "; $Q "EXPLAIN PIPELINE $FQ" --db ai_cost_management --set max_threads=2 --fmt TSVRaw 2>/dev/null | grep -c 'ReplacingSorted'
echo -n "FINAL ReplacingSorted streams (do_not_merge_across_partitions_select_final=1): "; $Q "EXPLAIN PIPELINE $FQ" --db ai_cost_management --set max_threads=2 --set do_not_merge_across_partitions_select_final=1 --fmt TSVRaw 2>/dev/null | grep -c 'ReplacingSorted'
SUBJ=$($Q "SELECT subject FROM usage_events WHERE tenant_id=toUUID('00000000-0000-4000-8000-000000000001') LIMIT 1" --db ai_cost_management --fmt TSV 2>/dev/null)
for f in "" FINAL; do id=vb-$RANDOM$RANDOM; $Q "SELECT count() FROM usage_events $f WHERE tenant_id = toUUID('00000000-0000-4000-8000-000000000001') AND subject = '$SUBJ'" --db ai_cost_management --set query_id=$id --fmt Null >/dev/null 2>&1; $Q "SYSTEM FLUSH LOGS" >/dev/null 2>&1; echo "subject-eq [$f]: $($Q "SELECT read_rows, projections FROM system.query_log WHERE query_id='$id' AND type='QueryFinish'" --fmt TSV 2>/dev/null)"; done
for f in $HERE/sql/queries/a2_*.sql $HERE/sql/queries/b3_*.sql; do id=vb-$RANDOM$RANDOM; $Q @$f --db ai_cost_management --set query_id=$id --fmt Null >/dev/null 2>&1; $Q "SYSTEM FLUSH LOGS" >/dev/null 2>&1; echo "$(basename $f .sql): $($Q "SELECT read_rows, projections FROM system.query_log WHERE query_id='$id' AND type='QueryFinish'" --fmt TSV 2>/dev/null)"; done
id=vb-$RANDOM$RANDOM; $Q "select id from usage_events final where tenant_id = toUUID('00000000-0000-4000-8000-000000000001') order by time desc, source asc, id asc limit 21" --db ai_cost_management --set query_id=$id --fmt Null >/dev/null 2>&1; $Q "SYSTEM FLUSH LOGS" >/dev/null 2>&1; echo "C1 unbounded FINAL read_rows: $($Q "SELECT read_rows FROM system.query_log WHERE query_id='$id' AND type='QueryFinish'" --fmt TSV 2>/dev/null) of tenant rows $($Q "SELECT count() FROM usage_events WHERE tenant_id = toUUID('00000000-0000-4000-8000-000000000001')" --db ai_cost_management --fmt TSV 2>/dev/null)"
echo "map: $($Q "SELECT name||'='||value FROM system.merge_tree_settings WHERE name IN ('map_buckets_min_avg_size','map_serialization_version_for_zero_level_parts')" --fmt TSV 2>/dev/null | tr '\n' ' ')"
docker rm -f $name >/dev/null 2>&1

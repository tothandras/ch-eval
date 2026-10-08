#!/usr/bin/env bash
# Q1: apply the verbatim AICM schema to a single-node server of each version and
# record which statements succeed. Usage: q1_features.sh <image> [<image>...]
set -uo pipefail
HERE="$(cd "$(dirname "$0")/.." && pwd)"
OUT="$HERE/results/q1-features"
mkdir -p "$OUT"
CH=/home/user/tools/ch/clickhouse
for img in "$@"; do
  name="q1-$(echo "$img" | tr '/:' '__')"
  docker rm -f "$name" >/dev/null 2>&1
  docker run -d --name "$name" -e CLICKHOUSE_SKIP_USER_SETUP=1 -e CLICKHOUSE_DEFAULT_ACCESS_MANAGEMENT=1 "$img" >/dev/null
  ip=$(docker inspect -f '{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}' "$name")
  for i in $(seq 1 60); do $CH client --host "$ip" -q "SELECT 1" >/dev/null 2>&1 && break; sleep 1; done
  log="$OUT/$(echo "$img" | tr '/:' '__').log"
  {
    echo "# image: $img"
    echo "# version: $($CH client --host "$ip" -q 'SELECT version()')"
    echo "# custom_settings_prefixes: $($CH client --host "$ip" -q "SELECT value FROM system.server_settings WHERE name='custom_settings_prefixes'")"
    for f in "$HERE"/sql/schema/0*.sql; do
      # split on ';' at end of line; statements in these files never contain ';' mid-line
      python3 - "$f" <<'PY' > /tmp/q1_stmts.txt
import sys,re
s=open(sys.argv[1]).read()
s="\n".join(l for l in s.splitlines() if not l.strip().startswith('--'))
parts=[p.strip() for p in re.split(r';\s*\n', s+"\n") if p.strip()]
print("\x1e".join(parts))
PY
      IFS=$'\x1e' read -r -d '' -a stmts < /tmp/q1_stmts.txt
      for st in "${stmts[@]}"; do
        st="$(echo "$st" | sed 's/;\s*$//')"
        [ -z "${st// }" ] && continue
        first="$(echo "$st" | head -1 | cut -c1-90)"
        if err=$($CH client --host "$ip" -q "$st" 2>&1); then echo "OK    $(basename "$f") :: $first"; else echo "FAIL  $(basename "$f") :: $first"; echo "      $(echo "$err" | head -2 | tr '\n' ' ' | cut -c1-400)"; fi
      done
    done
    echo "## SHOW CREATE TABLE usage_events"
    $CH client --host "$ip" -q "SHOW CREATE TABLE usage_events FORMAT TSVRaw" 2>&1
    echo "## engine/settings"
    $CH client --host "$ip" -q "SELECT name, engine FROM system.tables WHERE database='default' ORDER BY name" 2>&1
    echo "## relevant setting defaults"
    $CH client --host "$ip" -q "SELECT name, value FROM system.settings WHERE name IN ('allow_experimental_full_text_index','enable_full_text_index','use_skip_indexes_on_data_read','do_not_merge_across_partitions_select_final','enable_automatic_decision_for_merging_across_partitions_for_final','enable_vertical_final','optimize_use_projections','optimize_use_projection_filtering','min_table_rows_to_use_projection_index','use_query_condition_cache','async_insert','deduplicate_insert','insert_deduplicate','use_skip_indexes_if_final','use_skip_indexes_if_final_exact_mode','optimize_read_in_order','read_in_order_use_virtual_row','distributed_foreground_insert','skip_unavailable_shards','fallback_to_stale_replicas_for_distributed_queries','load_balancing','prefer_localhost_replica','serialize_query_plan','enable_parallel_replicas','distributed_push_down_limit') ORDER BY name FORMAT TSV" 2>&1
    $CH client --host "$ip" -q "SELECT name, value FROM system.merge_tree_settings WHERE name IN ('map_serialization_version','map_buckets_strategy','max_buckets_in_map','deduplicate_merge_projection_mode','allow_part_offset_column_in_projections','replicated_deduplication_window','non_replicated_deduplication_window','ttl_only_drop_parts','min_bytes_for_wide_part') ORDER BY name FORMAT TSV" 2>&1
  } > "$log" 2>&1
  docker rm -f "$name" >/dev/null 2>&1
  echo "== $img"; grep -cE '^OK' "$log"; grep -A1 -E '^FAIL' "$log"
done

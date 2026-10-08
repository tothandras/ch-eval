#!/usr/bin/env python3
"""D-13 trial: worker-style insert through the Distributed table, then the notifier's D-13 read
(throwIf on the needed offset + evaluation, load_balancing='in_order') from a given initiator.
Usage: d13_trial.py <writer_host> <reader_host> <base_offset> [writer setting k=v ...]"""
import sys, os, subprocess, json
sys.path.insert(0, os.path.dirname(__file__))
from chhttp import query
wh, rh, base = sys.argv[1], sys.argv[2], int(sys.argv[3])
wset = dict(kv.split('=', 1) for kv in sys.argv[4:])
rows = subprocess.run([sys.executable, '/tmp/claude-0/-home-user-ch-eval/da9d56ce-ef19-5887-9882-f7a54916e68c/scratchpad/batch.py', str(base)], capture_output=True, text=True).stdout
s = {'date_time_input_format': 'best_effort', 'distributed_foreground_insert': '1'}; s.update(wset)
ok, body, _ = query(wh, 'INSERT INTO usage_events FORMAT JSONEachRow\n' + rows, 'ai_cost_management', 'aicm_writer', 'writer', s)
if not ok: print('INSERT FAILED', body[:300]); sys.exit(1)
T9 = '00000000-0000-4000-8000-000000000009'
need = base + 299
mode = os.environ.get('CHECK', 'maxoffset')
cond = (f"maxIf(kafka_offset, kafka_partition = 5) < {need}" if mode == 'maxoffset'
        else f"countIf(kafka_partition = 5 AND kafka_offset BETWEEN {base} AND {need}) < 300")
q = f"""SELECT throwIf({cond}, 'needed rows not visible') AS chk,
  countIf(kafka_offset BETWEEN {base} AND {need}) AS batch_rows_visible, sum(toDecimal256(reported_cost_usd, 9)) AS total
FROM usage_events FINAL WHERE tenant_id = toUUID('{T9}') AND time >= toDateTime64('2026-10-01 00:00:00.000', 3, 'UTC') AND time < toDateTime64('2026-11-01 00:00:00.000', 3, 'UTC')"""
rset = {'SQL_tenant_id': T9, 'load_balancing': 'in_order'}; rset.update(dict(kv.split('=',1) for kv in os.environ.get('RSET','').split(',') if kv))
ok, body, _ = query(rh, q, 'ai_cost_management', 'aicm_reader', 'reader', rset, 'TSV')
res = body.strip().split('\t') if ok else None
# where did each shard's rows land? (admin view per replica)
_, where, _ = query(rh, f"SELECT hostName(), count() FROM clusterAllReplicas(prod_us, ai_cost_management_local.usage_events) WHERE kafka_offset BETWEEN {base} AND {need} GROUP BY 1 ORDER BY 1", None, 'default', '', None, 'TSV')
print(json.dumps({'check': mode, 'base': base, 'check_passed': ok, 'visible_of_300': int(res[1]) if res else None,
                  'error': None if ok else body.strip()[:90], 'rows_by_replica': dict(l.split('\t') for l in where.strip().splitlines())}))

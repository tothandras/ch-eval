#!/usr/bin/env python3
"""Run one query file and report HTTP status, exception, and query_log memory/read stats.
Usage: probe_query.py <host> <sqlfile> <label> [--user U --password P] [--fmt F] [--set k=v ...] [--cluster C]"""
import sys, os, argparse, uuid, time
sys.path.insert(0, os.path.dirname(__file__))
from chhttp import query
ap = argparse.ArgumentParser()
ap.add_argument('host'); ap.add_argument('sqlfile'); ap.add_argument('label')
ap.add_argument('--user', default='default'); ap.add_argument('--password', default='')
ap.add_argument('--fmt', default='Null'); ap.add_argument('--set', action='append', default=[])
ap.add_argument('--cluster'); ap.add_argument('--db', default='ai_cost_management')
a = ap.parse_args()
sql = open(a.sqlfile).read().strip().rstrip(';')
qid = f'{a.label}-{uuid.uuid4().hex[:8]}'
s = dict(kv.split('=', 1) for kv in a.set); s['query_id'] = qid
t0 = time.time(); ok, body, hdr = query(a.host, sql, a.db, a.user, a.password, s, a.fmt); wall = time.time() - t0
query(a.host, 'SYSTEM FLUSH LOGS' + (f' ON CLUSTER {a.cluster}' if a.cluster else ''))
src = f'clusterAllReplicas({a.cluster}, system.query_log)' if a.cluster else 'system.query_log'
_, st, _ = query(a.host, f"""SELECT maxIf(memory_usage, is_initial_query), maxIf(memory_usage, NOT is_initial_query), sum(read_rows),
  anyIf(exception_code, is_initial_query AND type != 'QueryFinish'), sumIf(ProfileEvents['ExternalAggregationWritePart'], 1), sumIf(ProfileEvents['ExternalSortWritePart'],1),
  sumIf(ProfileEvents['TemporaryFilesForAggregation'],1) FROM {src} WHERE initial_query_id = '{qid}' AND type != 'QueryStart'""", None, 'default', '', None, 'TSV')
f = st.strip().split('\t')
tail = body.strip().splitlines()[-1][:160] if (not ok or 'Code:' in body[-400:]) else ''
print(f"{a.label:40s} http_ok={ok} wall={wall:.2f}s mem_init={int(f[0])/2**20:.0f}MiB mem_remote={int(f[1])/2**20:.0f}MiB read_rows={f[2]} exc={f[3]} ext_agg_parts={f[4]} ext_sort_parts={f[5]} {('TAIL: '+tail) if tail else ''}")

#!/usr/bin/env python3
"""Run every brief query as the reader on a target and record digest, rows/bytes read,
projections used, and peak memory (initiator + remote sub-queries) from query_log.
Usage: run_queries.py <label> <host> [--cluster prod_us] [--set k=v ...] [--only a1,b3]
Writes results/queries/<label>.tsv"""
import sys, os, argparse, hashlib, time, uuid, json
sys.path.insert(0, os.path.dirname(__file__))
from chhttp import query
ap = argparse.ArgumentParser()
ap.add_argument('label'); ap.add_argument('host'); ap.add_argument('--cluster')
ap.add_argument('--set', action='append', default=[]); ap.add_argument('--only', default='')
ap.add_argument('--db', default='ai_cost_management')
a = ap.parse_args()
QDIR = os.environ.get('QDIR') or os.path.join(os.path.dirname(__file__), '..', 'sql', 'queries')
T1 = '00000000-0000-4000-8000-000000000001'
base = {'SQL_tenant_id': T1, 'max_execution_time': '20'}
base.update(dict(kv.split('=', 1) for kv in a.set))
names = sorted(f[:-4] for f in os.listdir(QDIR) if f.endswith('.sql') and not f.startswith('z_'))
if a.only: names = [n for n in names if n.split('_')[0] in a.only.split(',')]
rows = []
for n in names:
    sql = open(os.path.join(QDIR, n + '.sql')).read().strip().rstrip(';')
    qid = f'{a.label}-{n}-{uuid.uuid4().hex[:8]}'
    s = dict(base); s['query_id'] = qid
    t0 = time.time()
    ok, body, hdr = query(a.host, sql, a.db, 'aicm_reader', 'reader', s, 'TSV')
    wall = (time.time() - t0) * 1000
    digest = hashlib.sha256('\n'.join(sorted(body.splitlines())).encode()).hexdigest()[:12] if ok else 'ERROR'
    rows.append((n, qid, ok, digest, len(body.splitlines()) if ok else 0, round(wall, 1), body[:200].replace('\n', ' ') if not ok else ''))
src = f"clusterAllReplicas({a.cluster}, system.query_log)" if a.cluster else "system.query_log"
query(a.host, f"SYSTEM FLUSH LOGS{' ON CLUSTER ' + a.cluster if a.cluster else ''}")
time.sleep(1)
out = []
for (n, qid, ok, digest, nrows, wall, err) in rows:
    stats = ['', '', '', '', '', '']
    if ok:
        q = f"""SELECT sumIf(read_rows, 1), sumIf(read_bytes, 1), maxIf(memory_usage, is_initial_query), maxIf(memory_usage, NOT is_initial_query),
                arrayStringConcat(arrayDistinct(arrayFlatten(groupArray(projections))), ','), countIf(NOT is_initial_query)
                FROM {src} WHERE initial_query_id = '{qid}' AND type = 'QueryFinish'"""
        okq, b, _ = query(a.host, q, None, 'default', '', None, 'TSV')
        stats = b.strip().split('\t') if okq else ['?'] * 6
    out.append([n, 'OK' if ok else 'ERR', digest, str(nrows), str(wall)] + stats + [err])
os.makedirs(os.path.join(os.path.dirname(__file__), '..', 'results', 'queries'), exist_ok=True)
path = os.path.join(os.path.dirname(__file__), '..', 'results', 'queries', a.label + '.tsv')
hdr = ['query', 'status', 'digest', 'result_rows', 'client_ms', 'read_rows', 'read_bytes', 'mem_initiator', 'mem_remote_max', 'projections', 'remote_subqueries', 'error']
with open(path, 'w') as f:
    f.write('\t'.join(hdr) + '\n')
    for r in out: f.write('\t'.join(r) + '\n')
for r in [hdr] + out: print('\t'.join(r[:11]) + ('\t' + r[11][:120] if len(r) > 11 and r[11] else ''))

#!/usr/bin/env python3
"""Load deterministic synthetic events. Usage:
  gen.py <host> <database> <total_rows> [--chunk 250000] [--filter SQL] [--tshift-ms 0] [--set k=v ...]
"""
import sys, os, argparse, time
sys.path.insert(0, os.path.dirname(__file__))
from chhttp import query
ap = argparse.ArgumentParser()
ap.add_argument('host'); ap.add_argument('db'); ap.add_argument('rows', type=int)
ap.add_argument('--chunk', type=int, default=250000); ap.add_argument('--filter', default='1')
ap.add_argument('--tshift-ms', type=int, default=0); ap.add_argument('--set', action='append', default=[])
ap.add_argument('--user', default='default'); ap.add_argument('--password', default='')
a = ap.parse_args()
tmpl = open(os.path.join(os.path.dirname(__file__), '..', 'sql', 'gen', 'usage_events_gen.sql')).read()
tmpl = "\n".join(l for l in tmpl.splitlines() if not l.strip().startswith('--'))
settings = dict(kv.split('=', 1) for kv in a.set)
t0 = time.time()
for off in range(0, a.rows, a.chunk):
    n = min(a.chunk, a.rows - off)
    sql = tmpl.replace('{N}', str(n)).replace('{OFFSET}', str(off)).replace('{FILTER}', a.filter).replace('{TSHIFT_MS}', str(a.tshift_ms))
    ok, body, hdr = query(a.host, sql, a.db, a.user, a.password, settings)
    if not ok:
        print('FAIL', body[:800]); sys.exit(1)
print(f'loaded rows {a.rows} filter={a.filter!r} tshift={a.tshift_ms} in {time.time()-t0:.1f}s')

#!/usr/bin/env python3
"""Send a .sql file one statement at a time over HTTP with wait_end_of_query=1, the team's
runner contract. Usage:
  run_sql.py <host> <database> <file> [--user U --password P] [--set k=v ...]
             [--engine-replace 'OLD|NEW' ...] [--continue]
"""
import re, sys, argparse, os
sys.path.insert(0, os.path.dirname(__file__))
from chhttp import query
ap = argparse.ArgumentParser()
ap.add_argument('host'); ap.add_argument('database'); ap.add_argument('file')
ap.add_argument('--user', default='default'); ap.add_argument('--password', default='')
ap.add_argument('--set', action='append', default=[])
ap.add_argument('--engine-replace', action='append', default=[])
ap.add_argument('--continue', dest='cont', action='store_true')
a = ap.parse_args()
s = open(a.file).read()
for r in a.engine_replace:
    old, new = r.split('|', 1); s = s.replace(old, new)
s = "\n".join(l for l in s.splitlines() if not l.strip().startswith('--'))
stmts = [p.strip().rstrip(';') for p in re.split(r';\s*\n', s + "\n") if p.strip()]
rc = 0
for st in stmts:
    ok, body, _ = query(a.host, st, a.database, a.user, a.password, dict(kv.split('=', 1) for kv in a.set))
    head = st.splitlines()[0][:100]
    if ok:
        print(f"OK    {head}")
        if body.strip(): print(body.rstrip()[:2000])
    else:
        rc = 1
        print(f"FAIL  {head}\n      {body.strip()[:600]}")
        if not a.cont: sys.exit(1)
sys.exit(rc)

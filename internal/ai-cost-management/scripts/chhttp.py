#!/usr/bin/env python3
"""Minimal ClickHouse HTTP client mirroring the service: HTTP interface, settings as URL
parameters (including custom settings such as SQL_tenant_id), wait_end_of_query=1.
Library + CLI:
  chhttp.py <host> "<sql>" [--db D] [--user U] [--password P] [--set k=v ...] [--fmt TSV]
"""
import sys, urllib.request, urllib.parse, urllib.error, argparse

def query(host, sql, db=None, user='default', password='', settings=None, fmt=None, port=8123, timeout=600):
    params = {'wait_end_of_query': '1'}
    if db: params['database'] = db
    for k, v in (settings or {}).items(): params[k] = str(v)
    if fmt and ' format ' not in sql.lower():
        sql = sql.rstrip().rstrip(';') + f' FORMAT {fmt}'
    url = f'http://{host}:{port}/?' + urllib.parse.urlencode(params)
    req = urllib.request.Request(url, data=sql.encode(), method='POST')
    req.add_header('X-ClickHouse-User', user)
    req.add_header('X-ClickHouse-Key', password)
    try:
        with urllib.request.urlopen(req, timeout=timeout) as r:
            return True, r.read().decode("utf-8", "replace"), dict(r.headers)
    except urllib.error.HTTPError as e:
        return False, e.read().decode("utf-8", "replace"), dict(e.headers)

if __name__ == '__main__':
    ap = argparse.ArgumentParser()
    ap.add_argument('host'); ap.add_argument('sql')
    ap.add_argument('--db'); ap.add_argument('--user', default='default'); ap.add_argument('--password', default='')
    ap.add_argument('--set', action='append', default=[]); ap.add_argument('--fmt')
    a = ap.parse_args()
    sql = open(a.sql[1:]).read() if a.sql.startswith('@') else a.sql
    ok, body, hdr = query(a.host, sql, a.db, a.user, a.password, dict(kv.split('=', 1) for kv in a.set), a.fmt)
    sys.stdout.write(body)
    if 'X-ClickHouse-Summary' in hdr: sys.stderr.write('summary: ' + hdr['X-ClickHouse-Summary'] + '\n')
    sys.exit(0 if ok else 1)

#!/usr/bin/env python3
"""Port the verbatim OTel DDL (docs/research/otel-ddl/) into scenario migrations.

Each upstream statement is copied from its "BEGIN VERBATIM" block and only these
placeholders are substituted (ADR-0004), each asserted to match exactly once:

  database/table names  -> {{ db }}.{{ table "<name>" }}  (shard-local name when sharded)
  cluster placeholder   -> {{ onCluster }}
  engine placeholder    -> {{ engine "<MergeTree family>" }}  (Replicated*/Shared* per target)
  TTL placeholder       -> exporter default: none (ttl 0); ClickStack: toIntervalDay(3650)
                           instead of 720h, so pinned benchmark windows never expire
  SETTINGS clause       -> {{ settings "<upstream list>" }}  (appends storage_policy on s3 targets)

Distributed tables are added after each table on sharded topologies only.
Run from the repository root; it rewrites scenarios/otel-*/*/init/*.tmpl.
"""
import os, re, sys

SRC = 'docs/research/otel-ddl'
TTL_DAYS = 3650

def verbatim(fname, index=0):
    text = open(os.path.join(SRC, fname)).read()
    blocks = re.findall(r'-- ===== BEGIN VERBATIM =====\n(.*?)-- ===== END VERBATIM =====', text, re.S)
    if not blocks:
        sys.exit(f'{fname}: no verbatim block')
    b = blocks[-1]
    # drop leading comment lines inside the block (upstream file comments stay in the source file)
    return b.strip()

def sub1(s, old, new, where):
    n = s.count(old)
    if n != 1:
        sys.exit(f'{where}: expected exactly one {old!r}, found {n}')
    return s.replace(old, new)

def resub1(s, pattern, new, where):
    s2, n = re.subn(pattern, new, s, flags=re.S)
    if n != 1:
        sys.exit(f'{where}: expected exactly one match of {pattern!r}, found {n}')
    return s2

def split_statements(b):
    return [x.strip() for x in b.split(';') if x.strip() and not all(l.strip().startswith('--') or not l.strip() for l in x.strip().splitlines())]

def header(sources, notes):
    lines = ['-- Ported by migrations/tools/port_otel_ddl.py from the verbatim upstream DDL; edit the',
             '-- script, not this file. Only placeholders differ from the source (ADR-0004).']
    for s in sources:
        lines.append(f'-- Source: docs/research/otel-ddl/{s} (its header cites the upstream file and commit)')
    for n in notes:
        lines.append(f'-- {n}')
    return '\n'.join(lines) + '\n'

def write(path, text):
    os.makedirs(os.path.dirname(path), exist_ok=True)
    with open(path, 'w') as f:
        f.write(text)

def dist(name, key):
    return f'{{{{ distributed "{name}" "{key}" }}}}'

def drop(names_tables, mvs=()):
    out = []
    for mv in mvs:
        out.append(f'DROP TABLE IF EXISTS {{{{ db }}}}.{mv} {{{{ onCluster }}}} SYNC;')
    for t in names_tables:
        out.append(f'{{{{ dropDistributed "{t}" }}}};')
        out.append(f'DROP TABLE IF EXISTS {{{{ db }}}}.{{{{ table "{t}" }}}} {{{{ onCluster }}}} SYNC;')
    return '\n'.join(out) + '\n'

VERSION = '20261008000000'

# ---------------------------------------------------------------- exporter-default
def exporter_traces():
    w = 'exporter traces'
    t = verbatim('exporter_otel_traces.sql')
    t = sub1(t, 'CREATE TABLE IF NOT EXISTS %q.%q %s (', 'CREATE TABLE IF NOT EXISTS {{ db }}.{{ table "otel_traces" }} {{ onCluster }} (', w)
    t = sub1(t, ') ENGINE = %s', ') ENGINE = {{ engine "MergeTree" }}', w)
    t = resub1(t, r'ORDER BY \(ServiceName, SpanName, toDateTime\(Timestamp\)\)\n%s\n', 'ORDER BY (ServiceName, SpanName, toDateTime(Timestamp))\n', w)
    t = sub1(t, 'SETTINGS index_granularity=8192, ttl_only_drop_parts = 1', '{{ settings "index_granularity=8192, ttl_only_drop_parts = 1" }}', w)
    ts, mv = split_statements(verbatim('exporter_otel_traces_trace_id_ts.sql'))
    w = 'exporter trace_id_ts'
    ts = sub1(ts, 'CREATE TABLE IF NOT EXISTS "%s"."%s_trace_id_ts" %s (', 'CREATE TABLE IF NOT EXISTS {{ db }}.{{ table "otel_traces_trace_id_ts" }} {{ onCluster }} (', w)
    ts = sub1(ts, ') ENGINE = %s', ') ENGINE = {{ engine "MergeTree" }}', w)
    ts = resub1(ts, r'ORDER BY \(TraceId, Start\)\n    %s\n', 'ORDER BY (TraceId, Start)\n', w)
    ts = sub1(ts, 'SETTINGS index_granularity=8192, ttl_only_drop_parts = 1', '{{ settings "index_granularity=8192, ttl_only_drop_parts = 1" }}', w)
    mv = sub1(mv, 'CREATE MATERIALIZED VIEW IF NOT EXISTS "%s"."%s_trace_id_ts_mv" %s', 'CREATE MATERIALIZED VIEW IF NOT EXISTS {{ db }}.otel_traces_trace_id_ts_mv {{ onCluster }}', w)
    mv = sub1(mv, 'TO "%s"."%s_trace_id_ts"', 'TO {{ db }}.{{ table "otel_traces_trace_id_ts" }}', w)
    mv = sub1(mv, 'FROM "%s"."%s"', 'FROM {{ db }}.{{ table "otel_traces" }}', w)
    up = header(['exporter_otel_traces.sql', 'exporter_otel_traces_trace_id_ts.sql'],
                ['Exporter defaults: json: false, ttl 0 (no TTL clause), engine MergeTree().']) + '\n' + \
        ';\n\n'.join([t, dist('otel_traces', 'cityHash64(TraceId)'), ts, dist('otel_traces_trace_id_ts', 'cityHash64(TraceId)'), mv]) + ';\n'
    base = 'scenarios/otel-traces/exporter-default/init/' + VERSION + '_create_otel_traces'
    write(base + '.up.sql.tmpl', up)
    write(base + '.down.sql.tmpl', drop(['otel_traces_trace_id_ts', 'otel_traces'], ['otel_traces_trace_id_ts_mv']))

def exporter_logs():
    w = 'exporter logs'
    t = verbatim('exporter_otel_logs.sql')
    t = sub1(t, 'CREATE TABLE IF NOT EXISTS {{ident .Database}}.{{ident .TableName}} {{.ClusterString}} (', 'CREATE TABLE IF NOT EXISTS {{ db }}.{{ table "otel_logs" }} {{ onCluster }} (', w)
    # >= 26.2 on every target: keep the HasFullTextSearch branch (text indexes)
    t = resub1(t, r'\{\{- if \.HasFullTextSearch\}\}\n(.*?)\{\{- else\}\}\n.*?\{\{- end\}\}\n', r'\1', w)
    t = sub1(t, ') ENGINE = {{.Engine}}', ') ENGINE = {{ engine "MergeTree" }}', w)
    t = resub1(t, r'\n\{\{\.TTL\}\}\n', '\n', w)
    t = sub1(t, 'SETTINGS index_granularity = 8192, ttl_only_drop_parts = 1', '{{ settings "index_granularity = 8192, ttl_only_drop_parts = 1" }}', w)
    up = header(['exporter_otel_logs.sql'], ['Exporter defaults: json: false, ttl 0; HasFullTextSearch branch (every target is >= 26.2).']) + '\n' + \
        ';\n\n'.join([t, dist('otel_logs', 'rand()')]) + ';\n'
    base = 'scenarios/otel-logs/exporter-default/init/' + VERSION + '_create_otel_logs'
    write(base + '.up.sql.tmpl', up)
    write(base + '.down.sql.tmpl', drop(['otel_logs']))

METRIC_FILES = [('gauge', 'otel_metrics_gauge'), ('sum', 'otel_metrics_sum'), ('histogram', 'otel_metrics_histogram'),
                ('exponential_histogram', 'otel_metrics_exponential_histogram'), ('summary', 'otel_metrics_summary')]

def exporter_metrics():
    stmts, names = [], []
    for kind, name in METRIC_FILES:
        w = f'exporter metrics {kind}'
        t = verbatim(f'exporter_otel_metrics_{kind}.sql')
        t = sub1(t, 'CREATE TABLE IF NOT EXISTS "%s"."%s" %s (', f'CREATE TABLE IF NOT EXISTS {{{{ db }}}}.{{{{ table "{name}" }}}} {{{{ onCluster }}}} (', w)
        t = resub1(t, r'\) ENGINE = %s\n    %s\n', ') ENGINE = {{ engine "MergeTree" }}\n', w)
        t = sub1(t, 'SETTINGS index_granularity=8192, ttl_only_drop_parts = 1', '{{ settings "index_granularity=8192, ttl_only_drop_parts = 1" }}', w)
        stmts += [t, dist(name, 'rand()')]
        names.append(name)
    up = header([f'exporter_otel_metrics_{k}.sql' for k, _ in METRIC_FILES], ['Exporter defaults: ttl 0, engine MergeTree().']) + '\n' + ';\n\n'.join(stmts) + ';\n'
    base = 'scenarios/otel-metrics/exporter-default/init/' + VERSION + '_create_otel_metrics'
    write(base + '.up.sql.tmpl', up)
    write(base + '.down.sql.tmpl', drop(names))

# ---------------------------------------------------------------- clickstack
def cs(t, name, engine, w, ttl_macro, settings):
    t = sub1(t, f'CREATE TABLE IF NOT EXISTS ${{DATABASE}}.{name}\n', f'CREATE TABLE IF NOT EXISTS {{{{ db }}}}.{{{{ table "{name}" }}}} {{{{ onCluster }}}}\n', w)
    t = sub1(t, f'ENGINE = {engine}\n', f'ENGINE = {{{{ engine "{engine}" }}}}\n', w)
    if ttl_macro:
        t = sub1(t, '${' + ttl_macro + '}', f'toIntervalDay({TTL_DAYS})', w)
    t = sub1(t, 'SETTINGS ' + settings, '{{ settings "' + settings + '" }}', w)
    return t.rstrip().rstrip(';')

def cs_mv(mv, src, w):
    m = re.search(r"CREATE MATERIALIZED VIEW IF NOT EXISTS \$\{DATABASE\}\.(\w+) TO \$\{DATABASE\}\.(\w+)\n", mv)
    if not m:
        sys.exit(f'{w}: unexpected MV header')
    mvname, target = m.group(1), m.group(2)
    mv = sub1(mv, m.group(0), f'CREATE MATERIALIZED VIEW IF NOT EXISTS {{{{ db }}}}.{mvname} {{{{ onCluster }}}} TO {{{{ db }}}}.{{{{ table "{target}" }}}}\n', w)
    mv = mv.replace('FROM ${DATABASE}.' + src, 'FROM {{ db }}.{{ table "' + src + '" }}')
    if '${DATABASE}' in mv:
        sys.exit(f'{w}: unsubstituted ${{DATABASE}} in MV')
    return mvname, target, mv

def clickstack_signal(workload, main, main_file, rollup_file, ttl_macro, main_settings, key):
    w = f'clickstack {main}'
    t = cs(verbatim(main_file), main, 'MergeTree', w, ttl_macro, main_settings)
    rt, mv = split_statements(verbatim(rollup_file))
    rname = main + '_kv_rollup_15m'
    rt = cs(rt, rname, 'SummingMergeTree', w + ' rollup', ttl_macro, 'index_granularity = 8192, ttl_only_drop_parts = 1')
    mvname, target, mv = cs_mv(mv, main, w + ' mv')
    up = header([main_file.replace('clickstack_', 'clickstack_'), rollup_file],
                [f'ClickStack TTL macro ${{{ttl_macro}}} (default 720h) rendered as toIntervalDay({TTL_DAYS}) so pinned windows never expire.']) + '\n' + \
        ';\n\n'.join([t, dist(main, key), rt, dist(rname, 'rand()'), mv]) + ';\n'
    base = f'scenarios/{workload}/clickstack/init/{VERSION}_create_{main}'
    write(base + '.up.sql.tmpl', up)
    write(base + '.down.sql.tmpl', drop([rname, main], [mvname]))

def clickstack_metrics():
    stmts, names = [], []
    for kind, name in METRIC_FILES:
        w = f'clickstack metrics {kind}'
        t = cs(verbatim(f'clickstack_otel_metrics_{kind}.sql'), name, 'MergeTree', w, 'METRICS_TTL', 'ttl_only_drop_parts = 1')
        stmts += [t, dist(name, 'rand()')]
        names.append(name)
    up = header([f'clickstack_otel_metrics_{k}.sql' for k, _ in METRIC_FILES],
                [f'ClickStack TTL macro ${{METRICS_TTL}} rendered as toIntervalDay({TTL_DAYS}).',
                 'Not ported: 00008_otel_metrics_timeseries.sql (experimental TimeSeries engine, not written by',
                 'the OTLP exporter; covered by a feature probe) and 00004_hyperdx_sessions.sql (browser sessions).']) + '\n' + ';\n\n'.join(stmts) + ';\n'
    base = f'scenarios/otel-metrics/clickstack/init/{VERSION}_create_otel_metrics'
    write(base + '.up.sql.tmpl', up)
    write(base + '.down.sql.tmpl', drop(names))

exporter_traces(); exporter_logs(); exporter_metrics()
clickstack_signal('otel-traces', 'otel_traces', 'clickstack_otel_traces.sql', 'clickstack_otel_traces_kv_rollup_15m.sql', 'TRACES_TTL',
                  'index_granularity = 8192, ttl_only_drop_parts = 1', 'cityHash64(TraceId)')
clickstack_signal('otel-logs', 'otel_logs', 'clickstack_otel_logs.sql', 'clickstack_otel_logs_kv_rollup_15m.sql', 'LOGS_TTL',
                  'index_granularity = 8192, ttl_only_drop_parts = 1, enable_block_number_column = 1, enable_block_offset_column = 1', 'rand()')
clickstack_metrics()
print('ported')

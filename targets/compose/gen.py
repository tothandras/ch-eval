#!/usr/bin/env python3
"""Generate a docker-compose deployment for a local target tier.

  gen.py <out_dir> --project NAME --image IMAGE --keeper-image IMAGE --port-base N
         [--shards S --replicas R] [--single] [--mem-bytes B]

Single: one ClickHouse node. Cluster: S shards x R replicas, 3 ClickHouse Keeper nodes,
a SeaweedFS S3 endpoint (for the s3 storage policy), an interserver secret, and a
`ch_eval` remote_servers cluster. Ports: native port_base+i, HTTP port_base-877+i for
node i (0 = first node). Correctness and smoke testing only (ADR-0005).
"""
import argparse, os, secrets, textwrap
ap = argparse.ArgumentParser()
ap.add_argument('out'); ap.add_argument('--project', required=True); ap.add_argument('--image', required=True)
ap.add_argument('--keeper-image', default=''); ap.add_argument('--port-base', type=int, required=True)
ap.add_argument('--shards', type=int, default=1); ap.add_argument('--replicas', type=int, default=2)
ap.add_argument('--single', action='store_true'); ap.add_argument('--mem-bytes', type=int, default=2_500_000_000)
ap.add_argument('--s3-image', default='chrislusf/seaweedfs:4.48')
a = ap.parse_args()
os.makedirs(os.path.join(a.out, 'config'), exist_ok=True)
def w(name, text):
    with open(os.path.join(a.out, 'config', name), 'w') as f: f.write(textwrap.dedent(text).lstrip())

w('common.xml', f"""
<clickhouse>
  <logger><level>information</level><console>1</console></logger>
  <max_server_memory_usage>{a.mem_bytes}</max_server_memory_usage>
  <mark_cache_size>268435456</mark_cache_size>
  <!-- custom settings prefix used by row policies (AICM) and generally harmless -->
  <custom_settings_prefixes>SQL_</custom_settings_prefixes>
  <!-- the COGS async-insert attribution reads this log -->
  <asynchronous_insert_log><database>system</database><table>asynchronous_insert_log</table><flush_interval_milliseconds>7500</flush_interval_milliseconds></asynchronous_insert_log>
</clickhouse>
""")
w('users.xml', """
<clickhouse>
  <users>
    <default>
      <password></password>
      <networks><ip>::/0</ip></networks>
      <profile>default</profile>
      <quota>default</quota>
      <access_management>1</access_management>
    </default>
  </users>
</clickhouse>
""")
STORAGE = '''  <storage_configuration>
    <disks>
      <s3>
        <type>s3</type>
        <endpoint>http://s3:8333/ch-eval/{{node}}/</endpoint>
        <access_key_id>local</access_key_id>
        <secret_access_key>local</secret_access_key>
        <metadata_path>/var/lib/clickhouse/disks/s3/</metadata_path>
        <skip_access_check>true</skip_access_check>
      </s3>
      <s3_cache>
        <type>cache</type>
        <disk>s3</disk>
        <path>/var/lib/clickhouse/disks/s3_cache/</path>
        <max_size>2Gi</max_size>
      </s3_cache>
    </disks>
    <policies>
      <s3_cached><volumes><main><disk>s3_cache</disk></main></volumes></s3_cached>
    </policies>
  </storage_configuration>
'''
nodes = []
if a.single:
    nodes = [('ch', None, None)]
else:
    for s in range(1, a.shards + 1):
        for r in range(1, a.replicas + 1):
            nodes.append((f'ch-s{s}r{r}', s, r))
    # interserver secret: random per deployment, kept in the git-ignored state dir
    secret_path = os.path.join(a.out, 'interserver-secret')
    if not os.path.exists(secret_path):
        with open(secret_path, 'w') as f: f.write(secrets.token_hex(16))
    secret = open(secret_path).read().strip()
    shards_xml = ''
    for s in range(1, a.shards + 1):
        reps = ''.join(f'<replica><host>ch-s{s}r{r}</host><port>9000</port></replica>' for r in range(1, a.replicas + 1))
        shards_xml += f'<shard><internal_replication>true</internal_replication>{reps}</shard>\n      '
    w('cluster.xml', f"""
<clickhouse>
  <remote_servers replace="true">
    <ch_eval>
      <secret>{secret}</secret>
      {shards_xml}
    </ch_eval>
  </remote_servers>
  <zookeeper>
    <node><host>keeper1</host><port>9181</port></node>
    <node><host>keeper2</host><port>9181</port></node>
    <node><host>keeper3</host><port>9181</port></node>
  </zookeeper>
  <distributed_ddl><path>/clickhouse/task_queue/ddl</path></distributed_ddl>
</clickhouse>
""")
    for n, s, r in nodes:
        storage = STORAGE.replace('{{node}}', n)
        w(f'macros-{n}.xml', f'<clickhouse><macros><cluster>ch_eval</cluster><shard>{s}</shard><replica>{n}</replica></macros>\n{storage}</clickhouse>\n')
    w('s3.json', '{"identities": [{"name": "local", "credentials": [{"accessKey": "local", "secretKey": "local"}], "actions": ["Admin", "Read", "Write", "List", "Tagging"]}]}\n')
    for k in (1, 2, 3):
        raft = ''.join(f'<server><id>{i}</id><hostname>keeper{i}</hostname><port>9234</port></server>' for i in (1, 2, 3))
        w(f'keeper{k}.xml', f"""
<clickhouse>
  <logger><level>information</level><console>1</console></logger>
  <listen_host>0.0.0.0</listen_host>
  <keeper_server>
    <tcp_port>9181</tcp_port>
    <server_id>{k}</server_id>
    <log_storage_path>/var/lib/clickhouse/coordination/log</log_storage_path>
    <snapshot_storage_path>/var/lib/clickhouse/coordination/snapshots</snapshot_storage_path>
    <raft_configuration>{raft}</raft_configuration>
  </keeper_server>
</clickhouse>
""")

svc = []
if not a.single:
    for k in (1, 2, 3):
        svc.append(f"""  keeper{k}:
    image: {a.keeper_image}
    hostname: keeper{k}
    restart: on-failure
    volumes:
      - ./config/keeper{k}.xml:/etc/clickhouse-keeper/keeper_config.xml:ro
""")
    svc.append(f"""  s3:
    image: {a.s3_image}
    hostname: s3
    command: ["server", "-s3", "-s3.config=/etc/seaweedfs/s3.json", "-dir=/data", "-volume.max=0", "-master.volumeSizeLimitMB=256"]
    volumes:
      - ./config/s3.json:/etc/seaweedfs/s3.json:ro
""")
for i, (n, s, r) in enumerate(nodes):
    vols = ["./config/common.xml:/etc/clickhouse-server/config.d/common.xml:ro",
            "./config/users.xml:/etc/clickhouse-server/users.d/zz-default.xml:ro"]
    deps = ''
    if not a.single:
        vols += ["./config/cluster.xml:/etc/clickhouse-server/config.d/cluster.xml:ro",
                 f"./config/macros-{n}.xml:/etc/clickhouse-server/config.d/macros.xml:ro"]
        deps = "    depends_on: [keeper1, keeper2, keeper3, s3]\n"
    vol_yaml = ''.join(f"      - {v}\n" for v in vols)
    svc.append(f"""  {n}:
    image: {a.image}
    hostname: {n}
    restart: on-failure
    environment: {{CLICKHOUSE_SKIP_USER_SETUP: "1"}}
{deps}    ports: ["{a.port_base + i}:9000", "{a.port_base - 877 + i}:8123"]
    volumes:
{vol_yaml}""")
with open(os.path.join(a.out, 'docker-compose.yml'), 'w') as f:
    f.write(f"# generated by targets/compose/gen.py; do not edit\nname: {a.project}\nservices:\n" + ''.join(svc))
with open(os.path.join(a.out, 'nodes'), 'w') as f:
    f.write('\n'.join(n for n, _, _ in nodes) + '\n')

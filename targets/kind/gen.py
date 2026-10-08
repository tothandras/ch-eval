#!/usr/bin/env python3
"""Render the kind cluster config, CHK and CHI for a kind tier manifest.

  gen.py <manifest.json> <out_dir>

Writes kind-config.yaml (with the sandbox patches only on cgroup v1 hosts),
chk.yaml (3-node ClickHouse Keeper) and chi.yaml (the manifest's shards x
replicas, cluster ch_eval, default user without password for local use).
Correctness and smoke testing only (ADR-0005).
"""
import json, os, sys

man = json.load(open(sys.argv[1]))
out = sys.argv[2]
os.makedirs(out, exist_ok=True)

cgroup_v1 = os.popen('stat -fc %T /sys/fs/cgroup').read().strip() != 'cgroup2fs'
cfg = ['kind: Cluster', 'apiVersion: kind.x-k8s.io/v1alpha4']
if cgroup_v1:
    # kind on a cgroup v1 host without CAP_SYS_RESOURCE (this sandbox): kubelet
    # refuses cgroup v1 and containerd cannot set oom_score_adj otherwise.
    cfg += ['containerdConfigPatches:', '- |-', '  [plugins."io.containerd.grpc.v1.cri"]', '    restrict_oom_score_adj = true']
cfg += ['nodes:', '- role: control-plane']
if cgroup_v1:
    cfg += ['  kubeadmConfigPatches:', '  - |', '    kind: KubeletConfiguration', '    failCgroupV1: false']
open(os.path.join(out, 'kind-config.yaml'), 'w').write('\n'.join(cfg) + '\n')

open(os.path.join(out, 'chk.yaml'), 'w').write(f'''apiVersion: clickhouse-keeper.altinity.com/v1
kind: ClickHouseKeeperInstallation
metadata:
  name: keeper
spec:
  defaults:
    templates:
      podTemplate: keeper
      dataVolumeClaimTemplate: keeper-data
  configuration:
    clusters:
      - name: keeper
        layout:
          replicasCount: {man.get("keepers", 3)}
  templates:
    podTemplates:
      - name: keeper
        spec:
          containers:
            - name: clickhouse-keeper
              image: {man["keeper_image"]}
              imagePullPolicy: IfNotPresent
              resources:
                requests: {{cpu: 100m, memory: 256Mi}}
                limits: {{memory: 512Mi}}
    volumeClaimTemplates:
      - name: keeper-data
        spec:
          accessModes: [ReadWriteOnce]
          resources:
            requests: {{storage: 1Gi}}
''')

mem = man.get('mem_bytes_per_node', 1500000000)
open(os.path.join(out, 'chi.yaml'), 'w').write(f'''apiVersion: clickhouse.altinity.com/v1
kind: ClickHouseInstallation
metadata:
  name: cheval
spec:
  defaults:
    templates:
      podTemplate: ch
      dataVolumeClaimTemplate: data
  configuration:
    zookeeper:
      keeper:
        name: keeper
    users:
      default/networks/ip: "::/0"
      default/password: ""
      default/access_management: 1
    files:
      config.d/ch-eval.xml: |
        <clickhouse>
          <logger><level>warning</level></logger>
          <max_server_memory_usage>{mem}</max_server_memory_usage>
          <custom_settings_prefixes>SQL_</custom_settings_prefixes>
          <asynchronous_insert_log><database>system</database><table>asynchronous_insert_log</table><flush_interval_milliseconds>7500</flush_interval_milliseconds></asynchronous_insert_log>
          <trace_log remove="1"/>
          <text_log remove="1"/>
          <metric_log remove="1"/>
          <asynchronous_metric_log remove="1"/>
          <latency_log remove="1"/>
          <query_metric_log remove="1"/>
          <processors_profile_log remove="1"/>
          <query_thread_log remove="1"/>
          <error_log remove="1"/>
          <session_log remove="1"/>
        </clickhouse>
    clusters:
      - name: {man["cluster"]}
        secret:
          auto: "true"
        layout:
          shardsCount: {man["shards"]}
          replicasCount: {man["replicas"]}
  templates:
    podTemplates:
      - name: ch
        spec:
          containers:
            - name: clickhouse
              image: {man["image"]}
              imagePullPolicy: IfNotPresent
              resources:
                requests: {{cpu: 250m, memory: 512Mi}}
                limits: {{memory: {int(mem * 1.3) // (1024 * 1024)}Mi}}
    volumeClaimTemplates:
      - name: data
        spec:
          accessModes: [ReadWriteOnce]
          resources:
            requests: {{storage: 4Gi}}
''')
print(out, 'cgroup_v1' if cgroup_v1 else 'cgroup_v2')

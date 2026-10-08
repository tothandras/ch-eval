#!/usr/bin/env bash
# experiments/kind/e5-e6.sh <platform> [rows]
# Operator mechanics on the kind tier (design D15 E5/E6, D13 RS-12), local smoke only:
#   E6 add a replica (replicasCount 2 -> 3): does the operator create the schema on
#      the new replica, and how long until it has fetched every row?
#   E5 vertical change (pod memory limit): how long the operator's rolling restart
#      takes, one host at a time.
# Writes results/experiments/E6|E5/<platform>/kind/<run-id>/report.md.
# Availability during the roll needs an in-cluster load generator (Phase 6 on EKS).
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
platform="${1:?usage: e5-e6.sh <platform> [rows]}"; rows="${2:-1000000}"
state="$ROOT/targets/.state/kind-$platform-kind"
export KUBECONFIG="$state/kubeconfig"
native=$(python3 -c "import json; print(json.load(open('$ROOT/targets/$platform/kind.json'))['ports']['native_base'])")
q() { clickhouse client --port "$native" -q "$1"; }
now() { date +%s.%N; }
since() { python3 -c "print(round(($(now) - $1) * 1000))"; }
runid=$(date -u +%Y%m%dT%H%M%SZ)-$(head -c2 /dev/urandom | od -An -tx1 | tr -d ' \n')
version=$(q "SELECT version()")

# --- E6
q "DROP DATABASE IF EXISTS ch_eval_e6 ON CLUSTER 'ch-eval' SYNC"
q "CREATE DATABASE ch_eval_e6 ON CLUSTER 'ch-eval'"
q "CREATE TABLE ch_eval_e6.t ON CLUSTER 'ch-eval' (k UInt64, v String) ENGINE = ReplicatedMergeTree('/clickhouse/tables/{shard}/ch_eval_e6/t', '{replica}') ORDER BY k"
q "INSERT INTO ch_eval_e6.t SELECT number, toString(number) FROM numbers($rows)"
t0=$(now)
kubectl patch chi cheval --type=json -p='[{"op":"replace","path":"/spec/configuration/clusters/0/layout/replicasCount","value":3}]' >/dev/null
pod_ready=""; schema=""; data=""
for _ in $(seq 1 600); do
  hosts=$(q "SELECT count() FROM system.clusters WHERE cluster = 'ch-eval'" 2>/dev/null || echo 0)
  if [ -z "$pod_ready" ] && [ "$hosts" = 3 ]; then pod_ready=$(since "$t0"); fi
  if [ -n "$pod_ready" ]; then
    has=$(q "SELECT count() FROM clusterAllReplicas('ch-eval', system.tables) WHERE database = 'ch_eval_e6' AND name = 't'" 2>/dev/null || echo 0)
    [ -z "$schema" ] && [ "$has" = 3 ] && schema=$(since "$t0")
    if [ -n "$schema" ]; then
      full=$(q "SELECT countIf(c = $rows) FROM (SELECT hostName() h, count() c FROM clusterAllReplicas('ch-eval', ch_eval_e6.t) GROUP BY h)" 2>/dev/null || echo 0)
      [ "$full" = 3 ] && { data=$(since "$t0"); break; }
    fi
  fi
  sleep 1
done
status=$(kubectl get chi cheval -o jsonpath='{.status.status}')
dir="$ROOT/results/experiments/E6/$platform/kind/$runid"; mkdir -p "$dir"
cat > "$dir/report.md" <<MD
# E6: add a replica through the operator (replicasCount 2 -> 3)

Target \`$platform/kind\` (\`$version\`), operator 0.27.4, $rows rows of (UInt64, String). Local tier: mechanics only, timings are not production evidence.

| Milestone after the CHI patch | ms |
| --- | ---: |
| third host listed in the \`ch-eval\` cluster | ${pod_ready:-never} |
| table present on all three replicas (created by the operator, no DDL issued) | ${schema:-never} |
| all $rows rows fetched on the new replica | ${data:-never} |

CHI status at the end: $status.
MD
cat "$dir/report.md"

# --- E5
t0=$(now)
kubectl patch chi cheval --type=json -p='[{"op":"replace","path":"/spec/templates/podTemplates/0/spec/containers/0/resources/limits/memory","value":"1700Mi"}]' >/dev/null
sleep 5
for _ in $(seq 1 900); do
  [ "$(kubectl get chi cheval -o jsonpath='{.status.status}')" = "Completed" ] && break
  sleep 2
done
roll=$(since "$t0")
restarts=$(kubectl get pods -l clickhouse.altinity.com/chi=cheval -o jsonpath='{range .items[*]}{.metadata.name}{" started "}{.status.startTime}{"\n"}{end}')
dir="$ROOT/results/experiments/E5/$platform/kind/$runid"; mkdir -p "$dir"
cat > "$dir/report.md" <<MD
# E5: vertical change through the operator (pod memory limit)

Target \`$platform/kind\` (\`$version\`), 3 replicas. Local tier: mechanics only.

- CHI patch to \`Completed\`: ${roll} ms (the operator rolls one host at a time).
- Pod start times after the roll:

\`\`\`
$restarts
\`\`\`
MD
cat "$dir/report.md"
# Restore the tier: re-applying the rendered CHI returns to 2 replicas and the
# original limits (another roll), and re-creates the port-forwards the roll broke.
"$ROOT/targets/kind/up" "$platform" kind >/dev/null
q "DROP DATABASE IF EXISTS ch_eval_e6 ON CLUSTER 'ch-eval' SYNC" || true

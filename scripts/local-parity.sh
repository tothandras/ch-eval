#!/usr/bin/env bash
# scripts/local-parity.sh [targets...]
# Phase 2 local evidence: run every workload scenario on each local target with a
# pinned window and seed, then write the cross-target parity report
# results/parity/local-<date>.md. Smoke and parity only, never performance.
set -uo pipefail
cd "$(dirname "$0")/.."
TARGETS=("${@:-altinity/local-single upstream-oss/local-single altinity/local-cluster upstream-oss/local-cluster}")
read -r -a TARGETS <<< "${TARGETS[*]}"
BENCH=${BENCH:-bin/bench}
TIME_END=${TIME_END:-2026-06-01T00:00:00Z}
OM_ROWS=${OM_ROWS:-1000000}
OTEL_ROWS=${OTEL_ROWS:-500000}
for t in "${TARGETS[@]}"; do
  for w in openmeter otel-traces otel-logs otel-metrics; do
    rows=$OTEL_ROWS; [ "$w" = openmeter ] && rows=$OM_ROWS
    echo "== $t $w ($rows rows)"
    "$BENCH" --target "$t" --workload "$w" --rows "$rows" --iterations 3 --time-end "$TIME_END" --require-clean 2>&1 | grep -E "ERR|^error|seed done|wrote"
  done
done
out=results/parity/local-$(date -u +%Y%m%d).md
mkdir -p results/parity
{
  echo "# Local cross-target parity ($(date -u +%Y-%m-%d))"
  echo
  echo "Smoke and parity only, never performance. Targets: ${TARGETS[*]}. Window ends $TIME_END; seed 42;"
  echo "openmeter $OM_ROWS rows, OTel $OTEL_ROWS rows per scenario. Each section is \`bench compare\` output."
  for w in openmeter otel-traces otel-logs otel-metrics; do
    for s in $(ls scenarios/$w); do
      [ -d "scenarios/$w/$s/init" ] || continue
      echo; echo "---"; echo
      targets_csv=$(IFS=,; echo "${TARGETS[*]}")
      "$BENCH" compare --workload "$w" --scenario "$s" --targets "$targets_csv" > /tmp/ch-eval-compare.$$ 2>&1
      status=$?
      sed 's/^# Compare/## Compare/' /tmp/ch-eval-compare.$$; rm -f /tmp/ch-eval-compare.$$
      echo; echo "compare exit status: $status (non-zero = parity failed or not established)"
    done
  done
} > "$out"
echo "wrote $out"
grep -E "MATCH=|DIFFERS\*|not established" "$out" | sort | uniq -c

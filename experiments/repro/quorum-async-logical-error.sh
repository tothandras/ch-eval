#!/usr/bin/env bash
# experiments/repro/quorum-async-logical-error.sh <native-port> <cluster>
# Minimal reproduction: an async insert with insert_quorum and
# insert_quorum_parallel=0 is rejected with a proper error (code 2) when the data
# is in the query (VALUES), but reaches LOGICAL_ERROR (code 49, "Should be checked
# earlier") when the data is sent as a block (any FORMAT clause; clickhouse-go's
# batch insert). Found by the D7 consistency experiment on 26.3.33 (both builds).
set -uo pipefail
port="${1:?port}"; cluster="${2:?cluster}"
q() { clickhouse client --port "$port" "$@"; }
q -q "CREATE DATABASE IF NOT EXISTS ch_eval_repro ON CLUSTER '$cluster'" >/dev/null
q -q "CREATE TABLE IF NOT EXISTS ch_eval_repro.t ON CLUSTER '$cluster' (k UInt8) ENGINE = ReplicatedMergeTree('/clickhouse/tables/{shard}/ch_eval_repro/t', '{replica}') ORDER BY k" >/dev/null
s=(--async_insert 1 --wait_for_async_insert 1 --insert_quorum 2 --insert_quorum_parallel 0)
echo "server: $(q -q 'SELECT version()')"
echo "VALUES:       $(q "${s[@]}" -q "INSERT INTO ch_eval_repro.t VALUES (1)" < /dev/null 2>&1 | grep -o 'Code: [0-9]*' | head -1)"
echo "FORMAT TSV:   $(echo 1 | q "${s[@]}" -q "INSERT INTO ch_eval_repro.t FORMAT TSV" 2>&1 | grep -o 'Code: [0-9]*\|LOGICAL_ERROR' | head -2 | tr '\n' ' ')"
q -q "DROP DATABASE ch_eval_repro ON CLUSTER '$cluster' SYNC" >/dev/null

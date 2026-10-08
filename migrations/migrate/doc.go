// Package migrate is the evaluation's ClickHouse migration tool (ADR-0003).
//
// One source of truth, many targets: migration files follow golang-migrate naming
// (YYYYMMDDHHMMSS_name.up.sql / .down.sql, optionally with a .tmpl suffix) and are
// rendered per target with a closed set of template helpers, so dialect and topology
// differences (single node, ReplicatedMergeTree + ON CLUSTER, sharded Distributed
// tables, ClickHouse Cloud) never fork the files. Rendered output is plain SQL that
// golang-migrate could apply.
//
// The runner records state in a cluster-global schema_migrations_ch table
// (checksum, dirty/pending flags, duration), waits for mutations cluster-wide, and
// can capture per-statement resource cost from query_log/part_log so online schema
// changes can be compared across targets.
package migrate

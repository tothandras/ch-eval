package migrate

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
)

// StateTable records applied migrations. One row per transition (kept as history);
// the latest row per (scope, version) by applied_at is the state.
const StateTable = "schema_migrations_ch"

// Status values stored in the state table.
const (
	StatusApplied = "applied" // all statements succeeded
	StatusDirty   = "dirty"   // a statement failed; blocks further migrations until resolved
	StatusPending = "pending" // replicated DDL timed out (code 159): applied on active hosts, queued on the rest
	StatusDown    = "down"    // rolled back by its down migration
)

// Runner applies migrations of one scope (for example "openmeter/proposal") to one target.
type Runner struct {
	Conn   driver.Conn
	Topo   Topology
	Scope  string
	Target string // free-form target label recorded with each row
	RunID  string
	// DDLTimeout bounds how long replicated DDL waits for all hosts before the
	// statement is recorded as pending (distributed_ddl_task_timeout).
	DDLTimeout time.Duration
	// MutationTimeout bounds the cluster-wide wait for mutations to finish.
	MutationTimeout time.Duration
	// Measure captures per-statement cost from query_log/part_log.
	Measure bool
	Logf    func(string, ...any)
}

// State is the recorded state of one version.
type State struct {
	Version   uint64
	Name      string
	Checksum  string
	Status    string
	Topology  string // topology the version was rendered for
	AppliedAt time.Time
}

// StatementResult is one executed statement and, with Measure, its cost.
type StatementResult struct {
	SQL        string         `json:"sql"`
	WallMs     float64        `json:"wall_ms"`
	MutationMs float64        `json:"mutation_wait_ms,omitempty"`
	Error      string         `json:"error,omitempty"`
	Cost       *StatementCost `json:"cost,omitempty"`
}

// StatementCost is the resource cost of one statement across all replicas.
type StatementCost struct {
	QueryCPUSec    float64 `json:"query_cpu_sec"`
	MutateCPUSec   float64 `json:"mutate_cpu_sec"`
	MutateParts    uint64  `json:"mutate_parts"`
	ReadBytes      uint64  `json:"read_bytes"`
	WrittenBytes   uint64  `json:"written_bytes"`
	MutateReadB    uint64  `json:"mutate_read_bytes"`
	MutateWrittenB uint64  `json:"mutate_written_bytes"`
}

// Applied describes one migration step the runner executed.
type Applied struct {
	Version    uint64            `json:"version"`
	Name       string            `json:"name"`
	Direction  Direction         `json:"direction"`
	Status     string            `json:"status"`
	Checksum   string            `json:"checksum"`
	DurationMs float64           `json:"duration_ms"`
	Statements []StatementResult `json:"statements"`
}

func (r *Runner) logf(f string, a ...any) {
	if r.Logf != nil {
		r.Logf(f, a...)
	}
}

func (r *Runner) stateDDL() string {
	return fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s.%s %s (
  scope String,
  version UInt64,
  name String,
  checksum String,
  status LowCardinality(String),
  topology String,
  target String,
  run_id String,
  duration_ms UInt64,
  applied_at DateTime64(3, 'UTC')
) ENGINE = %s ORDER BY (scope, version)`,
		r.Topo.Database, StateTable, funcs(r.Topo)["onCluster"].(func() string)(),
		engine(r.Topo, "MergeTree", true, nil))
}

// EnsureState creates the database (single/cloud only) and the state table.
func (r *Runner) EnsureState(ctx context.Context) error {
	if err := r.Topo.Validate(); err != nil {
		return err
	}
	if r.Topo.Dialect != Replicated {
		if err := r.Conn.Exec(ctx, fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s", r.Topo.Database)); err != nil {
			return fmt.Errorf("migrate: create database: %w", err)
		}
	} else {
		if err := r.exec(ctx, fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s ON CLUSTER '%s'", r.Topo.Database, r.Topo.Cluster), "state"); err != nil {
			return fmt.Errorf("migrate: create database: %w", err)
		}
	}
	if err := r.exec(ctx, r.stateDDL(), "state"); err != nil {
		return fmt.Errorf("migrate: create state table: %w", err)
	}
	return nil
}

// Status returns the latest recorded state per version for the scope. On the
// replicated dialect the local replica first syncs the state table; on Cloud the
// read uses select_sequential_consistency so a just-written row is visible.
func (r *Runner) Status(ctx context.Context) (map[uint64]State, error) {
	sctx := ctx
	switch r.Topo.Dialect {
	case Replicated:
		if err := r.Conn.Exec(ctx, fmt.Sprintf("SYSTEM SYNC REPLICA %s.%s", r.Topo.Database, StateTable)); err != nil {
			return nil, fmt.Errorf("migrate: sync state replica: %w", err)
		}
	case Cloud:
		sctx = clickhouse.Context(ctx, clickhouse.WithSettings(clickhouse.Settings{"select_sequential_consistency": 1}))
	}
	rows, err := r.Conn.Query(sctx, fmt.Sprintf(`
SELECT version, argMax(name, applied_at), argMax(checksum, applied_at), argMax(status, applied_at), argMax(topology, applied_at), max(applied_at)
FROM %s.%s WHERE scope = ? GROUP BY version ORDER BY version`, r.Topo.Database, StateTable), r.Scope)
	if err != nil {
		return nil, fmt.Errorf("migrate: read state: %w", err)
	}
	defer rows.Close()
	out := map[uint64]State{}
	for rows.Next() {
		var s State
		if err := rows.Scan(&s.Version, &s.Name, &s.Checksum, &s.Status, &s.Topology, &s.AppliedAt); err != nil {
			return nil, err
		}
		out[s.Version] = s
	}
	return out, rows.Err()
}

func (r *Runner) record(ctx context.Context, m Migration, checksum, status string, dur time.Duration) error {
	return r.Conn.Exec(ctx, fmt.Sprintf(`INSERT INTO %s.%s (scope, version, name, checksum, status, topology, target, run_id, duration_ms, applied_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, now64(3))`, r.Topo.Database, StateTable),
		r.Scope, m.Version, m.Name, checksum, status, r.Topo.String(), r.Target, r.RunID, uint64(dur.Milliseconds()))
}

// ErrBlocked is returned when a dirty or pending version blocks further work.
var ErrBlocked = errors.New("migrate: blocked by a dirty or pending migration")

// Up applies every migration not yet applied, in version order. An applied version
// whose rendered checksum changed is an error (applied migrations are append-only).
func (r *Runner) Up(ctx context.Context, migs []Migration) ([]Applied, error) {
	state, err := r.Status(ctx)
	if err != nil {
		return nil, err
	}
	var done []Applied
	for _, m := range migs {
		rend, err := Render(m, Up, r.Topo)
		if err != nil {
			return done, err
		}
		if s, ok := state[m.Version]; ok {
			switch s.Status {
			case StatusApplied:
				if s.Checksum != rend.Checksum {
					if s.Topology != r.Topo.String() {
						return done, fmt.Errorf("migrate: %s/%d_%s was applied for topology %s but this target renders %s; roll it back first (bench migrate down, or a fresh bench run)", r.Scope, m.Version, m.Name, s.Topology, r.Topo.String())
					}
					return done, fmt.Errorf("migrate: %s/%d_%s was applied with a different checksum (edited after apply? ship a new migration instead)", r.Scope, m.Version, m.Name)
				}
				continue
			case StatusDirty, StatusPending:
				return done, fmt.Errorf("%w: %s/%d_%s is %s; resolve it (see `bench migrate status`) before continuing", ErrBlocked, r.Scope, m.Version, m.Name, s.Status)
			}
		}
		if lint := Lint(rend.Statements, r.Topo); len(lint) > 0 {
			return done, fmt.Errorf("migrate: lint %s: %s", m.UpPath, strings.Join(lint, "; "))
		}
		a, err := r.apply(ctx, m, Up, rend)
		done = append(done, a)
		if err != nil {
			return done, err
		}
	}
	return done, nil
}

// Down rolls back applied migrations newer than target (target 0 = all), newest first.
// A migration without a down file stops the rollback.
func (r *Runner) Down(ctx context.Context, migs []Migration, target uint64) ([]Applied, error) {
	state, err := r.Status(ctx)
	if err != nil {
		return nil, err
	}
	var done []Applied
	for i := len(migs) - 1; i >= 0; i-- {
		m := migs[i]
		if m.Version <= target {
			break
		}
		s, ok := state[m.Version]
		if !ok || s.Status == StatusDown {
			continue
		}
		if s.Status != StatusApplied {
			return done, fmt.Errorf("%w: %s/%d_%s is %s", ErrBlocked, r.Scope, m.Version, m.Name, s.Status)
		}
		if m.DownPath == "" {
			return done, fmt.Errorf("migrate: %s/%d_%s has no down migration", r.Scope, m.Version, m.Name)
		}
		rend, err := Render(m, Down, r.Topo)
		if err != nil {
			return done, err
		}
		a, err := r.apply(ctx, m, Down, rend)
		done = append(done, a)
		if err != nil {
			return done, err
		}
	}
	return done, nil
}

// Force overwrites the recorded status of a version (operator escape hatch after
// manual resolution of a dirty or pending migration).
func (r *Runner) Force(ctx context.Context, m Migration, status string) error {
	rend, err := Render(m, Up, r.Topo)
	if err != nil {
		return err
	}
	return r.record(ctx, m, rend.Checksum, status, 0)
}

func (r *Runner) apply(ctx context.Context, m Migration, dir Direction, rend Rendered) (Applied, error) {
	a := Applied{Version: m.Version, Name: m.Name, Direction: dir, Checksum: rend.Checksum}
	start := time.Now()
	r.logf("  %s %s/%d_%s (%d statements)", dir, r.Scope, m.Version, m.Name, len(rend.Statements))
	for i, st := range rend.Statements {
		res := r.runStatement(ctx, m, i, st)
		a.Statements = append(a.Statements, res)
		if res.Error != "" {
			a.DurationMs = float64(time.Since(start).Milliseconds())
			status := StatusDirty
			if isDDLTimeout(res.Error) {
				status = StatusPending
			}
			a.Status = status
			if err := r.record(ctx, m, rend.Checksum, status, time.Since(start)); err != nil {
				r.logf("  warning: could not record %s state: %v", status, err)
			}
			return a, fmt.Errorf("migrate: %s/%d_%s statement %d: %s (recorded as %s)", r.Scope, m.Version, m.Name, i+1, res.Error, status)
		}
	}
	a.DurationMs = float64(time.Since(start).Milliseconds())
	a.Status = StatusApplied
	if dir == Down {
		a.Status = StatusDown
	}
	return a, r.record(ctx, m, rend.Checksum, a.Status, time.Since(start))
}

func isDDLTimeout(msg string) bool {
	return strings.Contains(msg, "code: 159") || strings.Contains(msg, "TIMEOUT_EXCEEDED") || strings.Contains(msg, "Code: 159")
}

type logComment struct {
	Run       string `json:"migrate_run"`
	Migration string `json:"migration"`
	Stmt      int    `json:"stmt"`
}

func (r *Runner) settings(lc string) clickhouse.Settings {
	s := clickhouse.Settings{"log_comment": lc, "alter_sync": 2, "mutations_sync": 2}
	if r.Topo.Dialect == Replicated && r.DDLTimeout > 0 {
		s["distributed_ddl_task_timeout"] = int(r.DDLTimeout.Seconds())
	}
	return s
}

func (r *Runner) exec(ctx context.Context, sql, tag string) error {
	lc, _ := json.Marshal(logComment{Run: r.RunID, Migration: r.Scope + "/" + tag})
	return r.Conn.Exec(clickhouse.Context(ctx, clickhouse.WithSettings(r.settings(string(lc)))), sql)
}

func (r *Runner) runStatement(ctx context.Context, m Migration, i int, sql string) StatementResult {
	lcb, _ := json.Marshal(logComment{Run: r.RunID, Migration: fmt.Sprintf("%s/%d", r.Scope, m.Version), Stmt: i + 1})
	lc := string(lcb)
	res := StatementResult{SQL: sql}
	t0 := time.Now()
	err := r.Conn.Exec(clickhouse.Context(ctx, clickhouse.WithSettings(r.settings(lc))), sql)
	res.WallMs = float64(time.Since(t0).Microseconds()) / 1000
	if err != nil {
		res.Error = err.Error()
		return res
	}
	// mutations_sync=2 is not trusted on its own (observed on Cloud: returned before
	// the part mutation ran); wait cluster-wide and fail on latest_fail_reason.
	t1 := time.Now()
	if err := r.waitMutations(ctx); err != nil {
		res.Error = err.Error()
	}
	res.MutationMs = float64(time.Since(t1).Microseconds()) / 1000
	if r.Measure {
		if c, err := r.cost(ctx, lc, t0, time.Now()); err == nil {
			res.Cost = c
		} else {
			r.logf("  warning: cost capture: %v", err)
		}
	}
	return res
}

func (r *Runner) sysTable(name string) string {
	if r.Topo.Dialect == Replicated {
		return fmt.Sprintf("clusterAllReplicas('%s', system.%s)", r.Topo.Cluster, name)
	}
	if r.Topo.Dialect == Cloud {
		return fmt.Sprintf("clusterAllReplicas(default, system.%s)", name)
	}
	return "system." + name
}

func (r *Runner) waitMutations(ctx context.Context) error {
	timeout := r.MutationTimeout
	if timeout <= 0 {
		timeout = 30 * time.Minute
	}
	deadline := time.Now().Add(timeout)
	for {
		var pending uint64
		var fail string
		err := r.Conn.QueryRow(ctx, fmt.Sprintf(`SELECT countIf(NOT is_done), anyIf(latest_fail_reason, latest_fail_reason != '')
FROM %s WHERE database = ?`, r.sysTable("mutations")), r.Topo.Database).Scan(&pending, &fail)
		if err != nil {
			return fmt.Errorf("migrate: poll mutations: %w", err)
		}
		if fail != "" {
			return fmt.Errorf("migrate: mutation failed: %s", fail)
		}
		if pending == 0 {
			return nil
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("migrate: %d mutation(s) still running after %s", pending, timeout)
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(time.Second):
		}
	}
}

func (r *Runner) cost(ctx context.Context, lc string, from, to time.Time) (*StatementCost, error) {
	flush := "SYSTEM FLUSH LOGS"
	if r.Topo.Dialect == Replicated {
		flush += fmt.Sprintf(" ON CLUSTER '%s'", r.Topo.Cluster)
	}
	_ = r.Conn.Exec(ctx, flush)
	c := &StatementCost{}
	err := r.Conn.QueryRow(ctx, fmt.Sprintf(`SELECT toFloat64(sum(ProfileEvents['OSCPUVirtualTimeMicroseconds'])/1e6), toUInt64(sum(read_bytes)), toUInt64(sum(written_bytes))
FROM %s WHERE log_comment = ? AND type = 'QueryFinish'`, r.sysTable("query_log")), lc).Scan(&c.QueryCPUSec, &c.ReadBytes, &c.WrittenBytes)
	if err != nil {
		return nil, err
	}
	err = r.Conn.QueryRow(ctx, fmt.Sprintf(`SELECT count(), toFloat64(sum(ProfileEvents['OSCPUVirtualTimeMicroseconds'])/1e6), toUInt64(sum(read_bytes)), toUInt64(sum(size_in_bytes))
FROM %s WHERE event_type = 'MutatePart' AND database = ? AND event_time >= ? AND event_time <= ?`, r.sysTable("part_log")),
		r.Topo.Database, from.Add(-time.Second), to.Add(time.Second)).Scan(&c.MutateParts, &c.MutateCPUSec, &c.MutateReadB, &c.MutateWrittenB)
	return c, err
}

// Snapshot returns create statements of every table in the database except the
// state table, used by tests to prove up→down→up→up converges.
func Snapshot(ctx context.Context, conn driver.Conn, database string) (map[string]string, error) {
	rows, err := conn.Query(ctx, `SELECT name, create_table_query FROM system.tables WHERE database = ? AND name != ? AND NOT is_temporary ORDER BY name`, database, StateTable)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[string]string{}
	for rows.Next() {
		var n, q string
		if err := rows.Scan(&n, &q); err != nil {
			return nil, err
		}
		out[n] = q
	}
	return out, rows.Err()
}

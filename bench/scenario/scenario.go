// Package scenario discovers workload scenarios (scenarios/<workload>/<variant>/) and
// prepares them on a target through the migration tool. It replaces the flat
// scenarios/<name>/init.sql layout (scenario-format delta spec); query loading is
// delegated to runner.LoadQueries, unchanged.
package scenario

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"

	"github.com/tothandras/ch-eval/bench/runner"
	"github.com/tothandras/ch-eval/bench/target"
	"github.com/tothandras/ch-eval/migrations/migrate"
)

// Config is scenarios/<workload>/<variant>/scenario.json.
type Config struct {
	Workload  string   `json:"workload"`
	Database  string   `json:"database"`
	Tables    []string `json:"tables"`    // logical tables, first is the main (seeded) table
	Generator string   `json:"generator"` // registered data generator
	Notes     string   `json:"notes,omitempty"`
}

// Scenario is one runnable variant.
type Scenario struct {
	Workload string
	Name     string
	Dir      string
	Config   Config
}

// Scope is the migration scope recorded in schema_migrations_ch.
func (s Scenario) Scope() string { return s.Workload + "/" + s.Name }

// Discover returns the scenarios of a workload, optionally filtered by name.
func Discover(scenariosDir, workload string, only []string) ([]Scenario, error) {
	dir := filepath.Join(scenariosDir, workload)
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("scenario: read %s: %w", dir, err)
	}
	want := map[string]bool{}
	for _, o := range only {
		want[o] = true
	}
	var out []Scenario
	for _, e := range entries {
		if !e.IsDir() || strings.HasPrefix(e.Name(), ".") || strings.HasPrefix(e.Name(), "_") {
			continue
		}
		if len(want) > 0 && !want[e.Name()] {
			continue
		}
		sdir := filepath.Join(dir, e.Name())
		if _, err := os.Stat(filepath.Join(sdir, "init")); err != nil {
			fmt.Fprintf(os.Stderr, "warning: skipping %s (no init/ migration set)\n", sdir)
			continue
		}
		b, err := os.ReadFile(filepath.Join(sdir, "scenario.json"))
		if err != nil {
			return nil, fmt.Errorf("scenario %s: %w", sdir, err)
		}
		var c Config
		dec := json.NewDecoder(strings.NewReader(string(b)))
		dec.DisallowUnknownFields()
		if err := dec.Decode(&c); err != nil {
			return nil, fmt.Errorf("scenario %s/scenario.json: %w", sdir, err)
		}
		if c.Workload != workload || c.Database == "" || len(c.Tables) == 0 {
			return nil, fmt.Errorf("scenario %s/scenario.json: workload must be %q and database/tables set", sdir, workload)
		}
		out = append(out, Scenario{Workload: workload, Name: e.Name(), Dir: sdir, Config: c})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	if len(want) > 0 && len(out) != len(want) {
		return out, fmt.Errorf("scenario: some of %v not found under %s", only, dir)
	}
	return out, nil
}

// Migrations loads the scenario's init migration set.
func (s Scenario) Migrations() ([]migrate.Migration, error) {
	return migrate.Load(filepath.Join(s.Dir, "init"))
}

// Queries loads queries/*.sql with the unchanged runner loader.
func (s Scenario) Queries() []runner.Query {
	qdir := filepath.Join(s.Dir, "queries")
	entries, _ := os.ReadDir(qdir)
	rs := runner.Scenario{Name: s.Name, Dir: s.Dir}
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".sql") {
			rs.QueryDirs = append(rs.QueryDirs, filepath.Join(qdir, e.Name()))
		}
	}
	sort.Strings(rs.QueryDirs)
	return runner.LoadQueries(rs)
}

// ParityQueries loads parity/*.sql: untimed checks whose digests prove two runs
// stored identical data (for example a hash over every inserted column). They are
// named "parity/<file>" in the run's value digests.
func (s Scenario) ParityQueries() []runner.Query {
	pdir := filepath.Join(s.Dir, "parity")
	entries, _ := os.ReadDir(pdir)
	rs := runner.Scenario{Name: s.Name, Dir: s.Dir}
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".sql") {
			rs.QueryDirs = append(rs.QueryDirs, filepath.Join(pdir, e.Name()))
		}
	}
	sort.Strings(rs.QueryDirs)
	qs := runner.LoadQueries(rs)
	for i := range qs {
		qs[i].Name = "parity/" + qs[i].Name
	}
	return qs
}

// PhysicalTable is the table holding rows on a shard (the main table, or its
// shard-local table on sharded topologies).
func (s Scenario) PhysicalTable(m target.Manifest) string {
	if m.Dialect == "replicated" && m.Shards > 1 {
		return s.Config.Tables[0] + "_local"
	}
	return s.Config.Tables[0]
}

// PhysicalTables are the tables holding rows on a shard, for every logical table.
func (s Scenario) PhysicalTables(m target.Manifest) []string {
	out := make([]string, 0, len(s.Config.Tables))
	for _, t := range s.Config.Tables {
		if m.Dialect == "replicated" && m.Shards > 1 {
			t += "_local"
		}
		out = append(out, t)
	}
	return out
}

// Prepare applies the scenario's migrations on the target. fresh rolls every
// applied migration back first (down files drop the objects), so a reseed starts
// from an empty schema and the down path is exercised on every fresh run.
func (s Scenario) Prepare(ctx context.Context, admin driver.Conn, m target.Manifest, fresh bool, runID string, logf func(string, ...any)) ([]migrate.Applied, error) {
	topo, err := m.Topology(s.Config.Database)
	if err != nil {
		return nil, err
	}
	migs, err := s.Migrations()
	if err != nil {
		return nil, err
	}
	r := &migrate.Runner{Conn: admin, Topo: topo, Scope: s.Scope(), Target: m.Name, RunID: runID, Logf: logf}
	if err := r.EnsureState(ctx); err != nil {
		return nil, err
	}
	var applied []migrate.Applied
	if fresh {
		down, err := r.Down(ctx, migs, 0)
		applied = append(applied, down...)
		if err != nil {
			return applied, err
		}
	}
	up, err := r.Up(ctx, migs)
	return append(applied, up...), err
}

// Verify checks the live schema of the scenario's tables against the target's
// dialect on every reachable replica: each table exists everywhere, the engine
// family matches (Shared* on cloud, Replicated* on replicated, neither on
// single; Distributed for the routed table on sharded targets), and the CREATE
// statement is identical across replicas. It returns the problems found.
func (s Scenario) Verify(ctx context.Context, admin driver.Conn, m target.Manifest) ([]string, error) {
	type want struct{ name, engine string }
	var tables []want
	for _, t := range s.Config.Tables {
		if m.Dialect == "replicated" && m.Shards > 1 {
			tables = append(tables, want{t, "Distributed"}, want{t + "_local", "Replicated"})
			continue
		}
		switch m.Dialect {
		case "cloud":
			tables = append(tables, want{t, "Shared"})
		case "replicated":
			tables = append(tables, want{t, "Replicated"})
		default:
			tables = append(tables, want{t, ""})
		}
	}
	src := "system.tables"
	if c := m.SystemCluster(); c != "" {
		src = fmt.Sprintf("clusterAllReplicas('%s', system.tables)", c)
	}
	var replicas uint64
	if c := m.SystemCluster(); c != "" {
		if err := admin.QueryRow(ctx, fmt.Sprintf("SELECT count() FROM clusterAllReplicas('%s', system.one)", c)).Scan(&replicas); err != nil {
			return nil, err
		}
	} else {
		replicas = 1
	}
	var problems []string
	for _, w := range tables {
		rows, err := admin.Query(ctx, fmt.Sprintf(
			"SELECT hostName(), engine, create_table_query FROM %s WHERE database = ? AND name = ?", src), s.Config.Database, w.name)
		if err != nil {
			return nil, err
		}
		var n uint64
		ddl := map[string]bool{}
		for rows.Next() {
			var host, engine, create string
			if err := rows.Scan(&host, &engine, &create); err != nil {
				rows.Close()
				return nil, err
			}
			n++
			ddl[create] = true
			switch {
			case w.engine == "" && (strings.HasPrefix(engine, "Replicated") || strings.HasPrefix(engine, "Shared")):
				problems = append(problems, fmt.Sprintf("%s on %s: engine %s on a single-node target", w.name, host, engine))
			case w.engine != "" && !strings.HasPrefix(engine, w.engine):
				problems = append(problems, fmt.Sprintf("%s on %s: engine %s, want %s*", w.name, host, engine, w.engine))
			}
		}
		rows.Close()
		if n != replicas {
			problems = append(problems, fmt.Sprintf("%s exists on %d of %d replicas", w.name, n, replicas))
		}
		if len(ddl) > 1 {
			problems = append(problems, fmt.Sprintf("%s: CREATE statement differs across replicas (%d variants)", w.name, len(ddl)))
		}
	}
	return problems, nil
}

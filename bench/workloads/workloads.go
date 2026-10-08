// Package workloads registers the data generators that scenarios name in
// scenario.json ("generator"): how a workload is seeded, which window its queries
// cover, and the deterministic query parameters derived from the generated data
// (design D5). Parameters depend only on the options, never on the database, so
// identical options yield identical SQL and therefore comparable digests on every
// target.
package workloads

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"

	"github.com/tothandras/ch-eval/bench/runner"
)

// Opts are the seeding options shared by all generators.
type Opts struct {
	Rows       int // target rows in the main table(s); per-workload meaning is documented by each generator
	BatchSize  int
	Seed       uint64
	Namespaces int  // openmeter
	MixedValue bool // openmeter
	TimeEnd    time.Time
	// Settings apply to every insert (the harness sets async_insert explicitly).
	Settings map[string]any
	// IngestPath is "direct" (native insert, default) or "collector" (OTLP to an
	// OpenTelemetry Collector at CollectorURL; OTel workloads only).
	IngestPath   string
	CollectorURL string
	// Profile selects a generator profile (otel-metrics: "default" or "high-card").
	Profile string
}

// Result describes what a seed wrote.
type Result struct {
	Ingest *runner.IngestResult
	// Expected rows per table, for completeness checks (collector path) and the
	// run record.
	Rows  map[string]uint64
	Notes []string
}

// Generator seeds one workload.
type Generator interface {
	// Seed writes the data into db (tables are the scenario's logical tables).
	Seed(ctx context.Context, conn driver.Conn, db string, tables []string, o Opts) (Result, error)
	// Window is the [from, to) range the queries cover.
	Window(o Opts) (time.Time, time.Time)
	// Params are the rendered SQL literals for the {name:Type} placeholders.
	Params(o Opts) (map[string]string, error)
}

var registry = map[string]Generator{}

func register(name string, g Generator) { registry[name] = g }

// Get returns a registered generator.
func Get(name string) (Generator, error) {
	g, ok := registry[name]
	if !ok {
		names := make([]string, 0, len(registry))
		for n := range registry {
			names = append(names, n)
		}
		sort.Strings(names)
		return nil, fmt.Errorf("generator %q is not registered (have %v)", name, names)
	}
	return g, nil
}

// sqlString quotes s as a ClickHouse string literal.
func sqlString(s string) string {
	out := []byte{'\''}
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '\'', '\\':
			out = append(out, '\\')
		}
		out = append(out, s[i])
	}
	return string(append(out, '\''))
}

func unix(t time.Time) string { return fmt.Sprintf("%d", t.UTC().Unix()) }

package workloads

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"

	"github.com/tothandras/ch-eval/bench/runner"
	"github.com/tothandras/ch-eval/bench/seed"
)

// openmeterEvents is the forked harness's seeder (bench/seed) and its v1 fixed
// parameter set, unchanged.
type openmeterEvents struct{}

func init() { register("openmeter-events", openmeterEvents{}) }

func (openmeterEvents) Seed(ctx context.Context, conn driver.Conn, db string, tables []string, o Opts) (Result, error) {
	if o.IngestPath != "" && o.IngestPath != "direct" {
		return Result{}, fmt.Errorf("openmeter-events: ingest path %q not supported", o.IngestPath)
	}
	cfg := seed.DefaultConfig()
	cfg.Table = db + "." + tables[0]
	cfg.Rows = o.Rows
	cfg.BatchSize = o.BatchSize
	cfg.Seed = o.Seed
	cfg.Namespaces = o.Namespaces
	cfg.MixedValueStorage = o.MixedValue
	cfg.Settings = o.Settings
	cfg.TimeEnd = o.TimeEnd
	res, err := seed.Run(ctx, conn, cfg)
	if err != nil {
		return Result{}, err
	}
	return Result{Ingest: runner.FromSeedResult(res), Rows: map[string]uint64{tables[0]: uint64(res.Rows)}}, nil
}

func (openmeterEvents) Window(o Opts) (time.Time, time.Time) {
	cfg := seed.DefaultConfig()
	return o.TimeEnd.Add(-cfg.TimeSpan), o.TimeEnd
}

// Params is the v1 fixed parameter set documented in scenarios/README.md, rendered
// as SQL literals (clickhouse-benchmark has no bound parameters).
func (openmeterEvents) Params(o Opts) (map[string]string, error) {
	cfg := seed.DefaultConfig()
	subjects := seed.Subjects(10)
	from := o.TimeEnd.Add(-cfg.TimeSpan)
	lit := make([]string, len(subjects))
	for i, s := range subjects {
		lit[i] = "'" + strings.ReplaceAll(s, "'", "''") + "'"
	}
	q := func(s string) string { return "'" + strings.ReplaceAll(s, "'", "''") + "'" }
	return map[string]string{
		"namespace": q(cfg.Namespace),
		"type":      q(cfg.Types[0]),
		// from/to are Unix seconds, not wall-clock literals: a bare
		// 'YYYY-MM-DD hh:mm:ss' passed to toDateTime() is parsed in the server or
		// session timezone, which would shift the window off the seeded UTC instants.
		"from":     strconv.FormatInt(from.UTC().Unix(), 10),
		"to":       strconv.FormatInt(o.TimeEnd.UTC().Unix(), 10),
		"subjects": "(" + strings.Join(lit, ", ") + ")",
		"group1":   q(cfg.Group1[0]),
		"group2":   q(cfg.Group2[0]),
		"model":    q("claude-haiku"),
	}, nil
}

// Command bench runs workload scenarios against a target platform.
//
// Per-query measurement is delegated to the official `clickhouse-benchmark`
// CLI (which must be on PATH; the devenv provides it). The Go binary handles
// the scenario lifecycle on a target: migrations (down+up for a fresh seed),
// seeding via bench/seed, parameter rendering, value digests, target metadata
// capture, and result assembly.
//
// Usage:
//
//	bench --target altinity/local-single --workload openmeter --scenario proposal
//
// The DSN comes from the target manifest's dsn_env variable (see .envrc.example),
// then $CLICKHOUSE_DSN, then a local tier's passwordless default; --dsn overrides.
// Each scenario writes results/<workload>/<scenario>/<platform>/<tier>/<run-id>/.
package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/spf13/cobra"

	"github.com/tothandras/ch-eval/bench/runner"
	"github.com/tothandras/ch-eval/bench/runs"
	"github.com/tothandras/ch-eval/bench/scenario"
	"github.com/tothandras/ch-eval/bench/seed"
	"github.com/tothandras/ch-eval/bench/target"
	"github.com/tothandras/ch-eval/bench/workloads"
)

func main() {
	if err := newRootCmd().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

type flags struct {
	dsn             string
	target          string
	targetsDir      string
	workload        string
	scenarios       []string
	scenariosDir    string
	resultsDir      string
	iterations      int
	concurrency     []int
	coldPaired      bool
	repeat          int
	requireClean    bool
	rows            int
	batchSize       int
	asyncInsert     bool
	waitAsyncInsert bool
	rngSeed         uint64
	namespaces      int
	namespaceSkew   float64
	mixedValue      bool
	skipSeed        bool
	allowDrift      bool
	ingestPath      string
	collectorURL    string
	manageCollector bool
	profile         string
	timeEnd         string // optional RFC3339 pin for seed TimeEnd; shared across scenarios and targets for byte-identical time windows
}

// resolveTimeEnd returns the pinned TimeEnd (if --time-end set) or the seeder
// default. Parsed once and shared by every scenario in a run so their event
// time windows are byte-identical (required for cross-scenario per-row parity).
func (f *flags) resolveTimeEnd() (time.Time, error) {
	if f.timeEnd == "" {
		return seed.DefaultConfig().TimeEnd, nil
	}
	t, err := time.Parse(time.RFC3339, f.timeEnd)
	if err != nil {
		return time.Time{}, fmt.Errorf("parse --time-end %q (want RFC3339 like 2026-06-01T00:00:00Z): %w", f.timeEnd, err)
	}
	return t.UTC().Truncate(time.Minute), nil
}

func newRootCmd() *cobra.Command {
	f := &flags{}
	cmd := &cobra.Command{
		Use:   "bench",
		Short: "Run workload scenarios on a target via migrations and clickhouse-benchmark",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return run(cmd.Context(), f)
		},
	}
	cmd.Flags().StringVar(&f.target, "target", "", "target manifest <platform>/<tier> under --targets-dir (required), e.g. altinity/local-single")
	cmd.Flags().StringVar(&f.targetsDir, "targets-dir", "targets", "directory holding target manifests")
	cmd.Flags().StringVar(&f.workload, "workload", "openmeter", "workload under --scenarios-dir")
	cmd.Flags().StringVar(&f.dsn, "dsn", "", "ClickHouse DSN (overrides the manifest's dsn_env and $CLICKHOUSE_DSN)")
	cmd.Flags().StringSliceVar(&f.scenarios, "scenario", nil, "scenario name(s); default = all of the workload")
	cmd.Flags().StringVar(&f.scenariosDir, "scenarios-dir", "scenarios", "directory holding <workload>/<scenario>/")
	cmd.Flags().StringVar(&f.resultsDir, "results-dir", "results", "root of results/<workload>/<scenario>/<target>/<run-id>/")
	cmd.Flags().IntVar(&f.iterations, "iterations", 10, "measured iterations per query")
	cmd.Flags().IntSliceVar(&f.concurrency, "concurrency", []int{1}, "concurrency level(s), e.g. 1,8,16; each measured separately (clickhouse-benchmark -c)")
	cmd.Flags().BoolVar(&f.coldPaired, "cold-paired", false, "measure each query warm AND cold (enable_filesystem_cache=0) in one run")
	cmd.Flags().IntVar(&f.repeat, "repeat", 1, "run each scenario's query set N times (reuses seeded data) to gauge run-to-run variance")
	cmd.Flags().BoolVar(&f.requireClean, "require-clean", false, "refuse to run if the harness git tree is dirty (results would be unreproducible)")
	cmd.Flags().IntVar(&f.rows, "rows", 1_000_000, "events to seed per scenario")
	cmd.Flags().IntVar(&f.batchSize, "batch-size", 10_000, "INSERT batch size")
	cmd.Flags().BoolVar(&f.asyncInsert, "async-insert", false, "seed with async_insert=1 (default: async_insert=0 set explicitly, since 26.2+ servers default it on)")
	cmd.Flags().BoolVar(&f.waitAsyncInsert, "wait-async", false, "set wait_for_async_insert=1 when async-insert is true")
	cmd.Flags().Uint64Var(&f.rngSeed, "seed", 42, "RNG seed for deterministic data generation")
	cmd.Flags().Float64Var(&f.namespaceSkew, "namespace-skew", 0, "Zipf exponent (> 1) for the namespace draw, so one tenant dominates; 0 = uniform")
	cmd.Flags().IntVar(&f.namespaces, "namespaces", 1, "number of namespaces to spread seeded rows across (multi-tenant table)")
	cmd.Flags().BoolVar(&f.mixedValue, "mixed-value", false, "emit baseline `value` in mixed JSON storage (number/string/bigint) to exercise the type-agnostic correctness fix")
	cmd.Flags().StringVar(&f.ingestPath, "ingest-path", "direct", "OTel workloads: direct (native insert with the exporter's mapping) or collector (OTLP through the OpenTelemetry Collector)")
	cmd.Flags().StringVar(&f.collectorURL, "collector-url", "http://127.0.0.1:4318", "OTLP/HTTP endpoint of the collector for --ingest-path collector")
	cmd.Flags().BoolVar(&f.manageCollector, "manage-collector", true, "start targets/collector for each scenario (pointed at its database) and stop it afterwards")
	cmd.Flags().StringVar(&f.profile, "profile", "", "generator profile (otel-metrics: high-card for >= 100k series)")
	cmd.Flags().BoolVar(&f.allowDrift, "allow-drift", false, "run even when the live target differs from its manifest (recorded in run.json)")
	cmd.Flags().BoolVar(&f.skipSeed, "skip-seed", false, "skip the fresh migrate+seed (apply pending migrations only, query existing data)")
	cmd.Flags().StringVar(&f.timeEnd, "time-end", "", "pin seed TimeEnd (RFC3339, e.g. 2026-06-01T00:00:00Z); required for cross-target parity. Default: time.Now() truncated to the minute.")
	cmd.SilenceUsage = true
	cmd.AddCommand(newCompareCmd(), newCogsCmd(), newMigrateCmd(), newProbeCmd(), newFeatureMatrixCmd(), newConsistencyCmd(), newExperimentCmd())
	return cmd
}

// checkDrift fails the run when the live target differs from its manifest,
// unless allowed (ADR-0002); allowed drift is still recorded in the run.
func checkDrift(md target.Metadata, allow bool) error {
	for _, d := range md.Drift {
		fmt.Fprintf(os.Stderr, "warning: target drift: %s\n", d)
	}
	if len(md.Drift) > 0 && !allow {
		return fmt.Errorf("target %s drifted from its manifest (%d difference(s)); fix the manifest or pass --allow-drift", md.Manifest.Name, len(md.Drift))
	}
	return nil
}

// connect opens a connection with database overridden (empty keeps the DSN's).
func connect(ctx context.Context, opts clickhouse.Options, database string) (driver.Conn, error) {
	if database != "" {
		opts.Auth.Database = database
	}
	conn, err := clickhouse.Open(&opts)
	if err != nil {
		return nil, fmt.Errorf("connect: %w", err)
	}
	if err := conn.Ping(ctx); err != nil {
		conn.Close()
		return nil, fmt.Errorf("ping %v: %w", opts.Addr, err)
	}
	return conn, nil
}

// resolveTarget loads the manifest and parses its DSN.
func resolveTarget(f *flags) (target.Manifest, *clickhouse.Options, string, error) {
	if f.target == "" {
		return target.Manifest{}, nil, "", errors.New("--target is required (for example altinity/local-single); see targets/README.md")
	}
	m, err := target.Load(f.targetsDir, f.target)
	if err != nil {
		return m, nil, "", err
	}
	dsn, src := f.dsn, "--dsn"
	if dsn == "" {
		if dsn, src, err = m.DSN(); err != nil {
			return m, nil, "", err
		}
	}
	opts, err := clickhouse.ParseDSN(dsn)
	if err != nil {
		return m, nil, "", fmt.Errorf("parse DSN from %s: %w", src, err)
	}
	return m, opts, src, nil
}

func run(ctx context.Context, f *flags) error {
	if err := runner.EnsureBinary(); err != nil {
		return err
	}
	if _, err := f.resolveTimeEnd(); err != nil { // fail fast on a bad --time-end
		return err
	}

	commit := runner.HarnessCommit()
	if strings.HasSuffix(commit, "-dirty") {
		if f.requireClean {
			return fmt.Errorf("harness tree is dirty (%s); commit changes or drop --require-clean", commit)
		}
		fmt.Fprintf(os.Stderr, "warning: harness tree is dirty (%s); results will be tagged -dirty and are not reproducible from the commit alone\n", commit)
	}

	m, opts, dsnSource, err := resolveTarget(f)
	if err != nil {
		return err
	}
	admin, err := connect(ctx, *opts, "")
	if err != nil {
		return err
	}
	defer admin.Close()

	host, port, err := hostPort(opts)
	if err != nil {
		return err
	}

	scenarios, err := scenario.Discover(f.scenariosDir, f.workload, f.scenarios)
	if err != nil {
		return err
	}
	if len(scenarios) == 0 {
		return fmt.Errorf("no scenarios found in %s/%s", f.scenariosDir, f.workload)
	}

	fingerprint, err := runner.Fingerprint(ctx, admin)
	if err != nil {
		return fmt.Errorf("fingerprint: %w", err)
	}
	md := target.Capture(ctx, admin, m, dsnSource, commit)
	if err := checkDrift(md, f.allowDrift); err != nil {
		return err
	}
	if m.IsLocal() {
		fmt.Printf("target %s (%s, %s): local tier, results are smoke/parity only, never performance\n", m.Name, md.Version, m.Dialect)
	}

	// Per-invocation sweep id correlates query_log rows to this run's queries.
	sweepID := time.Now().UTC().Format("20060102T150405Z")
	var cpuProbe runner.CPUProbe
	switch {
	case !runner.QueryLogAvailable(ctx, admin):
		fmt.Fprintln(os.Stderr, "warning: system.query_log unavailable on target; per-query CPU will be null")
	case m.SystemCluster() != "" && m.Replicas*max(m.Shards, 1) > 1:
		cpuProbe = runner.NewClusterCPUProbe(admin, m.SystemCluster())
	default:
		cpuProbe = runner.NewCPUProbe(admin)
	}

	for _, sc := range scenarios {
		fmt.Printf("\n== %s/%s on %s ==\n", sc.Workload, sc.Name, m.Name)
		startedAt := time.Now()
		runID := runs.NewRunID(startedAt)
		logf := func(format string, a ...any) { fmt.Printf("   "+format+"\n", a...) }

		// A fresh seed rolls the scenario's migrations down and up again, so the
		// down path is exercised on every run and no stale schema shadows the
		// current one. --skip-seed only applies pending migrations.
		fmt.Printf(" migrations (fresh=%v) ...\n", !f.skipSeed)
		applied, err := sc.Prepare(ctx, admin, m, !f.skipSeed, runID, logf)
		if err != nil {
			return fmt.Errorf("%s: migrations: %w", sc.Scope(), err)
		}

		conn, err := connect(ctx, *opts, sc.Config.Database)
		if err != nil {
			return err
		}
		table := sc.Config.Tables[0]
		gen, err := workloads.Get(sc.Config.Generator)
		if err != nil {
			conn.Close()
			return fmt.Errorf("%s: %w", sc.Scope(), err)
		}
		wopts, err := f.workloadOpts(m)
		if err != nil {
			conn.Close()
			return err
		}
		params, err := gen.Params(wopts)
		if err != nil {
			conn.Close()
			return fmt.Errorf("%s: params: %w", sc.Scope(), err)
		}

		var ingest *runner.IngestResult
		var seedBlock *runs.SeedBlock
		if !f.skipSeed {
			ingest, seedBlock, err = doSeed(ctx, conn, admin, m, sc, gen, wopts, host, port, f)
			if err != nil {
				conn.Close()
				return err
			}
		}

		queries := sc.Queries()
		fmt.Printf(" queries: %d (iterations=%d, concurrency=%v, repeat=%d, cold-paired=%v)\n",
			len(queries), f.iterations, f.concurrency, f.repeat, f.coldPaired)

		// Cache states to measure per query: warm always; cold too when paired.
		cacheStates := []bool{false}
		if f.coldPaired {
			cacheStates = append(cacheStates, true)
		}

		// Capture the index-pruning signal once (it's a property of the table +
		// index, independent of repeat/concurrency/cache). Attached to the
		// matching query's results below.
		pruning := captureIndexPruning(ctx, conn, queries, table, params)

		// Capture each query's normalized result digest once (a property of
		// (scenario, query, params), independent of the timed variants). Stored in
		// the run JSON so `bench compare` can check value parity across scenarios
		// and targets with no extra DB round-trip.
		from, to := gen.Window(wopts)
		valueParity := captureValueDigests(ctx, conn, sc.Name, append(append([]runner.Query{}, queries...), sc.ParityQueries()...), params, from, to, f.timeEnd)
		conn.Close()

		var results []runner.BenchResult
		for rep := 0; rep < f.repeat; rep++ {
			for _, conc := range f.concurrency {
				for _, cold := range cacheStates {
					benchOpts := runner.BenchOpts{
						Host:        host,
						Port:        port,
						Secure:      opts.TLS != nil,
						Database:    sc.Config.Database,
						User:        opts.Auth.Username,
						Password:    opts.Auth.Password,
						Iterations:  f.iterations,
						Concurrency: conc,
						ColdCache:   cold,
						Params:      params,
						Scenario:    sc.Name,
						SweepID:     sweepID,
						CPUProbe:    cpuProbe,
					}
					for _, q := range queries {
						label := q.Name
						if len(f.concurrency) > 1 || f.coldPaired || f.repeat > 1 {
							label = fmt.Sprintf("%s [c=%d %s r%d]", q.Name, conc, cacheTag(cold), rep+1)
						}
						fmt.Printf("  %-52s", label)
						res := runner.Bench(ctx, benchOpts, q)
						if res.Error != "" {
							fmt.Printf("ERR: %s\n", res.Error)
						} else {
							cpu := "cpu=n/a"
							if res.CPUp50Us != nil {
								cpu = fmt.Sprintf("cpu_p50=%.1fms", *res.CPUp50Us/1000)
							}
							fmt.Printf("p50=%.1fms p95=%.1fms %s QPS=%.1f\n",
								res.P50Sec*1000, res.P95Sec*1000, cpu, res.QPS)
						}
						if pruning != nil && res.Name == lookupByIDQuery {
							res.IndexPruning = pruning
						}
						results = append(results, res)
					}
				}
			}
		}

		rec := runs.Run{
			Run: runner.Run{
				Scenario:           sc.Name,
				HarnessCommit:      commit,
				SweepID:            sweepID,
				StartedAt:          startedAt,
				FinishedAt:         time.Now(),
				ClusterFingerprint: fingerprint,
				Concurrency:        f.concurrency,
				Repeat:             f.repeat,
				Ingest:             ingest,
				Queries:            results,
				ValueParity:        valueParity,
			},
			RunID:      runID,
			Workload:   sc.Workload,
			Target:     runs.Block(m, md),
			Seed:       seedBlock,
			Migrations: applied,
			SmokeOnly:  m.IsLocal(),
		}
		report := runner.RenderReport(rec.Run, queries)
		if rec.SmokeOnly {
			report = fmt.Sprintf("> Target `%s` is a local tier: this run backs parity and correctness only, never performance.\n\n", m.Name) + report
		}
		dir, err := runs.Write(f.resultsDir, rec, report, md)
		if err != nil {
			return fmt.Errorf("write results: %w", err)
		}
		fmt.Printf(" wrote: %s/{run.json,report.md,metadata.json}\n", dir)
	}

	return nil
}

// lookupByIDQuery is the query name whose bloom_filter pruning we capture via
// EXPLAIN. Only scenarios that ship this query (with a bloom on id) get a
// pruning signal; others get nil.
const lookupByIDQuery = "lookup_by_id"

func cacheTag(cold bool) string {
	if cold {
		return "cold"
	}
	return "warm"
}

// captureIndexPruning runs `EXPLAIN indexes=1` for a literal-id lookup on table,
// once with skip indexes disabled and once enabled, and returns the granule/part
// counts each scans. Returns nil (no error surfaced) when the scenario has no
// lookup_by_id query, the table is empty, or EXPLAIN fails — the signal is
// best-effort enrichment, never a reason to abort the run.
func captureIndexPruning(ctx context.Context, conn driver.Conn, queries []runner.Query, table string, params map[string]string) *runner.IndexPruning {
	hasLookup := false
	for _, q := range queries {
		if q.Name == lookupByIDQuery {
			hasLookup = true
			break
		}
	}
	if !hasLookup {
		return nil
	}

	ns := strings.Trim(params["namespace"], "'")
	var literalID string
	idQ := fmt.Sprintf("SELECT id FROM %s WHERE namespace = ? ORDER BY namespace, type, subject, time LIMIT 1 OFFSET 1000", table)
	if err := conn.QueryRow(ctx, idQ, ns).Scan(&literalID); err != nil {
		fmt.Fprintf(os.Stderr, "warning: index-pruning capture skipped (resolve id: %v)\n", err)
		return nil
	}

	explain := func(useSkip int) (granules, parts int, err error) {
		q := fmt.Sprintf(
			"EXPLAIN indexes = 1 SELECT count() FROM %s WHERE namespace = ? AND id = ? SETTINGS use_skip_indexes = %d",
			table, useSkip)
		rows, err := conn.Query(ctx, q, ns, literalID)
		if err != nil {
			return 0, 0, err
		}
		defer rows.Close()
		// EXPLAIN indexes=1 prints a Granules:/Parts: line per index (MinMax,
		// Partition, PrimaryKey, then each Skip index), in application order; each
		// numerator is the running survivor count, so the LAST line is the
		// granules/parts actually scanned. Keeping the last match is therefore
		// correct. Assumes ≤1 *pruning* skip index (here only the id bloom prunes
		// an id-lookup; the stored_at minmax is pass-through), so ordering can't
		// pick the wrong index. Verified against EXPLAIN: bloom prunes 26→1.
		for rows.Next() {
			var line string
			if err := rows.Scan(&line); err != nil {
				return 0, 0, err
			}
			if g, ok := parseExplainCount(line, "Granules:"); ok {
				granules = g
			}
			if p, ok := parseExplainCount(line, "Parts:"); ok {
				parts = p
			}
		}
		return granules, parts, rows.Err()
	}

	gWithout, pWithout, err := explain(0)
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: index-pruning capture skipped (explain no-index: %v)\n", err)
		return nil
	}
	gWith, pWith, err := explain(1)
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: index-pruning capture skipped (explain index: %v)\n", err)
		return nil
	}
	return &runner.IndexPruning{
		LiteralID:       literalID,
		GranulesWithout: gWithout,
		GranulesWith:    gWith,
		PartsWithout:    pWithout,
		PartsWith:       pWith,
	}
}

// captureValueDigests computes a normalized result digest for every query in
// the scenario, once, using the same params the timed run binds. The digests +
// the resolved time window go into the run JSON so `bench compare` can verify
// value parity across scenarios (same events → same meter output) by diffing
// digests, gated on an identical window. Best-effort: a per-query failure is
// recorded as a digest error, never aborts the run; returns nil only if params
// can't be built at all.
func captureValueDigests(ctx context.Context, conn driver.Conn, scenarioName string, queries []runner.Query, params map[string]string, from, to time.Time, timeEnd string) *runner.ValueParity {
	const utcFmt = "2006-01-02 15:04:05"
	vp := &runner.ValueParity{
		From:    from.UTC().Format(utcFmt),
		To:      to.UTC().Format(utcFmt),
		TimeEnd: timeEnd,
		Digests: map[string]runner.QueryDigest{},
	}
	for _, q := range queries {
		sql, err := runner.RenderParams(q.SQL, params)
		if err != nil {
			vp.Digests[q.Name] = runner.QueryDigest{Error: err.Error()}
			continue
		}
		d, err := runner.DigestResult(ctx, conn, strings.TrimRight(strings.TrimSpace(sql), ";"))
		if err != nil {
			fmt.Fprintf(os.Stderr, "warning: value digest for %s/%s skipped: %v\n", scenarioName, q.Name, err)
		}
		vp.Digests[q.Name] = d
	}
	return vp
}

// parseExplainCount pulls the surviving count from an EXPLAIN line of the form
// "<label> <surviving>/<total>" (e.g. "Granules: 12/1224" → 12). The label may
// be indented and appear mid-line.
func parseExplainCount(line, label string) (int, bool) {
	idx := strings.Index(line, label)
	if idx < 0 {
		return 0, false
	}
	rest := strings.TrimSpace(line[idx+len(label):])
	slash := strings.Index(rest, "/")
	if slash < 0 {
		return 0, false
	}
	n, err := strconv.Atoi(strings.TrimSpace(rest[:slash]))
	if err != nil {
		return 0, false
	}
	return n, true
}

// hostPort extracts the first host:port from the parsed DSN. clickhouse-benchmark
// takes --host/--port (singular) and we don't try to span multiple shards from a
// single bench process.
func hostPort(opts *clickhouse.Options) (string, int, error) {
	if len(opts.Addr) == 0 {
		return "", 0, errors.New("DSN has no addresses")
	}
	addr := opts.Addr[0]
	h, p, err := net.SplitHostPort(addr)
	if err != nil {
		return "", 0, fmt.Errorf("split host:port from %q: %w", addr, err)
	}
	port, err := strconv.Atoi(p)
	if err != nil {
		return "", 0, fmt.Errorf("parse port from %q: %w", p, err)
	}
	return h, port, nil
}

// workloadOpts builds the generator options. Seeding sets async_insert
// explicitly: 26.2+ servers default async_insert=1, and the harness must not
// inherit a server default that differs across targets. On a sharded target the
// insert is synchronous (distributed_foreground_insert=1) so the data is on the
// shards when the timed queries start.
func (f *flags) workloadOpts(m target.Manifest) (workloads.Opts, error) {
	te, err := f.resolveTimeEnd()
	if err != nil {
		return workloads.Opts{}, err
	}
	settings := map[string]any{"async_insert": 0}
	if f.asyncInsert {
		settings["async_insert"] = 1
		settings["wait_for_async_insert"] = map[bool]int{true: 1, false: 0}[f.waitAsyncInsert]
	}
	if m.Dialect == "replicated" && m.Shards > 1 {
		settings["distributed_foreground_insert"] = 1
	}
	return workloads.Opts{
		Rows: f.rows, BatchSize: f.batchSize, Seed: f.rngSeed, Namespaces: max(f.namespaces, 1), NamespaceSkew: f.namespaceSkew, MixedValue: f.mixedValue,
		TimeEnd: te, Settings: settings, IngestPath: f.ingestPath, CollectorURL: f.collectorURL, Profile: f.profile,
	}, nil
}

// doSeed loads the scenario through its registered generator. The collector path
// starts the pinned collector for the scenario's database, waits until every
// table holds the expected rows (the exporter inserts asynchronously and may
// retry), and stops it. On a replicated target the replicas are synced before
// measuring so digests do not depend on which replica answers.
func doSeed(ctx context.Context, conn, admin driver.Conn, m target.Manifest, sc scenario.Scenario, gen workloads.Generator, o workloads.Opts, host string, port int, f *flags) (*runner.IngestResult, *runs.SeedBlock, error) {
	collector := o.IngestPath == "collector"
	if collector && f.manageCollector {
		addr := fmt.Sprintf("%s:%d", host, port)
		if out, err := exec.CommandContext(ctx, "targets/collector/up", addr, sc.Config.Database).CombinedOutput(); err != nil {
			return nil, nil, fmt.Errorf("collector up: %v: %s", err, out)
		}
		defer func() {
			if out, err := exec.Command("targets/collector/down").CombinedOutput(); err != nil {
				fmt.Fprintf(os.Stderr, "warning: collector down: %v: %s\n", err, out)
			}
		}()
	}
	fmt.Printf(" seed: %s, %d rows into %s (path=%s, batch=%d, settings=%v) ...\n",
		sc.Config.Generator, o.Rows, sc.Config.Database, firstNonEmpty(o.IngestPath, "direct"), o.BatchSize, o.Settings)
	start := time.Now()
	res, err := gen.Seed(ctx, conn, sc.Config.Database, sc.Config.Tables, o)
	if err != nil {
		return nil, nil, err
	}
	if collector {
		// The send loop returns once the collector accepted the data; the rows are
		// in ClickHouse only after the exporter's batches and async inserts land.
		if err := waitRows(ctx, conn, sc.Config.Database, res.Rows, 10*time.Minute); err != nil {
			return nil, nil, err
		}
		res.Ingest.DurationSeconds = time.Since(start).Seconds()
		res.Ingest.EventsPerSecond = float64(res.Ingest.Rows) / res.Ingest.DurationSeconds
	}
	fmt.Printf(" seed done: %d rows in %.1fs (%.0f rows/sec) %v\n", res.Ingest.Rows, res.Ingest.DurationSeconds, res.Ingest.EventsPerSecond, res.Notes)

	if m.Dialect == "replicated" && m.Replicas > 1 {
		for _, t := range sc.PhysicalTables(m) {
			q := fmt.Sprintf("SYSTEM SYNC REPLICA ON CLUSTER '%s' %s.%s", m.Cluster, sc.Config.Database, t)
			if err := admin.Exec(ctx, q); err != nil {
				return nil, nil, fmt.Errorf("sync replicas after seed: %w", err)
			}
		}
	}
	var cache *runs.CacheSnapshot
	src := "system.filesystem_cache"
	if c := m.SystemCluster(); c != "" {
		src = fmt.Sprintf("clusterAllReplicas('%s', system.filesystem_cache)", c)
	}
	var segs, bytes uint64
	if err := admin.QueryRow(ctx, "SELECT count(), sum(size) FROM "+src).Scan(&segs, &bytes); err == nil && segs > 0 {
		cache = &runs.CacheSnapshot{Segments: segs, Bytes: bytes}
	}
	block := &runs.SeedBlock{
		Generator: sc.Config.Generator, Rows: res.Ingest.Rows, RNGSeed: o.Seed, Namespaces: o.Namespaces, NamespaceSkew: o.NamespaceSkew,
		TimeEnd: o.TimeEnd.UTC().Format(time.RFC3339), Table: sc.Config.Database + "." + sc.Config.Tables[0], Settings: o.Settings,
		IngestPath: firstNonEmpty(o.IngestPath, "direct"), Profile: o.Profile, TableRows: res.Rows, Notes: res.Notes,
		CacheAfterSeed: cache,
	}
	return res.Ingest, block, nil
}

// waitRows polls until each table holds exactly the expected rows. More rows than
// expected means the ingest path duplicated data, which breaks path parity.
func waitRows(ctx context.Context, conn driver.Conn, db string, want map[string]uint64, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for {
		done := true
		for t, n := range want {
			var got uint64
			if err := conn.QueryRow(ctx, fmt.Sprintf("SELECT count() FROM %s.%s", db, t)).Scan(&got); err != nil {
				return err
			}
			if got > n {
				return fmt.Errorf("%s.%s holds %d rows, expected %d: the ingest path duplicated data", db, t, got, n)
			}
			done = done && got == n
		}
		if done {
			return nil
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("timed out waiting for %v rows in %s", want, db)
		}
		time.Sleep(time.Second)
	}
}

func firstNonEmpty(a, b string) string {
	if a != "" {
		return a
	}
	return b
}

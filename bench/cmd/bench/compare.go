package main

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/tothandras/ch-eval/bench/runner"
	"github.com/tothandras/ch-eval/bench/runs"
)

// newCompareCmd builds `bench compare`. Two modes:
//
//	--scenario S --targets a,b[,c]   cross-target: the same scenario on several
//	                                 targets; the first target is the baseline.
//	--target T --scenarios A,B       cross-scenario on one target (the forked
//	                                 harness's original table-design compare).
//
// It reads the latest valid run per side from results/<workload>/<scenario>/<target>/,
// prints per-query p50 and CPU deltas against the baseline, and diffs value
// digests. A digest mismatch, or (cross-target) runs whose seeded inputs differ,
// exits non-zero: no performance statement may be made for that scenario until
// parity holds (cross-target-comparison spec). Any local-tier side, or sides on
// different tiers, label the output "smoke, not performance".
func newCompareCmd() *cobra.Command {
	var resultsDir, workload, scenarioName, targetName string
	var targets, scenarios, runDirs []string
	cmd := &cobra.Command{
		Use:   "compare",
		Short: "Compare the latest runs across targets (or scenarios) with a value-parity gate",
		Args:  cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			var sides []side
			crossTarget := len(targets) > 0
			switch {
			case len(runDirs) >= 2:
				// explicit runs, for example the same scenario and target loaded
				// through two ingest paths (path parity)
				crossTarget = true
				for _, d := range runDirs {
					r, path, err := runs.Read(d)
					if err != nil {
						return err
					}
					label := r.Target.Name + " " + r.RunID
					if r.Seed != nil {
						label += " (" + r.Seed.IngestPath + ")"
					}
					sides = append(sides, side{label: label, run: r, path: path})
				}
			case crossTarget && len(scenarios) == 0 && scenarioName != "" && len(targets) >= 2:
				for _, t := range targets {
					r, path, err := runs.Latest(resultsDir, workload, scenarioName, t)
					if err != nil {
						return err
					}
					sides = append(sides, side{label: t, run: r, path: path})
				}
			case !crossTarget && targetName != "" && len(scenarios) >= 2:
				for _, sc := range scenarios {
					r, path, err := runs.Latest(resultsDir, workload, sc, targetName)
					if err != nil {
						return err
					}
					sides = append(sides, side{label: sc, run: r, path: path})
				}
			default:
				return fmt.Errorf("use --scenario S --targets a,b[,...], --target T --scenarios A,B[,...], or --runs <run dir>,<run dir>[,...]")
			}
			return compareSides(os.Stdout, sides, crossTarget)
		},
	}
	cmd.Flags().StringVar(&resultsDir, "results-dir", "results", "results root")
	cmd.Flags().StringVar(&workload, "workload", "openmeter", "workload")
	cmd.Flags().StringVar(&scenarioName, "scenario", "", "scenario to compare across --targets")
	cmd.Flags().StringSliceVar(&targets, "targets", nil, "targets <platform>/<tier>, baseline first")
	cmd.Flags().StringVar(&targetName, "target", "", "target for a cross-scenario compare")
	cmd.Flags().StringSliceVar(&scenarios, "scenarios", nil, "scenarios to compare on --target, baseline first")
	cmd.Flags().StringSliceVar(&runDirs, "runs", nil, "explicit run directories (or run.json files), baseline first")
	cmd.SilenceUsage = true
	return cmd
}

type side struct {
	label string
	run   runs.Run
	path  string
}

// smokeOnly reports whether the comparison may only be read as a smoke check.
func smokeOnly(sides []side) bool {
	tiers := map[string]bool{}
	for _, s := range sides {
		if s.run.SmokeOnly || s.run.Target.Tier == "" {
			return true
		}
		tiers[s.run.Target.Tier] = true
	}
	return len(tiers) > 1
}

func compareSides(w io.Writer, sides []side, crossTarget bool) error {
	base := sides[0]
	fmt.Fprintf(w, "# Compare %s/%s\n\n", base.run.Workload, base.run.Scenario)
	if smokeOnly(sides) {
		fmt.Fprintf(w, "> **Smoke, not performance.** At least one side is a local tier, or the sides are on different tiers. Latency and CPU deltas below are not evidence about the platforms.\n\n")
	}
	fmt.Fprintln(w, "| Side | Target | Version | Topology | Run | Commit |")
	fmt.Fprintln(w, "| --- | --- | --- | --- | --- | --- |")
	for _, s := range sides {
		t := s.run.Target
		fmt.Fprintf(w, "| %s | %s (%s) | %s | %s %dx%d %s | %s | %s |\n", s.label, t.Name, t.Tier, t.Version,
			t.Dialect, max(t.Shards, 1), max(t.Replicas, 1), t.Storage, s.run.RunID, s.run.HarnessCommit)
	}

	var failures []string
	for _, cand := range sides[1:] {
		fmt.Fprintf(w, "\n## %s → %s\n\n", base.label, cand.label)
		b, c := indexQueries(base.run.Queries), indexQueries(cand.run.Queries)
		fmt.Fprintln(w, "| Query | p50 Δ | CPU Δ | base p50 ms | cand p50 ms |")
		fmt.Fprintln(w, "| --- | ---: | ---: | ---: | ---: |")
		for _, n := range unionNames(b, c) {
			bq, bok := b[n]
			cq, cok := c[n]
			switch {
			case bok && cok:
				fmt.Fprintf(w, "| %s | %s | %s | %.1f | %.1f |\n", n, pctDelta(bq.P50Sec, cq.P50Sec), cpuDelta(bq, cq), bq.P50Sec*1000, cq.P50Sec*1000)
			case bok:
				fmt.Fprintf(w, "| %s | _(base only)_ | | %.1f | — |\n", n, bq.P50Sec*1000)
			default:
				fmt.Fprintf(w, "| %s | _(candidate only)_ | | — | %.1f |\n", n, cq.P50Sec*1000)
			}
		}
		fmt.Fprintf(w, "\nIngest: base %s, candidate %s (%s)\n", ingestStr(base.run.Ingest), ingestStr(cand.run.Ingest), ingestDelta(base.run.Ingest, cand.run.Ingest))

		if crossTarget {
			if why := inputsDiffer(base.run, cand.run); why != "" {
				fmt.Fprintf(w, "\n## Value parity\n\nnot established: %s. Re-run both with identical seed flags and the same `--time-end`.\n", why)
				failures = append(failures, fmt.Sprintf("%s vs %s: parity not established (%s)", base.label, cand.label, why))
				continue
			}
		}
		if err := compareValues(w, base.run.Run, cand.run.Run, base.label, cand.label); err != nil {
			failures = append(failures, err.Error())
		}
	}
	if len(failures) > 0 {
		return fmt.Errorf("%s", strings.Join(failures, "; "))
	}
	return nil
}

// inputsDiffer explains why two runs did not see the same events, or "".
func inputsDiffer(a, b runs.Run) string {
	if a.Seed == nil || b.Seed == nil {
		return "a run has no seed block (seeded with --skip-seed or by an older harness)"
	}
	switch {
	case a.Seed.Rows != b.Seed.Rows:
		return fmt.Sprintf("rows %d vs %d", a.Seed.Rows, b.Seed.Rows)
	case a.Seed.RNGSeed != b.Seed.RNGSeed:
		return fmt.Sprintf("rng seed %d vs %d", a.Seed.RNGSeed, b.Seed.RNGSeed)
	case a.Seed.Namespaces != b.Seed.Namespaces:
		return fmt.Sprintf("namespaces %d vs %d", a.Seed.Namespaces, b.Seed.Namespaces)
	case a.Seed.TimeEnd != b.Seed.TimeEnd:
		return fmt.Sprintf("time_end %s vs %s", a.Seed.TimeEnd, b.Seed.TimeEnd)
	case a.Seed.Generator != b.Seed.Generator:
		return fmt.Sprintf("generator %s vs %s", a.Seed.Generator, b.Seed.Generator)
	}
	return ""
}

// compareValues diffs the per-query result digests recorded by each run, proving
// the two table designs compute the same meter values. It only compares when
// both runs covered an IDENTICAL seeded window — otherwise the events differ
// (e.g. an unpinned per-scenario time.Now() TimeEnd) and a digest mismatch would
// be a false alarm, not a design difference. Returns a non-nil error if any
// query's values genuinely differ, so the command exits non-zero (CI gate).
func compareValues(w io.Writer, base, cand runner.Run, baseName, candName string) error {
	bp, cp := base.ValueParity, cand.ValueParity
	if bp == nil || cp == nil {
		fmt.Fprintln(w, "\n## Value parity\n\nskipped — one or both runs predate value-digest capture (re-run the bench to populate it).")
		return nil
	}
	if bp.From != cp.From || bp.To != cp.To {
		fmt.Fprintf(w, "\n## Value parity\n\nskipped — runs cover different windows (%s: %s→%s, %s: %s→%s). "+
			"Re-run both with the same `--time-end` so they seed an identical event stream, then compare.\n",
			baseName, bp.From, bp.To, candName, cp.From, cp.To)
		return nil
	}

	names := make([]string, 0, len(bp.Digests))
	for n := range bp.Digests {
		names = append(names, n)
	}
	sort.Strings(names)

	fmt.Fprintf(w, "\n## Value parity (window %s → %s)\n\n", bp.From, bp.To)
	fmt.Fprintln(w, "| Query | result | rows |")
	fmt.Fprintln(w, "| --- | --- | ---: |")
	var match, differ, missing int
	var differing []string
	for _, n := range names {
		b := bp.Digests[n]
		c, ok := cp.Digests[n]
		switch {
		case !ok:
			fmt.Fprintf(w, "| %s | _(base only)_ | %d |\n", n, b.Rows)
			missing++
		case b.Error != "" || c.Error != "":
			fmt.Fprintf(w, "| %s | ERROR | — |\n", n)
			differ++
			differing = append(differing, n)
		case b.Hash == c.Hash:
			fmt.Fprintf(w, "| %s | MATCH | %d |\n", n, b.Rows)
			match++
		default:
			fmt.Fprintf(w, "| %s | **DIFFERS** | %d vs %d |\n", n, b.Rows, c.Rows)
			differ++
			differing = append(differing, n)
		}
	}
	// candidate-only queries (present in cand, absent in base)
	for n := range cp.Digests {
		if _, ok := bp.Digests[n]; !ok {
			fmt.Fprintf(w, "| %s | _(candidate only)_ | %d |\n", n, cp.Digests[n].Rows)
			missing++
		}
	}
	fmt.Fprintf(w, "\nMATCH=%d DIFFERS=%d (one side only=%d)\n", match, differ, missing)
	if differ > 0 {
		return fmt.Errorf("value parity: %d quer(y/ies) differ between %s and %s: %s",
			differ, baseName, candName, strings.Join(differing, ", "))
	}
	return nil
}

// indexQueries keeps, per query name, the warm result at the lowest concurrency
// (the canonical apples-to-apples measurement). A repeated name from a sweep
// keeps the first warm/lowest-c hit.
func indexQueries(qs []runner.BenchResult) map[string]runner.BenchResult {
	out := map[string]runner.BenchResult{}
	for _, q := range qs {
		if q.Error != "" {
			continue
		}
		if q.CacheState == "cold" {
			continue
		}
		if existing, ok := out[q.Name]; ok && existing.Concurrency <= q.Concurrency {
			continue
		}
		out[q.Name] = q
	}
	return out
}

func unionNames(a, b map[string]runner.BenchResult) []string {
	set := map[string]bool{}
	for n := range a {
		set[n] = true
	}
	for n := range b {
		set[n] = true
	}
	names := make([]string, 0, len(set))
	for n := range set {
		names = append(names, n)
	}
	sort.Strings(names)
	return names
}

func pctDelta(base, cand float64) string {
	if base == 0 {
		return "—"
	}
	return fmt.Sprintf("%+.0f%%", (cand-base)/base*100)
}

func cpuDelta(b, c runner.BenchResult) string {
	if b.CPUp50Us == nil || c.CPUp50Us == nil {
		return "n/a"
	}
	return pctDelta(*b.CPUp50Us, *c.CPUp50Us)
}

func ingestStr(i *runner.IngestResult) string {
	if i == nil || i.EventsPerSecond == 0 {
		return "n/a"
	}
	return fmt.Sprintf("%.0f events/s", i.EventsPerSecond)
}

func ingestDelta(b, c *runner.IngestResult) string {
	if b == nil || c == nil || b.EventsPerSecond == 0 {
		return "n/a"
	}
	return "Δ " + pctDelta(b.EventsPerSecond, c.EventsPerSecond)
}

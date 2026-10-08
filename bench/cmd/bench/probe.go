package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/tothandras/ch-eval/bench/probe"
	"github.com/tothandras/ch-eval/bench/runner"
	"github.com/tothandras/ch-eval/bench/runs"
	"github.com/tothandras/ch-eval/bench/target"
)

const probeDB = "ch_eval_probe"

// probeRun is results/probes/<platform>/<tier>/<run-id> (tier = the manifest file name)/probes.json.
type probeRun struct {
	RunID         string         `json:"run_id"`
	Target        string         `json:"target"`
	Platform      string         `json:"platform"`
	Tier          string         `json:"tier"`
	Version       string         `json:"version"`
	HarnessCommit string         `json:"harness_commit"`
	StartedAt     time.Time      `json:"started_at"`
	Results       []probe.Result `json:"results"`
}

func newProbeCmd() *cobra.Command {
	f := &flags{}
	var probesDir string
	var only []string
	cmd := &cobra.Command{
		Use:   "probe",
		Short: "Run the feature probes against a target and record the outcomes",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runProbes(cmd.Context(), f, probesDir, only)
		},
	}
	cmd.Flags().StringVar(&f.target, "target", "", "target manifest <platform>/<tier> (required)")
	cmd.Flags().StringVar(&f.targetsDir, "targets-dir", "targets", "directory holding target manifests")
	cmd.Flags().StringVar(&f.dsn, "dsn", "", "ClickHouse DSN override")
	cmd.Flags().StringVar(&f.resultsDir, "results-dir", "results", "results root; probes go to <root>/probes/<platform>/<tier>/<run-id>/")
	cmd.Flags().StringVar(&probesDir, "probes-dir", "probes", "directory holding *.sql probes")
	cmd.Flags().StringSliceVar(&only, "only", nil, "run only probes whose path contains one of these substrings")
	cmd.SilenceUsage = true
	return cmd
}

func runProbes(ctx context.Context, f *flags, probesDir string, only []string) error {
	m, opts, dsnSource, err := resolveTarget(f)
	if err != nil {
		return err
	}
	conn, err := connect(ctx, *opts, "")
	if err != nil {
		return err
	}
	defer conn.Close()
	commit := runner.HarnessCommit()
	md := target.Capture(ctx, conn, m, dsnSource, commit)
	probes, err := probe.Load(probesDir)
	if err != nil {
		return err
	}
	// The scratch database exists on every node of a self-managed cluster so that
	// ON CLUSTER probes can create tables in it.
	onCluster := ""
	if c := m.SystemCluster(); c != "" && m.Dialect == "replicated" {
		onCluster = fmt.Sprintf(" ON CLUSTER '%s'", c)
	}
	if err := conn.Exec(ctx, "DROP DATABASE IF EXISTS "+probeDB+onCluster+" SYNC"); err != nil {
		return fmt.Errorf("probe scratch database: %w", err)
	}
	if err := conn.Exec(ctx, "CREATE DATABASE "+probeDB+onCluster); err != nil {
		return fmt.Errorf("probe scratch database: %w", err)
	}
	defer func() {
		_ = conn.Exec(context.WithoutCancel(ctx), "DROP DATABASE IF EXISTS "+probeDB+onCluster+" SYNC")
	}()

	env := probe.Env{Conn: conn, Options: opts, Manifest: m, DB: probeDB, Nodes: max(len(md.Replicas), 1)}
	run := probeRun{RunID: runs.NewRunID(time.Now()), Target: m.Name, Platform: m.Platform, Tier: m.Tier,
		Version: md.Version, HarnessCommit: commit, StartedAt: time.Now().UTC()}
	var pass, fail, skip int
	for _, p := range probes {
		if len(only) > 0 && !containsAny(p.Path, only) {
			continue
		}
		r := probe.Run(ctx, env, p)
		status := "PASS"
		switch {
		case r.Observed.Kind == "skipped":
			status = "SKIP"
			skip++
		case r.Pass:
			pass++
		default:
			status = "FAIL"
			fail++
		}
		obs := r.Observed.String()
		if r.Observed.Kind == "error" || r.Observed.Kind == "skipped" {
			obs += " (" + r.Observed.Message + ")"
		}
		fmt.Printf("%-4s %-48s expect %-24s observed %s\n", status, strings.TrimPrefix(p.Path, "probes/"), r.Expect, truncate(obs, 160))
		run.Results = append(run.Results, r)
	}
	dir := filepath.Join(f.resultsDir, "probes", filepath.FromSlash(m.Name), run.RunID)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	for name, v := range map[string]any{"probes.json": run, "metadata.json": md} {
		b, _ := json.MarshalIndent(v, "", "  ")
		if err := os.WriteFile(filepath.Join(dir, name), append(b, '\n'), 0o644); err != nil {
			return err
		}
	}
	fmt.Printf("\n%s (%s): PASS %d, FAIL %d, SKIP %d -> %s\n", m.Name, md.Version, pass, fail, skip, dir)
	return nil
}

func containsAny(s string, subs []string) bool {
	for _, x := range subs {
		if strings.Contains(s, x) {
			return true
		}
	}
	return false
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}

func newFeatureMatrixCmd() *cobra.Command {
	var yamlPath, resultsDir, out string
	cmd := &cobra.Command{
		Use:   "feature-matrix",
		Short: "Render docs/feature-matrix.md from docs/feature-matrix.yaml and the latest probe results",
		RunE: func(_ *cobra.Command, _ []string) error {
			m, err := probe.LoadMatrix(yamlPath)
			if err != nil {
				return err
			}
			results, coverage, err := latestProbeResults(filepath.Join(resultsDir, "probes"))
			if err != nil {
				return err
			}
			return os.WriteFile(out, []byte(m.Render(results, renderStatus(coverage))), 0o644)
		},
	}
	cmd.Flags().StringVar(&yamlPath, "yaml", "docs/feature-matrix.yaml", "matrix source")
	cmd.Flags().StringVar(&resultsDir, "results-dir", "results", "results root")
	cmd.Flags().StringVar(&out, "out", "docs/feature-matrix.md", "output markdown")
	cmd.SilenceUsage = true
	return cmd
}

// latestProbeResults returns, per probe path and platform, the result from the
// newest probe run of that platform (any tier), plus the newest run per platform.
func latestProbeResults(root string) (map[string]map[string]probe.TargetResult, map[string]probeRun, error) {
	files, _ := filepath.Glob(filepath.Join(root, "*", "*", "*", "probes.json"))
	var all []probeRun
	for _, p := range files {
		b, err := os.ReadFile(p)
		if err != nil {
			return nil, nil, err
		}
		var r probeRun
		if err := json.Unmarshal(b, &r); err != nil {
			return nil, nil, fmt.Errorf("%s: %w", p, err)
		}
		all = append(all, r)
	}
	sort.Slice(all, func(i, j int) bool { return all[i].RunID < all[j].RunID })
	out := map[string]map[string]probe.TargetResult{}
	coverage := map[string]probeRun{}
	for _, r := range all {
		coverage[r.Platform] = r
		for _, res := range r.Results {
			if res.Observed.Kind == "skipped" {
				continue
			}
			if out[res.Path] == nil {
				out[res.Path] = map[string]probe.TargetResult{}
			}
			out[res.Path][r.Platform] = probe.TargetResult{Platform: r.Platform, Tier: r.Tier, Version: r.Version, RunID: r.RunID, Result: res}
		}
	}
	return out, coverage, nil
}

func renderStatus(coverage map[string]probeRun) string {
	var parts []string
	for _, p := range []string{"cloud", "altinity", "upstream-oss"} {
		r, ok := coverage[p]
		if !ok {
			parts = append(parts, fmt.Sprintf("**%s:** not probed yet; its cells are cited hypotheses *(h)*.", p))
			continue
		}
		pass, fail := 0, 0
		for _, x := range r.Results {
			if x.Observed.Kind == "skipped" {
				continue
			}
			if x.Pass {
				pass++
			} else {
				fail++
			}
		}
		parts = append(parts, fmt.Sprintf("**%s:** probed on `%s` (tier `%s`, run `%s`): %d confirmed, %d contradicted.", p, r.Version, r.Tier, r.RunID, pass, fail))
	}
	return "> **Status: generated by `bench feature-matrix`.** A cell with a **Probed:** note carries the\n" +
		"> outcome of its probe(s) on the exact server version; a cell without one is a cited hypothesis\n" +
		"> *(h)*. Local tiers establish capability and defaults only, never performance.\n>\n> " +
		strings.Join(parts, "\n> ")
}

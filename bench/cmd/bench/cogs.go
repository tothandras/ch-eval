package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"time"

	"github.com/spf13/cobra"

	"github.com/tothandras/ch-eval/bench/cogs"
	"github.com/tothandras/ch-eval/bench/runner"
	"github.com/tothandras/ch-eval/bench/runs"
	"github.com/tothandras/ch-eval/bench/scenario"
	"github.com/tothandras/ch-eval/bench/target"
)

type cogsFlags struct {
	dsn          string
	target       string
	targetsDir   string
	workload     string
	cell         string
	runProfile   string
	skipInit     bool
	allowDrift   bool
	requireClean bool
	resultsDir   string
	scenariosDir string
	cellsDir     string
	pricingDir   string
	usageExport  string

	pricingProfile string // override the manifest's pricing_profile (e.g. local-zero for smoke runs)
	preloadRows    int    // override the manifest's preload_rows; -1 = manifest value
	preloadWorkers int    // parallel preload connections (disjoint generator ranges)
}

func newCogsCmd() *cobra.Command {
	f := &cogsFlags{}
	cmd := &cobra.Command{
		Use:   "cogs",
		Short: "Run a COGS workload cell: paced ingest + weighted query replay, attributed and priced",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runCogs(cmd.Context(), f)
		},
	}
	cmd.Flags().StringVar(&f.dsn, "dsn", "", "ClickHouse DSN (overrides the manifest's dsn_env and $CLICKHOUSE_DSN)")
	cmd.Flags().StringVar(&f.target, "target", "", "target manifest <platform>/<tier> (required)")
	cmd.Flags().StringVar(&f.targetsDir, "targets-dir", "targets", "directory holding target manifests")
	cmd.Flags().StringVar(&f.workload, "workload", "openmeter", "workload of the cell's scenario (a cell may also name <workload>/<scenario>)")
	cmd.Flags().StringVar(&f.cell, "cell", "", "workload cell: a name resolved in --cells-dir, or a path to a manifest (required)")
	cmd.Flags().StringVar(&f.runProfile, "profile", "", "run profile: \"ci\" shortens phases to 2m/3m/1m for smoke runs")
	cmd.Flags().BoolVar(&f.skipInit, "skip-init", false, "reuse the existing table and data (skips the fresh migration and preload)")
	cmd.Flags().BoolVar(&f.requireClean, "require-clean", false, "refuse to run if the harness git tree is dirty")
	cmd.Flags().BoolVar(&f.allowDrift, "allow-drift", false, "run even when the live target differs from its manifest")
	cmd.Flags().StringVar(&f.resultsDir, "results-dir", "results", "results root; cogs runs go to <root>/<workload>/<scenario>/<platform>/<tier>/cogs/")
	cmd.Flags().StringVar(&f.scenariosDir, "scenarios-dir", "scenarios", "directory holding <workload>/<scenario>/")
	cmd.Flags().StringVar(&f.cellsDir, "cells-dir", "cells", "directory holding cell manifests")
	cmd.Flags().StringVar(&f.pricingDir, "pricing-dir", "pricing", "directory holding pricing profiles")
	cmd.Flags().StringVar(&f.usageExport, "usage-export", "", "ClickHouse Cloud usage export (cogs-usage/v1) to reconcile against")
	cmd.Flags().StringVar(&f.pricingProfile, "pricing-profile", "", "override the cell's pricing profile (e.g. local-zero for resource-accounting-only runs)")
	cmd.Flags().IntVar(&f.preloadRows, "preload-rows", -1, "override the cell's preload_rows (smoke runs); -1 keeps the manifest value")
	cmd.Flags().IntVar(&f.preloadWorkers, "preload-workers", 4, "parallel preload connections; deterministic (disjoint generator index ranges)")
	_ = cmd.MarkFlagRequired("cell")

	cmd.AddCommand(newCogsCompareCmd(), newCogsValidateCmd(), newCogsReconcileCmd())
	return cmd
}

func newCogsReconcileCmd() *cobra.Command {
	var resultsDir string
	var write bool
	cmd := &cobra.Command{
		Use:   "reconcile <run> <usage-file>",
		Short: "Reconcile a finished cogs run against a Cloud usage export (cogs-usage/v1 JSON or the usage-statement CSV)",
		Args:  cobra.ExactArgs(2),
		RunE: func(_ *cobra.Command, args []string) error {
			res, path, err := cogs.LoadResult(resultsDir, args[0])
			if err != nil {
				return err
			}
			modelStorageTB := float64(res.Accounting.Storage.DrainEnd.CompressedBytes) / 1e12
			rec, err := cogs.ReconcileFile(args[1], res.Pricing, res.Phases.Measure, modelStorageTB)
			if err != nil {
				return err
			}
			res.Reconciliation = &rec

			fmt.Printf("billed compute unit-hours: %.3f\nmodel  compute unit-hours: %.3f\ndelta: %.1f%%  (granularity: %s, flagged: %v)\n",
				rec.BilledComputeUnitHours, rec.ModelComputeUnitHours, rec.DeltaPct, rec.Granularity, rec.Flagged)

			if write {
				if err := os.WriteFile(path, marshalResult(res), 0o644); err != nil {
					return err
				}
				if _, err := cogs.WriteReport(resultsDir, res); err != nil {
					return err
				}
				fmt.Fprintf(os.Stderr, "updated: %s (+ markdown)\n", path)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&resultsDir, "results-dir", "results/openmeter", "workload results root; a run ref is <scenario>[/<platform>/<tier>] or a result path")
	cmd.Flags().BoolVar(&write, "write", false, "write the reconciliation block back into the result JSON and regenerate the markdown")
	return cmd
}

func marshalResult(r cogs.Result) []byte {
	b, _ := json.MarshalIndent(r, "", "  ")
	return append(b, '\n')
}

func runCogs(ctx context.Context, f *cogsFlags) error {
	commit := runner.HarnessCommit()
	if strings.HasSuffix(commit, "-dirty") {
		if f.requireClean {
			return fmt.Errorf("harness tree is dirty (%s); commit changes or drop --require-clean", commit)
		}
		fmt.Fprintf(os.Stderr, "warning: harness tree is dirty (%s); results will be tagged -dirty\n", commit)
	}

	m, opts, dsnSource, err := resolveTarget(&flags{dsn: f.dsn, target: f.target, targetsDir: f.targetsDir})
	if err != nil {
		return err
	}
	if m.IsLocal() && f.pricingProfile == "" {
		fmt.Fprintf(os.Stderr, "warning: %s is a local tier; its costs are resource-accounting smoke only (consider --pricing-profile local-zero)\n", m.Name)
	}

	cellPath := f.cell
	if !strings.HasSuffix(cellPath, ".json") {
		cellPath = filepath.Join(f.cellsDir, f.cell+".json")
	}
	cell, err := cogs.LoadCell(cellPath)
	if err != nil {
		return err
	}
	workload, scName := f.workload, cell.Scenario
	if w, n, ok := strings.Cut(cell.Scenario, "/"); ok {
		workload, scName = w, n
	}
	scs, err := scenario.Discover(f.scenariosDir, workload, []string{scName})
	if err != nil {
		return err
	}
	sc := scs[0]

	admin, err := connect(ctx, *opts, "")
	if err != nil {
		return err
	}
	defer admin.Close()
	md := target.Capture(ctx, admin, m, dsnSource, commit)
	if err := checkDrift(md, f.allowDrift); err != nil {
		return err
	}
	if !f.skipInit {
		fmt.Fprintf(os.Stderr, " migrations %s (fresh) ...\n", sc.Scope())
		if _, err := sc.Prepare(ctx, admin, m, true, runs.NewRunID(time.Now()), func(format string, a ...any) {
			fmt.Fprintf(os.Stderr, "   "+format+"\n", a...)
		}); err != nil {
			return fmt.Errorf("%s: migrations: %w", sc.Scope(), err)
		}
	}
	conn, err := connect(ctx, *opts, sc.Config.Database)
	if err != nil {
		return err
	}
	defer conn.Close()

	// SIGINT during measure truncates the window; the run still collects,
	// prices, and reports over what it measured.
	runCtx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	res, err := cogs.RunCell(runCtx, conn, cogs.Opts{
		CellPath:        cellPath,
		RunProfile:      f.runProfile,
		SkipInit:        f.skipInit,
		ResultsDir:      f.resultsDir,
		PricingDir:      f.pricingDir,
		Scenario:        sc,
		Target:          m,
		Database:        sc.Config.Database,
		HarnessCommit:   commit,
		UsageExport:     f.usageExport,
		PricingOverride: f.pricingProfile,
		PreloadOverride: f.preloadRows,
		PreloadWorkers:  f.preloadWorkers,
	})
	if err != nil {
		return err
	}

	root := filepath.Join(f.resultsDir, sc.Workload)
	jsonPath, err := cogs.Write(root, res)
	if err != nil {
		return fmt.Errorf("write result: %w", err)
	}
	mdPath, err := cogs.WriteReport(root, res)
	if err != nil {
		return fmt.Errorf("write report: %w", err)
	}
	for _, p := range []string{jsonPath, mdPath} {
		if rel, err := filepath.Rel(".", p); err == nil && !strings.HasPrefix(rel, "..") {
			p = rel
		}
		fmt.Printf(" wrote: %s\n", p)
	}
	if len(res.Errors) > 0 {
		fmt.Fprintf(os.Stderr, "completed with %d recorded error(s); see the result file\n", len(res.Errors))
	}
	return nil
}

func newCogsCompareCmd() *cobra.Command {
	var resultsDir string
	var allowProfileMismatch bool
	cmd := &cobra.Command{
		Use:   "compare <run-a> <run-b>",
		Short: "Diff two cogs runs' unit costs (args: result paths, or scenario names resolving to the latest run)",
		Args:  cobra.ExactArgs(2),
		RunE: func(_ *cobra.Command, args []string) error {
			a, pathA, err := cogs.LoadResult(resultsDir, args[0])
			if err != nil {
				return err
			}
			b, pathB, err := cogs.LoadResult(resultsDir, args[1])
			if err != nil {
				return err
			}
			fmt.Fprintf(os.Stderr, "A: %s\nB: %s\n\n", pathA, pathB)
			out, err := cogs.Compare(a, b, allowProfileMismatch)
			if err != nil {
				return err
			}
			fmt.Print(out)
			return nil
		},
	}
	cmd.Flags().StringVar(&resultsDir, "results-dir", "results/openmeter", "workload results root; a run ref is <scenario>[/<platform>/<tier>] or a result path")
	cmd.Flags().BoolVar(&allowProfileMismatch, "allow-profile-mismatch", false, "compare resource lines even when pricing profiles differ (prices are omitted)")
	return cmd
}

func newCogsValidateCmd() *cobra.Command {
	var resultsDir string
	cmd := &cobra.Command{
		Use:   "validate <ingest-only> <query-only> <mixed>",
		Short: "Check cpu-linear additivity across an ingest-only, a query-only, and a mixed run",
		Args:  cobra.ExactArgs(3),
		RunE: func(_ *cobra.Command, args []string) error {
			var rs [3]cogs.Result
			for i, ref := range args {
				r, path, err := cogs.LoadResult(resultsDir, ref)
				if err != nil {
					return err
				}
				fmt.Fprintf(os.Stderr, "%d: %s\n", i+1, path)
				rs[i] = r
			}
			out, err := cogs.ValidateAdditivity(rs[0], rs[1], rs[2])
			if err != nil {
				return err
			}
			fmt.Print(out)
			return nil
		},
	}
	cmd.Flags().StringVar(&resultsDir, "results-dir", "results/openmeter", "workload results root; a run ref is <scenario>[/<platform>/<tier>] or a result path")
	return cmd
}

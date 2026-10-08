package main

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"

	"github.com/tothandras/ch-eval/bench/runs"
	"github.com/tothandras/ch-eval/bench/scenario"
	"github.com/tothandras/ch-eval/migrations/migrate"
)

// newMigrateCmd builds `bench migrate <status|up|down|force|render>` for one
// scenario's init migrations on one target (schema-migrations spec).
func newMigrateCmd() *cobra.Command {
	f := &flags{}
	var to, version uint64
	var status string
	var measure bool
	cmd := &cobra.Command{
		Use:   "migrate <status|up|down|verify|force|render>",
		Short: "Apply, roll back, inspect or render a scenario's migrations on a target",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return migrateCmd(cmd.Context(), f, args[0], to, version, status, measure)
		},
	}
	cmd.Flags().StringVar(&f.target, "target", "", "target manifest <platform>/<tier> (required)")
	cmd.Flags().StringVar(&f.targetsDir, "targets-dir", "targets", "directory holding target manifests")
	cmd.Flags().StringVar(&f.workload, "workload", "openmeter", "workload")
	cmd.Flags().StringSliceVar(&f.scenarios, "scenario", nil, "scenario name (exactly one)")
	cmd.Flags().StringVar(&f.scenariosDir, "scenarios-dir", "scenarios", "directory holding <workload>/<scenario>/")
	cmd.Flags().StringVar(&f.dsn, "dsn", "", "ClickHouse DSN override")
	cmd.Flags().Uint64Var(&to, "to", 0, "down: roll back versions newer than this (0 = all)")
	cmd.Flags().Uint64Var(&version, "version", 0, "force: version to overwrite")
	cmd.Flags().StringVar(&status, "status", "", "force: status to record (applied|down)")
	cmd.Flags().BoolVar(&measure, "measure", false, "up/down: capture per-statement cost from query_log and part_log")
	cmd.SilenceUsage = true
	return cmd
}

func migrateCmd(ctx context.Context, f *flags, action string, to, version uint64, status string, measure bool) error {
	if len(f.scenarios) != 1 {
		return fmt.Errorf("--scenario must name exactly one scenario")
	}
	m, opts, _, err := resolveTarget(f)
	if err != nil {
		return err
	}
	scs, err := scenario.Discover(f.scenariosDir, f.workload, f.scenarios)
	if err != nil {
		return err
	}
	sc := scs[0]
	migs, err := sc.Migrations()
	if err != nil {
		return err
	}
	topo, err := m.Topology(sc.Config.Database)
	if err != nil {
		return err
	}
	if action == "render" { // offline: print what would run on this target
		for _, mg := range migs {
			for _, dir := range []migrate.Direction{migrate.Up, migrate.Down} {
				rend, err := migrate.Render(mg, dir, topo)
				if err != nil {
					return err
				}
				fmt.Printf("-- %d_%s.%s (%s) checksum %s\n", mg.Version, mg.Name, dir, topo, rend.Checksum)
				for _, s := range rend.Statements {
					fmt.Println(s + ";")
				}
				for _, l := range migrate.Lint(rend.Statements, topo) {
					fmt.Println("-- LINT:", l)
				}
			}
		}
		return nil
	}
	conn, err := connect(ctx, *opts, "")
	if err != nil {
		return err
	}
	defer conn.Close()
	r := &migrate.Runner{Conn: conn, Topo: topo, Scope: sc.Scope(), Target: m.Name,
		RunID: runs.NewRunID(time.Now()), Measure: measure,
		Logf: func(format string, a ...any) { fmt.Printf(format+"\n", a...) }}
	if err := r.EnsureState(ctx); err != nil {
		return err
	}
	switch action {
	case "status":
		st, err := r.Status(ctx)
		if err != nil {
			return err
		}
		tw := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
		fmt.Fprintf(tw, "VERSION\tNAME\tSTATUS\tAPPLIED_AT\n")
		for _, mg := range migs {
			s, ok := st[mg.Version]
			if !ok {
				fmt.Fprintf(tw, "%d\t%s\tnot applied\t\n", mg.Version, mg.Name)
				continue
			}
			fmt.Fprintf(tw, "%d\t%s\t%s\t%s\n", mg.Version, mg.Name, s.Status, s.AppliedAt.UTC().Format(time.RFC3339))
		}
		return tw.Flush()
	case "up":
		done, err := r.Up(ctx, migs)
		printApplied(done)
		return err
	case "verify":
		problems, err := sc.Verify(ctx, conn, m)
		if err != nil {
			return err
		}
		for _, p := range problems {
			fmt.Println("FAIL", p)
		}
		if len(problems) > 0 {
			return fmt.Errorf("%s on %s: %d schema problem(s)", sc.Scope(), m.Name, len(problems))
		}
		fmt.Printf("OK %s on %s: schema matches the %s dialect on every replica\n", sc.Scope(), m.Name, m.Dialect)
		return nil
	case "down":
		done, err := r.Down(ctx, migs, to)
		printApplied(done)
		return err
	case "force":
		if status != migrate.StatusApplied && status != migrate.StatusDown {
			return fmt.Errorf("--status must be applied or down")
		}
		for _, mg := range migs {
			if mg.Version == version {
				return r.Force(ctx, mg, status)
			}
		}
		return fmt.Errorf("version %s not found in %s", strconv.FormatUint(version, 10), sc.Scope())
	}
	return fmt.Errorf("unknown action %q (status|up|down|verify|force|render)", action)
}

func printApplied(done []migrate.Applied) {
	for _, a := range done {
		fmt.Printf("%s %d_%s: %s (%.0f ms, %d statements)\n", a.Direction, a.Version, a.Name, a.Status, a.DurationMs, len(a.Statements))
		for _, s := range a.Statements {
			if s.Error != "" {
				fmt.Printf("  error: %s\n", s.Error)
			}
		}
	}
}

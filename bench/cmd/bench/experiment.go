package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/spf13/cobra"

	"github.com/tothandras/ch-eval/bench/experiments"
	"github.com/tothandras/ch-eval/bench/runner"
	"github.com/tothandras/ch-eval/bench/runs"
	"github.com/tothandras/ch-eval/bench/target"
)

var experimentFuncs = map[string]func(context.Context, experiments.Env, driver.Conn) (experiments.Result, error){
	"S3": experiments.S3, "S4": experiments.S4, "R6": experiments.R6, "R3": experiments.R3,
}

func newExperimentCmd() *cobra.Command {
	f := &flags{}
	cmd := &cobra.Command{
		Use:   "experiment <S3|S4|R6|R3>...",
		Short: "Run local correctness experiments that stop and start target processes (D13)",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runExperiments(cmd.Context(), f, args)
		},
	}
	cmd.Flags().StringVar(&f.target, "target", "", "local multi-node target <platform>/<tier> (required)")
	cmd.Flags().StringVar(&f.targetsDir, "targets-dir", "targets", "directory holding target manifests")
	cmd.Flags().StringVar(&f.resultsDir, "results-dir", "results", "results root; runs go to <root>/experiments/<id>/<platform>/<tier>/<run-id>/")
	cmd.SilenceUsage = true
	return cmd
}

func runExperiments(ctx context.Context, f *flags, ids []string) error {
	m, opts, dsnSource, err := resolveTarget(f)
	if err != nil {
		return err
	}
	if !m.IsLocal() || m.Ports == nil || !strings.HasPrefix(m.Deployment, "compose-") {
		return fmt.Errorf("%s: these experiments stop containers and run on compose tiers only", m.Name)
	}
	ctl, err := experiments.NewCompose(m)
	if err != nil {
		return err
	}
	admin, err := connect(ctx, *opts, "")
	if err != nil {
		return err
	}
	defer admin.Close()
	md := target.Capture(ctx, admin, m, dsnSource, runner.HarnessCommit())
	env := experiments.Env{M: m, Opts: opts, Control: ctl, Logf: func(string, ...any) {}}
	for _, id := range ids {
		fn, ok := experimentFuncs[strings.ToUpper(id)]
		if !ok {
			return fmt.Errorf("unknown experiment %q", id)
		}
		fmt.Printf("== %s on %s\n", strings.ToUpper(id), m.Name)
		res, err := fn(ctx, env, admin)
		res.Version = md.Version
		for _, s := range res.Steps {
			fmt.Printf("   %-72s %-14s %8.0f ms %s\n", s.Name, s.Outcome, s.Millis, truncate(s.Detail, 100))
		}
		for _, x := range res.Findings {
			fmt.Println("   finding:", x)
		}
		if err != nil {
			fmt.Fprintf(os.Stderr, "   error: %v\n", err)
		}
		dir := filepath.Join(f.resultsDir, "experiments", res.ID, filepath.FromSlash(m.Name), runs.NewRunID(time.Now()))
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
		b, _ := json.MarshalIndent(res, "", "  ")
		_ = os.WriteFile(filepath.Join(dir, "result.json"), append(b, '\n'), 0o644)
		_ = os.WriteFile(filepath.Join(dir, "report.md"), []byte(experimentReport(res, m)), 0o644)
		if err != nil {
			return err
		}
		// reconnect: the admin node may have been restarted
		if admin.Ping(ctx) != nil {
			admin.Close()
			if admin, err = connect(ctx, *opts, ""); err != nil {
				return err
			}
		}
	}
	return nil
}

func experimentReport(r experiments.Result, m target.Manifest) string {
	var b strings.Builder
	fmt.Fprintf(&b, "# %s: %s\n\nTarget `%s` (`%s`), %s.\n\n", r.ID, r.Question, r.Target, r.Version, r.StartedAt.Format(time.RFC3339))
	if m.IsLocal() {
		b.WriteString("> Local tier: behaviour (errors, partial results, catch-up) only; timings are not production evidence.\n\n")
	}
	b.WriteString("| Step | Outcome | ms | Detail |\n| --- | --- | ---: | --- |\n")
	for _, s := range r.Steps {
		fmt.Fprintf(&b, "| %s | %s | %.0f | %s |\n", s.Name, s.Outcome, s.Millis, strings.ReplaceAll(s.Detail, "|", "\\|"))
	}
	b.WriteString("\n## Findings\n\n")
	for _, x := range r.Findings {
		b.WriteString("- " + x + "\n")
	}
	return b.String()
}

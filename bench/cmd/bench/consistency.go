package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/spf13/cobra"

	"github.com/tothandras/ch-eval/bench/consistency"
	"github.com/tothandras/ch-eval/bench/runner"
	"github.com/tothandras/ch-eval/bench/runs"
	"github.com/tothandras/ch-eval/bench/target"
)

func newConsistencyCmd() *cobra.Command {
	f := &flags{}
	var policies, inserts []string
	var c consistency.Config
	var restartNode int
	var restartCmd string
	cmd := &cobra.Command{
		Use:   "consistency",
		Short: "Replica-consistency experiment: ack-to-visible lag per replica and stale/backward reads per policy",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runConsistency(cmd.Context(), f, c, policies, inserts, restartNode, restartCmd)
		},
	}
	cmd.Flags().StringVar(&f.target, "target", "", "target manifest <platform>/<tier> (required; needs replicas)")
	cmd.Flags().StringVar(&f.targetsDir, "targets-dir", "targets", "directory holding target manifests")
	cmd.Flags().StringVar(&f.dsn, "dsn", "", "ClickHouse DSN override")
	cmd.Flags().StringVar(&f.resultsDir, "results-dir", "results", "results root; runs go to <root>/consistency/<platform>/<tier>/<run-id>/")
	cmd.Flags().StringSliceVar(&policies, "policy", nil, "policies (default: default,quorum,sync-replica on self-managed; default,seq on cloud)")
	cmd.Flags().StringSliceVar(&inserts, "insert", []string{"sync", "async"}, "insert modes: sync, async")
	cmd.Flags().DurationVar(&c.Duration, "duration", 60*time.Second, "duration per policy run")
	cmd.Flags().DurationVar(&c.BatchEvery, "batch-every", 200*time.Millisecond, "one insert batch per interval")
	cmd.Flags().IntVar(&c.BatchRows, "batch-rows", 100, "rows per batch")
	cmd.Flags().IntVar(&c.Readers, "readers", 4, "concurrent readers")
	cmd.Flags().DurationVar(&c.ReadEvery, "read-every", 50*time.Millisecond, "pause between a reader's reads")
	cmd.Flags().DurationVar(&c.PollEvery, "poll-every", 50*time.Millisecond, "per-replica visibility poll interval")
	cmd.Flags().IntVar(&restartNode, "restart-node", 0, "compose tiers: restart this node (1-based, in targets/.state/<deployment>/nodes) at half time")
	cmd.Flags().StringVar(&restartCmd, "restart-cmd", "", "shell command run at half time (overrides --restart-node)")
	cmd.SilenceUsage = true
	return cmd
}

func runConsistency(ctx context.Context, f *flags, c consistency.Config, policies, inserts []string, restartNode int, restartCmd string) error {
	m, opts, dsnSource, err := resolveTarget(f)
	if err != nil {
		return err
	}
	if m.Dialect == "single" {
		return fmt.Errorf("%s is a single node: the experiment needs replicas", m.Name)
	}
	admin, err := connect(ctx, *opts, "")
	if err != nil {
		return err
	}
	defer admin.Close()
	md := target.Capture(ctx, admin, m, dsnSource, runner.HarnessCommit())
	n := consistency.Nodes{Admin: admin}
	if m.Ports != nil {
		host, _, _ := net.SplitHostPort(opts.Addr[0])
		for i := 0; i < m.Shards*m.Replicas; i++ {
			o := *opts
			o.Addr = []string{net.JoinHostPort(host, strconv.Itoa(m.Ports.NativeBase+i))}
			conn, err := connect(ctx, o, "")
			if err != nil {
				return err
			}
			defer conn.Close()
			n.Node = append(n.Node, conn)
			n.Names = append(n.Names, fmt.Sprintf("node%d", i+1))
		}
	} else {
		// Cloud or a service endpoint: the endpoint itself load-balances.
		n.Node, n.Names = []driver.Conn{admin}, []string{"endpoint"}
	}
	if restartCmd == "" && restartNode > 0 && strings.HasPrefix(m.Deployment, "compose-") {
		b, err := os.ReadFile(filepath.Join("targets", ".state", m.Platform+"-"+m.Deployment, "nodes"))
		if err != nil {
			return err
		}
		nodes := strings.Fields(string(b))
		if restartNode > len(nodes) {
			return fmt.Errorf("--restart-node %d: deployment has %d nodes", restartNode, len(nodes))
		}
		restartCmd = fmt.Sprintf("docker restart cheval-%s-%s-%s-1", m.Platform, m.Deployment, nodes[restartNode-1])
	}
	c.Database, c.Table = "ch_eval_consistency", "events"
	c.Cluster = m.SystemCluster()
	c.Cloud = m.Dialect == "cloud"
	c.Replicas = max(len(md.Replicas)/max(m.Shards, 1), 1)
	c.RestartCmd = restartCmd
	c.Logf = func(format string, a ...any) { fmt.Fprintf(os.Stderr, "  "+format+"\n", a...) }
	if len(policies) == 0 {
		policies = []string{"default", "quorum", "sync-replica"}
		if c.Cloud {
			policies = []string{"default", "seq"}
		}
	}
	runID := runs.NewRunID(time.Now())
	dir := filepath.Join(f.resultsDir, "consistency", filepath.FromSlash(m.Name), runID)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	var all []consistency.Result
	for _, p := range policies {
		for _, ins := range inserts {
			cc := c
			cc.Policy = consistency.Policy(p)
			cc.AsyncInsert = ins == "async"
			fmt.Printf("== %s policy=%s insert=%s duration=%s restart=%v\n", m.Name, p, ins, cc.Duration, cc.RestartCmd != "")
			res, err := consistency.Run(ctx, n, cc)
			if err != nil {
				return err
			}
			fmt.Printf("   batches=%d insert p50/p99=%.1f/%.1fms errors=%d | lag p50/p99/max=%.0f/%.0f/%.0fms | reads=%d stale=%d (%.2f%%) violations=%d read p50/p99=%.1f/%.1fms errors=%d\n",
				res.Batches, res.InsertMs.P50, res.InsertMs.P99, res.InsertErrors, res.LagMs["all"].P50, res.LagMs["all"].P99, res.LagMs["all"].Max,
				res.Reads, res.Stale, 100*res.StaleRate, res.Violations, res.ReadMs.P50, res.ReadMs.P99, res.ReadErrors)
			name := fmt.Sprintf("%s-%s.json", p, ins)
			b, _ := json.MarshalIndent(res, "", "  ")
			if err := os.WriteFile(filepath.Join(dir, name), append(b, '\n'), 0o644); err != nil {
				return err
			}
			all = append(all, res)
		}
	}
	_ = admin.Exec(ctx, fmt.Sprintf("DROP TABLE IF EXISTS %s.%s%s SYNC", c.Database, c.Table, onClusterOf(m)))
	b, _ := json.MarshalIndent(md, "", "  ")
	_ = os.WriteFile(filepath.Join(dir, "metadata.json"), append(b, '\n'), 0o644)
	report := consistencyReport(m, md.Version, runID, all)
	if err := os.WriteFile(filepath.Join(dir, "report.md"), []byte(report), 0o644); err != nil {
		return err
	}
	fmt.Printf("wrote %s\n", dir)
	return nil
}

func onClusterOf(m target.Manifest) string {
	if m.Dialect == "replicated" {
		return fmt.Sprintf(" ON CLUSTER '%s'", m.Cluster)
	}
	return ""
}

func consistencyReport(m target.Manifest, version, runID string, rs []consistency.Result) string {
	var b strings.Builder
	fmt.Fprintf(&b, "# Replica consistency: %s (`%s`), run `%s`\n\n", m.Name, version, runID)
	if m.IsLocal() {
		fmt.Fprintf(&b, "> Local tier: this run checks that the experiment works and shows the shape of the behaviour; it is not evidence about production latencies.\n\n")
	}
	b.WriteString("| Policy | Insert | Restart | Batches | Insert p50/p99 ms | Insert errors | Lag p50/p99/max ms | Never visible | Reads | Stale | Stale rate | Violations /10k | Deficit p99 rows / ms | Read p50/p99 ms | Read errors |\n")
	b.WriteString("| --- | --- | --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |\n")
	for _, r := range rs {
		ins := "sync"
		if r.Config.AsyncInsert {
			ins = "async"
		}
		never := 0
		for h, v := range r.NotVisible {
			if h != "all" {
				never = max(never, v)
			}
		}
		fmt.Fprintf(&b, "| %s | %s | %v | %d | %.1f / %.1f | %d | %.0f / %.0f / %.0f | %d | %d | %d | %.2f%% | %.1f | %.0f / %.0f | %.1f / %.1f | %d |\n",
			r.Config.Policy, ins, r.RestartAt != nil, r.Batches, r.InsertMs.P50, r.InsertMs.P99, r.InsertErrors,
			r.LagMs["all"].P50, r.LagMs["all"].P99, r.LagMs["all"].Max, never, r.Reads, r.Stale, 100*r.StaleRate,
			r.ViolationsPer10k, r.DeficitRows.P99, r.DeficitMs.P99, r.ReadMs.P50, r.ReadMs.P99, r.ReadErrors)
	}
	b.WriteString("\nLag is measured from insert acknowledgement to the batch being complete on a replica (per-replica polls through `clusterAllReplicas`). A read is stale when its total is below the total acknowledged before it started; a violation is a total lower than the same reader's previous read.\n")
	return b.String()
}

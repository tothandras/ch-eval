// Package accounting collects cogs resource consumption from ClickHouse
// system tables. All log collectors are multi-replica aware: system.query_log
// and system.part_log are PER-REPLICA tables, and on ClickHouse Cloud
// connections are load-balanced across replicas, so reading only the connected
// node's tables undercounts nondeterministically (the error lands in the idle
// residual). Collectors therefore read via clusterAllReplicas(default, ...)
// whenever the target exposes a `default` cluster, falling back to local
// tables on single-node setups without one.
package accounting

import (
	"context"
	"fmt"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
)

// Window is an attribution window: [Start, End).
type Window struct {
	Start time.Time
	End   time.Time
}

// Target describes how to read system tables on the connected service.
type Target struct {
	// UseCluster selects clusterAllReplicas(<Cluster>, system.X) over plain
	// system.X reads.
	UseCluster bool
	// Cluster is the cluster whose replicas are read; "default" on Cloud, the
	// manifest's cluster (for example ch_eval) on self-managed targets.
	Cluster string
	// Replicas is the reachable node count of Cluster (1 when absent).
	Replicas int
}

// DetectTarget probes the `default` cluster. On ClickHouse Cloud it exposes
// all replicas (the documented clusterAllReplicas pattern); the devenv single
// node also exposes it with one entry; bare servers without it read locally.
//
// Replica count comes from clusterAllReplicas(default, system.one) — one row
// per REACHABLE replica — not from counting system.clusters rows: Cloud's
// cluster metadata can list more pods than the service actually runs
// (observed: 6 listed, 3 reachable on a 3-replica service), which would
// inflate the coverage denominator.
func DetectTarget(ctx context.Context, conn driver.Conn) (Target, error) {
	return DetectTargetCluster(ctx, conn, "default")
}

// DetectTargetCluster is DetectTarget for a named cluster: self-managed targets
// name their cluster in the target manifest (ADR-0002), and their `default`
// cluster, when present, need not span the replicas.
func DetectTargetCluster(ctx context.Context, conn driver.Conn, cluster string) (Target, error) {
	if cluster == "" {
		return Target{UseCluster: false, Replicas: 1}, nil
	}
	var listed uint64
	err := conn.QueryRow(ctx, "SELECT count() FROM system.clusters WHERE cluster = ?", cluster).Scan(&listed)
	if err != nil {
		return Target{}, fmt.Errorf("accounting: probe %s cluster: %w", cluster, err)
	}
	if listed == 0 {
		return Target{UseCluster: false, Replicas: 1}, nil
	}
	var reachable uint64
	err = conn.QueryRow(ctx,
		fmt.Sprintf("SELECT count() FROM clusterAllReplicas(%s, system.one) SETTINGS skip_unavailable_shards = 1", cluster)).Scan(&reachable)
	if err != nil {
		return Target{}, fmt.Errorf("accounting: count reachable replicas: %w", err)
	}
	if reachable == 0 {
		reachable = 1
	}
	return Target{UseCluster: true, Cluster: cluster, Replicas: int(reachable)}, nil
}

// table renders the FROM source for a system log table on this target.
func (t Target) table(name string) string {
	if t.UseCluster {
		return "clusterAllReplicas(" + t.clusterName() + ", system." + name + ")"
	}
	return "system." + name
}

func (t Target) clusterName() string {
	if t.Cluster == "" {
		return "default"
	}
	return t.Cluster
}

// settingsSuffix tolerates cluster metadata listing unreachable pods (the
// Cloud over-listing above) on cluster-wide reads.
func (t Target) settingsSuffix() string {
	if t.UseCluster {
		return "\nSETTINGS skip_unavailable_shards = 1"
	}
	return ""
}

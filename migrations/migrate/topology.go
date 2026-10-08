package migrate

import "fmt"

// Dialect selects how DDL is rendered.
type Dialect string

const (
	// Single is one self-managed node: plain MergeTree engines, no ON CLUSTER.
	Single Dialect = "single"
	// Replicated is a self-managed cluster with ClickHouse Keeper: ReplicatedMergeTree
	// engines with explicit Keeper paths and ON CLUSTER DDL.
	Replicated Dialect = "replicated"
	// Cloud is ClickHouse Cloud: plain engine names (Cloud substitutes Shared* engines,
	// https://clickhouse.com/docs/products/cloud/guides/cloud-compatibility) and no ON CLUSTER.
	Cloud Dialect = "cloud"
)

// Topology is everything rendering needs to know about a target.
type Topology struct {
	Dialect  Dialect
	Database string
	// Cluster is the remote_servers cluster used for ON CLUSTER and Distributed tables
	// (replicated dialect only).
	Cluster string
	// Shards > 1 makes `table` resolve to the shard-local name and `distributed`
	// emit a Distributed table over it.
	Shards int
	// StoragePolicy, when set, is appended to every table's SETTINGS by the
	// `settings` helper (self-managed S3-backed MergeTree mode).
	StoragePolicy string
}

// Validate checks the topology is renderable.
func (t Topology) Validate() error {
	switch t.Dialect {
	case Single, Cloud:
		if t.Shards > 1 {
			return fmt.Errorf("migrate: dialect %q cannot be sharded", t.Dialect)
		}
	case Replicated:
		if t.Cluster == "" {
			return fmt.Errorf("migrate: replicated dialect needs a cluster name")
		}
	default:
		return fmt.Errorf("migrate: unknown dialect %q", t.Dialect)
	}
	if t.Database == "" {
		return fmt.Errorf("migrate: database is required")
	}
	return nil
}

// Sharded reports whether tables are split into local + Distributed.
func (t Topology) Sharded() bool { return t.Dialect == Replicated && t.Shards > 1 }

// String is a compact label recorded with applied migrations.
func (t Topology) String() string {
	if t.Dialect == Replicated {
		return fmt.Sprintf("%s(%s,shards=%d)", t.Dialect, t.Cluster, max(1, t.Shards))
	}
	return string(t.Dialect)
}

package target

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
)

// Metadata is everything about the live target a run must record (metadata.json).
type Metadata struct {
	CapturedAt  time.Time         `json:"captured_at"`
	Manifest    Manifest          `json:"manifest"`
	DSNSource   string            `json:"dsn_source"`
	Version     string            `json:"version"`
	Replicas    []ReplicaInfo     `json:"replicas"`
	Build       map[string]string `json:"build_options"`
	Clusters    []ClusterRow      `json:"clusters"`
	Macros      map[string]string `json:"macros"`
	Disks       []DiskRow         `json:"disks"`
	Policies    []string          `json:"storage_policies"`
	Caches      []CacheRow        `json:"filesystem_caches,omitempty"`
	CacheConfig []CacheSettings   `json:"filesystem_cache_settings,omitempty"`
	Changed     Changed           `json:"changed_settings"`
	Limits      Limits            `json:"limits"`
	Client      Client            `json:"client"`
	Drift       []string          `json:"drift,omitempty"`
	CaptureErrs []string          `json:"capture_errors,omitempty"`
}

// ReplicaInfo is one reachable server.
type ReplicaInfo struct {
	Host    string `json:"host"`
	Version string `json:"version"`
	Uptime  uint64 `json:"uptime_sec"`
}

// ClusterRow is one system.clusters row of the target's own cluster.
type ClusterRow struct {
	Cluster string `json:"cluster"`
	Shard   uint32 `json:"shard_num"`
	Replica uint32 `json:"replica_num"`
	Host    string `json:"host_name"`
}

// DiskRow is one system.disks row.
type DiskRow struct {
	Name  string `json:"name"`
	Type  string `json:"type"`
	Total uint64 `json:"total_bytes"`
	Free  uint64 `json:"free_bytes"`
	// Locality (design D14): object storage type and the filesystem cache in front
	// of it, when the disk has one.
	ObjectStorage string `json:"object_storage_type,omitempty"`
	CachePath     string `json:"cache_path,omitempty"`
}

// CacheRow is the filesystem cache's live state (system.filesystem_cache totals).
type CacheRow struct {
	CachePath string `json:"cache_path"`
	Segments  uint64 `json:"segments"`
	Bytes     uint64 `json:"bytes"`
}

// CacheSettings is a filesystem cache's effective configuration on the connected
// node (system.filesystem_cache_settings): the knobs D14 compares across targets.
type CacheSettings struct {
	Name    string `json:"cache_name"`
	Path    string `json:"path"`
	MaxSize uint64 `json:"max_size"`
	OnWrite bool   `json:"cache_on_write_operations"`
	Policy  string `json:"cache_policy"`
}

// Changed holds settings that differ from defaults, per scope.
type Changed struct {
	Session   map[string]string `json:"session"`
	MergeTree map[string]string `json:"merge_tree"`
	Server    map[string]string `json:"server"`
}

// Limits are the compute limits the coverage denominator and shape check use.
type Limits struct {
	CGroupMaxCPU    float64 `json:"cgroup_max_cpu"`
	MaxThreads      uint64  `json:"max_threads"`
	OSMemoryTotal   uint64  `json:"os_memory_total"`
	CGroupMemLimit  uint64  `json:"cgroup_memory_limit"`
	MaxServerMemory uint64  `json:"max_server_memory_usage"`
}

// Client describes where the harness ran.
type Client struct {
	Hostname      string `json:"hostname"`
	GoVersion     string `json:"go_version"`
	HarnessCommit string `json:"harness_commit"`
	Benchmark     string `json:"clickhouse_benchmark,omitempty"`
}

// Capture records live metadata. Individual probe failures are recorded, not fatal:
// "cannot be measured on this target" is itself a finding.
func Capture(ctx context.Context, conn driver.Conn, m Manifest, dsnSource, harnessCommit string) Metadata {
	md := Metadata{CapturedAt: time.Now().UTC(), Manifest: m, DSNSource: dsnSource,
		Build: map[string]string{}, Macros: map[string]string{},
		Changed: Changed{Session: map[string]string{}, MergeTree: map[string]string{}, Server: map[string]string{}}}
	errf := func(what string, err error) {
		md.CaptureErrs = append(md.CaptureErrs, fmt.Sprintf("%s: %v", what, err))
	}

	if err := conn.QueryRow(ctx, "SELECT version()").Scan(&md.Version); err != nil {
		errf("version", err)
	}
	src := "system.one"
	if c := m.SystemCluster(); c != "" {
		src = fmt.Sprintf("clusterAllReplicas('%s', system.one)", c)
	}
	if rows, err := conn.Query(ctx, fmt.Sprintf("SELECT hostName(), version(), toUInt64(uptime()) FROM %s ORDER BY 1 SETTINGS skip_unavailable_shards = 1", src)); err == nil {
		for rows.Next() {
			var r ReplicaInfo
			if err := rows.Scan(&r.Host, &r.Version, &r.Uptime); err == nil {
				md.Replicas = append(md.Replicas, r)
			}
		}
		rows.Close()
	} else {
		errf("replicas", err)
	}
	kv := func(what, q string, into map[string]string) {
		rows, err := conn.Query(ctx, q)
		if err != nil {
			errf(what, err)
			return
		}
		defer rows.Close()
		for rows.Next() {
			var k, v string
			if err := rows.Scan(&k, &v); err == nil {
				into[k] = v
			}
		}
	}
	kv("build_options", "SELECT name, value FROM system.build_options WHERE name IN ('VERSION_FULL','VERSION_DESCRIBE','VERSION_GITHASH','GIT_HASH','BUILD_TYPE','SYSTEM_PROCESSOR','SYSTEM','CXX_COMPILER','VERSION_OFFICIAL')", md.Build)
	kv("macros", "SELECT macro, substitution FROM system.macros", md.Macros)
	kv("changed session settings", "SELECT name, value FROM system.settings WHERE changed", md.Changed.Session)
	kv("changed merge_tree settings", "SELECT name, value FROM system.merge_tree_settings WHERE changed", md.Changed.MergeTree)
	kv("changed server settings", "SELECT name, value FROM system.server_settings WHERE changed", md.Changed.Server)

	if m.Cluster != "" {
		if rows, err := conn.Query(ctx, "SELECT cluster, shard_num, replica_num, host_name FROM system.clusters WHERE cluster = ? ORDER BY shard_num, replica_num", m.Cluster); err == nil {
			for rows.Next() {
				var r ClusterRow
				if err := rows.Scan(&r.Cluster, &r.Shard, &r.Replica, &r.Host); err == nil {
					md.Clusters = append(md.Clusters, r)
				}
			}
			rows.Close()
		} else {
			errf("clusters", err)
		}
	}
	if rows, err := conn.Query(ctx, "SELECT name, toString(type), total_space, free_space, toString(object_storage_type), cache_path FROM system.disks ORDER BY name"); err == nil {
		for rows.Next() {
			var d DiskRow
			if err := rows.Scan(&d.Name, &d.Type, &d.Total, &d.Free, &d.ObjectStorage, &d.CachePath); err == nil {
				md.Disks = append(md.Disks, d)
			}
		}
		rows.Close()
	} else {
		errf("disks", err)
	}
	if rows, err := conn.Query(ctx, "SELECT cache_path, count(), sum(size) FROM system.filesystem_cache GROUP BY cache_path ORDER BY cache_path"); err == nil {
		for rows.Next() {
			var c CacheRow
			if rows.Scan(&c.CachePath, &c.Segments, &c.Bytes) == nil {
				md.Caches = append(md.Caches, c)
			}
		}
		rows.Close()
	} else {
		errf("filesystem_cache", err)
	}
	if rows, err := conn.Query(ctx, "SELECT cache_name, path, max_size, toBool(cache_on_write_operations), cache_policy FROM system.filesystem_cache_settings ORDER BY cache_name"); err == nil {
		for rows.Next() {
			var c CacheSettings
			if rows.Scan(&c.Name, &c.Path, &c.MaxSize, &c.OnWrite, &c.Policy) == nil {
				md.CacheConfig = append(md.CacheConfig, c)
			}
		}
		rows.Close()
	} else {
		errf("filesystem_cache_settings", err)
	}
	if rows, err := conn.Query(ctx, "SELECT DISTINCT policy_name FROM system.storage_policies ORDER BY 1"); err == nil {
		for rows.Next() {
			var p string
			if rows.Scan(&p) == nil {
				md.Policies = append(md.Policies, p)
			}
		}
		rows.Close()
	} else {
		errf("storage_policies", err)
	}
	_ = conn.QueryRow(ctx, "SELECT toFloat64(value) FROM system.asynchronous_metrics WHERE metric = 'CGroupMaxCPU'").Scan(&md.Limits.CGroupMaxCPU)
	_ = conn.QueryRow(ctx, "SELECT toUInt64(value) FROM system.asynchronous_metrics WHERE metric = 'OSMemoryTotal'").Scan(&md.Limits.OSMemoryTotal)
	_ = conn.QueryRow(ctx, "SELECT toUInt64(value) FROM system.asynchronous_metrics WHERE metric = 'CGroupMemoryTotal'").Scan(&md.Limits.CGroupMemLimit)
	_ = conn.QueryRow(ctx, "SELECT toUInt64(getSetting('max_threads'))").Scan(&md.Limits.MaxThreads)
	_ = conn.QueryRow(ctx, "SELECT toUInt64OrZero(value) FROM system.server_settings WHERE name = 'max_server_memory_usage'").Scan(&md.Limits.MaxServerMemory)

	host, _ := os.Hostname()
	md.Client = Client{Hostname: host, GoVersion: runtime.Version(), HarnessCommit: harnessCommit}
	if out, err := exec.Command("clickhouse-benchmark", "--version").Output(); err == nil {
		md.Client.Benchmark = strings.TrimSpace(string(out))
	}
	md.Drift = Drift(m, md)
	return md
}

// Drift lists contradictions between the manifest and the live target.
func Drift(m Manifest, md Metadata) []string {
	var out []string
	if m.ExpectedVersion != "" && md.Version != "" && md.Version != m.ExpectedVersion {
		out = append(out, fmt.Sprintf("version: manifest expects %s, server reports %s", m.ExpectedVersion, md.Version))
	}
	versions := map[string]bool{}
	for _, r := range md.Replicas {
		versions[r.Version] = true
	}
	if len(versions) > 1 {
		vs := make([]string, 0, len(versions))
		for v := range versions {
			vs = append(vs, v)
		}
		sort.Strings(vs)
		out = append(out, "replicas run mixed versions: "+strings.Join(vs, ", "))
	}
	want := m.Shards * m.Replicas
	if m.Dialect != "cloud" && want > 0 && len(md.Replicas) > 0 && len(md.Replicas) != want {
		out = append(out, fmt.Sprintf("replicas: manifest expects %d (%dx%d), %d reachable", want, m.Shards, m.Replicas, len(md.Replicas)))
	}
	if m.Dialect == "cloud" && m.Replicas > 0 && len(md.Replicas) > 0 && len(md.Replicas) != m.Replicas {
		out = append(out, fmt.Sprintf("replicas: manifest expects %d, %d reachable", m.Replicas, len(md.Replicas)))
	}
	if m.StoragePolicy != "" {
		found := false
		for _, p := range md.Policies {
			found = found || p == m.StoragePolicy
		}
		if !found {
			out = append(out, fmt.Sprintf("storage policy %q not defined on the server", m.StoragePolicy))
		}
	}
	return out
}

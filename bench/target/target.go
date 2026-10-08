// Package target makes the target platform a first-class benchmark dimension
// (ADR-0002): manifests under targets/<platform>/<tier>.json, connection resolution
// from the environment, live metadata capture, and a drift check between the
// manifest and the server actually reached.
package target

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/tothandras/ch-eval/migrations/migrate"
)

// Manifest is targets/<platform>/<tier>.json.
type Manifest struct {
	Name            string `json:"name"` // "<platform>/<tier-file>"
	Platform        string `json:"platform"`
	Tier            string `json:"tier"` // local-single | local-cluster | kind | aws | cloud
	Deployment      string `json:"deployment"`
	Image           string `json:"image"`
	ExpectedVersion string `json:"expected_version"`
	Dialect         string `json:"dialect"`
	Cluster         string `json:"cluster"`
	Shards          int    `json:"shards"`
	Replicas        int    `json:"replicas"`
	Storage         string `json:"storage"`
	StoragePolicy   string `json:"storage_policy"`
	Shape           struct {
		VCPUPerReplica *float64 `json:"vcpu_per_replica"`
		GiBPerReplica  *float64 `json:"gib_per_replica"`
	} `json:"shape"`
	Location string `json:"location"`
	DSNEnv   string `json:"dsn_env"`
	Notes    string `json:"notes,omitempty"`
	// Remaining fields are deployment-specific (ports, keeper image, memory).
	Ports *struct {
		NativeBase int `json:"native_base"`
		HTTPBase   int `json:"http_base"`
	} `json:"ports,omitempty"`
	KeeperImage     string `json:"keeper_image,omitempty"`
	Keepers         int    `json:"keepers,omitempty"`
	MemBytesPerNode int64  `json:"mem_bytes_per_node,omitempty"`
}

// LocalTiers can back correctness, parity and smoke checks only (ADR-0005).
var LocalTiers = map[string]bool{"local-single": true, "local-cluster": true, "kind": true}

// IsLocal reports whether results from this target must never back performance or
// cost conclusions.
func (m Manifest) IsLocal() bool { return LocalTiers[m.Tier] }

// Load reads targets/<platform>/<tier>.json. name is "<platform>/<tier>".
func Load(targetsDir, name string) (Manifest, error) {
	parts := strings.SplitN(name, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return Manifest{}, fmt.Errorf("target: name must be <platform>/<tier>, got %q", name)
	}
	path := filepath.Join(targetsDir, parts[0], parts[1]+".json")
	f, err := os.Open(path)
	if err != nil {
		return Manifest{}, fmt.Errorf("target: %w", err)
	}
	defer f.Close()
	dec := json.NewDecoder(f)
	dec.DisallowUnknownFields()
	var m Manifest
	if err := dec.Decode(&m); err != nil {
		return Manifest{}, fmt.Errorf("target %s: %w", path, err)
	}
	if m.Name != name {
		return Manifest{}, fmt.Errorf("target %s: name field %q must equal %q", path, m.Name, name)
	}
	if _, err := m.Topology("x"); err != nil {
		return Manifest{}, fmt.Errorf("target %s: %w", path, err)
	}
	return m, nil
}

// DSN resolves the connection string: the manifest's environment variable first,
// then CLICKHOUSE_DSN, then (local tiers only) the passwordless `default` user on
// the tier's first native port. No connection string is stored in the repository;
// paid tiers resolve only from the environment.
func (m Manifest) DSN() (string, string, error) {
	if m.DSNEnv != "" {
		if v := strings.Trim(os.Getenv(m.DSNEnv), `"`); v != "" {
			return v, m.DSNEnv, nil
		}
	}
	if v := strings.Trim(os.Getenv("CLICKHOUSE_DSN"), `"`); v != "" {
		return v, "CLICKHOUSE_DSN", nil
	}
	if m.IsLocal() && m.Ports != nil && m.Ports.NativeBase > 0 {
		return fmt.Sprintf("clickhouse://default@127.0.0.1:%d/default", m.Ports.NativeBase), "local tier default", nil
	}
	return "", "", fmt.Errorf("target %s: no DSN: set %s (in .envrc)", m.Name, m.DSNEnv)
}

// Topology is the migration topology for a database on this target.
func (m Manifest) Topology(database string) (migrate.Topology, error) {
	t := migrate.Topology{
		Dialect:       migrate.Dialect(m.Dialect),
		Database:      database,
		Cluster:       m.Cluster,
		Shards:        m.Shards,
		StoragePolicy: m.StoragePolicy,
	}
	if t.Dialect != migrate.Replicated {
		t.Cluster = ""
	}
	return t, t.Validate()
}

// SystemCluster is the cluster for clusterAllReplicas reads of system tables, or ""
// to read the connected node only.
func (m Manifest) SystemCluster() string {
	switch m.Dialect {
	case "replicated":
		return m.Cluster
	case "cloud":
		return "default"
	}
	return ""
}

// Package runs persists target-aware runs in the cross-target layout
// results/<workload>/<scenario>/<platform>/<tier>/<run-id>/{run.json,report.md,metadata.json}
// (cross-target-comparison spec) and finds the latest run per target for compare.
package runs

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/tothandras/ch-eval/bench/runner"
	"github.com/tothandras/ch-eval/bench/target"
	"github.com/tothandras/ch-eval/migrations/migrate"
)

// Run is run.json: the forked harness's run schema (flattened, so existing readers
// keep working) plus the target, seed and migration blocks.
type Run struct {
	runner.Run
	RunID      string            `json:"run_id"`
	Workload   string            `json:"workload"`
	Target     TargetBlock       `json:"target"`
	Seed       *SeedBlock        `json:"seed,omitempty"`
	Migrations []migrate.Applied `json:"migrations,omitempty"`
	// SmokeOnly is true for local tiers: the run backs parity and correctness
	// only, never performance or cost conclusions (ADR-0005).
	SmokeOnly bool `json:"smoke_only"`
}

// TargetBlock identifies where the run executed.
type TargetBlock struct {
	Name     string   `json:"name"`
	Platform string   `json:"platform"`
	Tier     string   `json:"tier"`
	Dialect  string   `json:"dialect"`
	Shards   int      `json:"shards"`
	Replicas int      `json:"replicas"`
	Storage  string   `json:"storage"`
	Version  string   `json:"version"`
	Drift    []string `json:"drift,omitempty"`
}

// SeedBlock records how the data was generated and inserted.
type SeedBlock struct {
	Generator     string         `json:"generator"`
	Rows          int            `json:"rows"`
	RNGSeed       uint64         `json:"rng_seed"`
	Namespaces    int            `json:"namespaces"`
	NamespaceSkew float64        `json:"namespace_skew,omitempty"`
	TimeEnd       string         `json:"time_end"`
	Table         string         `json:"table"`
	Settings      map[string]any `json:"settings"`
	IngestPath    string         `json:"ingest_path,omitempty"`
	Profile       string         `json:"profile,omitempty"`
	// TableRows is the expected row count per table the generator wrote.
	TableRows map[string]uint64 `json:"table_rows,omitempty"`
	// CacheAfterSeed is the filesystem-cache occupancy summed over every replica
	// right after seeding (design D14: what the write path put in the cache).
	CacheAfterSeed *CacheSnapshot `json:"cache_after_seed,omitempty"`
	Notes          []string       `json:"notes,omitempty"`
}

// CacheSnapshot is filesystem-cache occupancy across replicas.
type CacheSnapshot struct {
	Segments uint64 `json:"segments"`
	Bytes    uint64 `json:"bytes"`
}

// NewRunID returns <UTC timestamp>-<4 hex>; the same id tags any cloud resources.
func NewRunID(t time.Time) string {
	b := make([]byte, 2)
	_, _ = rand.Read(b)
	return t.UTC().Format("20060102T150405Z") + "-" + hex.EncodeToString(b)
}

// Block builds the target block from a manifest and the captured metadata.
func Block(m target.Manifest, md target.Metadata) TargetBlock {
	return TargetBlock{Name: m.Name, Platform: m.Platform, Tier: m.Tier, Dialect: m.Dialect,
		Shards: m.Shards, Replicas: m.Replicas, Storage: m.Storage, Version: md.Version, Drift: md.Drift}
}

// Dir is the run directory for a run.
func Dir(root string, r Run) string {
	return filepath.Join(root, r.Workload, r.Scenario, filepath.FromSlash(r.Target.Name), r.RunID)
}

// Write persists run.json, report.md and metadata.json and returns the directory.
func Write(root string, r Run, report string, md target.Metadata) (string, error) {
	dir := Dir(root, r)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	if err := writeJSON(filepath.Join(dir, "run.json"), r); err != nil {
		return "", err
	}
	if err := writeJSON(filepath.Join(dir, "metadata.json"), md); err != nil {
		return "", err
	}
	if err := os.WriteFile(filepath.Join(dir, "report.md"), []byte(report), 0o644); err != nil {
		return "", err
	}
	return dir, nil
}

func writeJSON(path string, v any) error {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(b, '\n'), 0o644)
}

// Read loads a run from its directory or run.json path.
func Read(path string) (Run, string, error) {
	if st, err := os.Stat(path); err == nil && st.IsDir() {
		path = filepath.Join(path, "run.json")
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return Run{}, "", err
	}
	var r Run
	if err := json.Unmarshal(b, &r); err != nil {
		return Run{}, "", fmt.Errorf("%s: %w", path, err)
	}
	return r, path, nil
}

// Latest returns the newest valid run (run.json parses and has queries) of a
// scenario on a target. Run ids sort chronologically.
func Latest(root, workload, scenario, targetName string) (Run, string, error) {
	dir := filepath.Join(root, workload, scenario, filepath.FromSlash(targetName))
	entries, err := os.ReadDir(dir)
	if err != nil {
		return Run{}, "", fmt.Errorf("no runs of %s/%s on %s: %w", workload, scenario, targetName, err)
	}
	var ids []string
	for _, e := range entries {
		if e.IsDir() && !strings.HasPrefix(e.Name(), ".") {
			ids = append(ids, e.Name())
		}
	}
	sort.Sort(sort.Reverse(sort.StringSlice(ids)))
	for _, id := range ids {
		path := filepath.Join(dir, id, "run.json")
		b, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		var r Run
		if err := json.Unmarshal(b, &r); err != nil || len(r.Queries) == 0 {
			continue
		}
		return r, path, nil
	}
	return Run{}, "", fmt.Errorf("no valid run of %s/%s on %s under %s", workload, scenario, targetName, dir)
}

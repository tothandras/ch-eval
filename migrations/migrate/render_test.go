package migrate

import (
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

var update = flag.Bool("update", false, "rewrite golden files")

var topologies = map[string]Topology{
	"single":         {Dialect: Single, Database: "openmeter"},
	"replicated-1x2": {Dialect: Replicated, Database: "openmeter", Cluster: "ch_eval", Shards: 1},
	"replicated-2x2": {Dialect: Replicated, Database: "openmeter", Cluster: "ch_eval", Shards: 2},
	"cloud":          {Dialect: Cloud, Database: "openmeter"},
	"single-s3":      {Dialect: Single, Database: "openmeter", StoragePolicy: "s3_cached"},
}

// TestRenderGolden renders the real OpenMeter proposal migration for every topology
// and compares against golden files: one source, no forked files.
func TestRenderGolden(t *testing.T) {
	migs, err := Load("../../scenarios/openmeter/proposal/init")
	if err != nil {
		t.Fatal(err)
	}
	if len(migs) != 1 || migs[0].DownPath == "" {
		t.Fatalf("want one migration with a down file, got %+v", migs)
	}
	for name, topo := range topologies {
		for _, dir := range []Direction{Up, Down} {
			r, err := Render(migs[0], dir, topo)
			if err != nil {
				t.Fatalf("%s %s: %v", name, dir, err)
			}
			if lint := Lint(r.Statements, topo); len(lint) > 0 {
				t.Errorf("%s %s lint: %v", name, dir, lint)
			}
			got := strings.Join(r.Statements, ";\n\n") + ";\n"
			path := filepath.Join("testdata", "proposal."+name+"."+string(dir)+".sql")
			if *update {
				if err := os.WriteFile(path, []byte(got), 0o644); err != nil {
					t.Fatal(err)
				}
				continue
			}
			want, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("missing golden %s (run with -update): %v", path, err)
			}
			if got != string(want) {
				t.Errorf("%s %s: render differs from %s\n--- got ---\n%s", name, dir, path, got)
			}
		}
	}
}

func TestShardedRendersDistributedAndLocal(t *testing.T) {
	migs, _ := Load("../../scenarios/openmeter/proposal/init")
	r, _ := Render(migs[0], Up, topologies["replicated-2x2"])
	if len(r.Statements) != 2 {
		t.Fatalf("sharded render: want local + Distributed statements, got %d", len(r.Statements))
	}
	if !strings.Contains(r.Statements[0], "openmeter.proposal_events_local ON CLUSTER 'ch_eval'") ||
		!strings.Contains(r.Statements[0], "ReplicatedMergeTree('/clickhouse/tables/{shard}/{database}/{table}', '{replica}')") {
		t.Errorf("local table statement wrong:\n%s", r.Statements[0])
	}
	if !strings.Contains(r.Statements[1], "ENGINE = Distributed('ch_eval', 'openmeter', 'proposal_events_local', cityHash64(namespace))") {
		t.Errorf("distributed statement wrong:\n%s", r.Statements[1])
	}
	single, _ := Render(migs[0], Up, topologies["single"])
	if len(single.Statements) != 1 || strings.Contains(single.Statements[0], "ON CLUSTER") || !strings.Contains(single.Statements[0], "ENGINE = MergeTree\n") {
		t.Errorf("single render wrong: %q", single.Statements)
	}
	if single.Checksum == r.Checksum {
		t.Error("checksums must differ across topologies")
	}
}

func TestLint(t *testing.T) {
	topo := topologies["single"]
	bad := []string{
		"CREATE TABLE t (a UInt8) ENGINE = MergeTree ORDER BY a",
		"DROP TABLE t",
		"ALTER TABLE t ADD COLUMN b UInt8",
		"ALTER TABLE t DROP INDEX i",
		"CREATE TABLE IF NOT EXISTS t ON CLUSTER c (a UInt8) ENGINE = MergeTree ORDER BY a",
	}
	for _, s := range bad {
		if len(Lint([]string{s}, topo)) == 0 {
			t.Errorf("lint should reject: %s", s)
		}
	}
	good := []string{
		"CREATE TABLE IF NOT EXISTS t (a UInt8) ENGINE = MergeTree ORDER BY a",
		"CREATE OR REPLACE FUNCTION f AS (x) -> x",
		"CREATE ROW POLICY OR REPLACE p ON t FOR SELECT USING 1 TO r",
		"DROP TABLE IF EXISTS t SYNC",
		"ALTER TABLE t ADD COLUMN IF NOT EXISTS b UInt8, ADD INDEX IF NOT EXISTS i b TYPE minmax",
		"ALTER TABLE t MATERIALIZE INDEX i",
		"INSERT INTO t SELECT 1",
	}
	for _, s := range good {
		if l := Lint([]string{s}, topo); len(l) > 0 {
			t.Errorf("lint should accept %q: %v", s, l)
		}
	}
}

func TestSplitStatements(t *testing.T) {
	in := "-- comment; with semicolon\nCREATE TABLE a (x UInt8) ENGINE = MergeTree ORDER BY x;\n\n  ;\nSELECT 1 -- trailing\n;\n"
	got := SplitStatements(in)
	if len(got) != 2 || !strings.HasPrefix(got[0], "CREATE TABLE a") || !strings.HasPrefix(got[1], "SELECT 1") {
		t.Fatalf("split: %q", got)
	}
}

func TestLoadRejectsBadNames(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "001_init.up.sql"), []byte("SELECT 1"), 0o644)
	if _, err := Load(dir); err == nil {
		t.Fatal("want error for a non-timestamp version")
	}
	dir2 := t.TempDir()
	os.WriteFile(filepath.Join(dir2, "20260101000000_x.down.sql"), []byte("SELECT 1"), 0o644)
	if _, err := Load(dir2); err == nil {
		t.Fatal("want error for a down file without an up file")
	}
}

func TestUnknownHelperFailsAtParse(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "20260101000000_x.up.sql.tmpl"), []byte("SELECT {{ shell \"rm\" }}"), 0o644)
	migs, err := Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Render(migs[0], Up, topologies["single"]); err == nil {
		t.Fatal("unknown helper must fail (closed helper set)")
	}
}

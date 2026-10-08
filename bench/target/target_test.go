package target

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadAllManifests(t *testing.T) {
	root := "../../targets"
	plats, _ := os.ReadDir(root)
	n := 0
	for _, p := range plats {
		if !p.IsDir() || p.Name() == "compose" || p.Name() == "kind" || p.Name() == "aws" || p.Name()[0] == '.' {
			continue
		}
		files, _ := filepath.Glob(filepath.Join(root, p.Name(), "*.json"))
		for _, f := range files {
			name := p.Name() + "/" + filepath.Base(f[:len(f)-5])
			m, err := Load(root, name)
			if err != nil {
				t.Errorf("%s: %v", name, err)
				continue
			}
			n++
			if !m.IsLocal() && m.Ports != nil {
				t.Errorf("%s: paid tier must not carry local ports (it would get a default DSN)", name)
			}
		}
	}
	if n < 9 {
		t.Fatalf("expected at least 9 manifests, loaded %d", n)
	}
}

func TestDSNResolution(t *testing.T) {
	m, err := Load("../../targets", "altinity/local-single")
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("CLICKHOUSE_DSN", "")
	t.Setenv(m.DSNEnv, "")
	dsn, src, err := m.DSN()
	if err != nil || dsn != "clickhouse://default@127.0.0.1:19000/default" || src != "local tier default" {
		t.Fatalf("local default: %q %q %v", dsn, src, err)
	}
	t.Setenv(m.DSNEnv, "clickhouse://x@h:9000/db")
	if dsn, src, _ = m.DSN(); dsn != "clickhouse://x@h:9000/db" || src != m.DSNEnv {
		t.Fatalf("env override: %q %q", dsn, src)
	}
	c, _ := Load("../../targets", "cloud/scale")
	t.Setenv("CLICKHOUSE_DSN", "")
	t.Setenv(c.DSNEnv, "")
	if _, _, err := c.DSN(); err == nil {
		t.Fatal("paid tier without env must fail")
	}
}

func TestDrift(t *testing.T) {
	m := Manifest{ExpectedVersion: "26.3.33.24", Dialect: "replicated", Shards: 1, Replicas: 2, StoragePolicy: "s3_cached"}
	md := Metadata{Version: "26.3.33.24", Replicas: []ReplicaInfo{{Version: "26.3.33.24"}}, Policies: []string{"default"}}
	d := Drift(m, md)
	if len(d) != 2 {
		t.Fatalf("want replica-count and policy drift, got %v", d)
	}
}

package migrate

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"text/template"
)

// Direction of a migration.
type Direction string

const (
	Up   Direction = "up"
	Down Direction = "down"
)

// Migration is one versioned step with its up and (optional) down source.
type Migration struct {
	Version  uint64
	Name     string
	UpPath   string
	DownPath string // "" when no down file exists
}

var fileRE = regexp.MustCompile(`^(\d{14})_([a-z0-9][a-z0-9_\-]*)\.(up|down)\.sql(\.tmpl)?$`)

// Load reads a migration directory. Every file must match the naming convention;
// a stray file is an error so typos cannot silently skip a migration.
func Load(dir string) ([]Migration, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	byVersion := map[uint64]*Migration{}
	for _, e := range entries {
		if e.IsDir() || strings.HasPrefix(e.Name(), ".") || strings.EqualFold(e.Name(), "README.md") {
			continue
		}
		m := fileRE.FindStringSubmatch(e.Name())
		if m == nil {
			return nil, fmt.Errorf("migrate: %s: file name must be YYYYMMDDHHMMSS_name.(up|down).sql[.tmpl]", filepath.Join(dir, e.Name()))
		}
		v, _ := strconv.ParseUint(m[1], 10, 64)
		mig := byVersion[v]
		if mig == nil {
			mig = &Migration{Version: v, Name: m[2]}
			byVersion[v] = mig
		} else if mig.Name != m[2] {
			return nil, fmt.Errorf("migrate: version %d has two names: %q and %q", v, mig.Name, m[2])
		}
		p := filepath.Join(dir, e.Name())
		switch m[3] {
		case "up":
			if mig.UpPath != "" {
				return nil, fmt.Errorf("migrate: duplicate up file for version %d", v)
			}
			mig.UpPath = p
		case "down":
			if mig.DownPath != "" {
				return nil, fmt.Errorf("migrate: duplicate down file for version %d", v)
			}
			mig.DownPath = p
		}
	}
	out := make([]Migration, 0, len(byVersion))
	for _, m := range byVersion {
		if m.UpPath == "" {
			return nil, fmt.Errorf("migrate: version %d (%s) has a down file but no up file", m.Version, m.Name)
		}
		out = append(out, *m)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Version < out[j].Version })
	return out, nil
}

// Rendered is a migration rendered for one topology.
type Rendered struct {
	Statements []string
	Checksum   string // sha256 over the rendered statements
}

// Render renders one direction of a migration for a topology. A missing down file
// renders to zero statements.
func Render(m Migration, dir Direction, topo Topology) (Rendered, error) {
	if err := topo.Validate(); err != nil {
		return Rendered{}, err
	}
	path := m.UpPath
	if dir == Down {
		path = m.DownPath
	}
	if path == "" {
		return Rendered{Checksum: checksum(nil)}, nil
	}
	src, err := os.ReadFile(path)
	if err != nil {
		return Rendered{}, err
	}
	text := string(src)
	if strings.HasSuffix(path, ".tmpl") {
		t, err := template.New(filepath.Base(path)).Option("missingkey=error").Funcs(funcs(topo)).Parse(text)
		if err != nil {
			return Rendered{}, fmt.Errorf("migrate: parse %s: %w", path, err)
		}
		var buf bytes.Buffer
		if err := t.Execute(&buf, nil); err != nil {
			return Rendered{}, fmt.Errorf("migrate: render %s: %w", path, err)
		}
		text = buf.String()
	}
	stmts := SplitStatements(text)
	return Rendered{Statements: stmts, Checksum: checksum(stmts)}, nil
}

func checksum(stmts []string) string {
	h := sha256.New()
	for _, s := range stmts {
		h.Write([]byte(s))
		h.Write([]byte{0})
	}
	return hex.EncodeToString(h.Sum(nil))
}

// SplitStatements removes full-line `--` comments and splits on `;` at end of line.
// Statements in migration files must not contain `;` followed by a newline inside
// string literals. Empty statements (for example a helper that rendered to nothing)
// are dropped.
func SplitStatements(text string) []string {
	var lines []string
	for _, l := range strings.Split(text, "\n") {
		if strings.HasPrefix(strings.TrimSpace(l), "--") {
			continue
		}
		lines = append(lines, l)
	}
	joined := strings.Join(lines, "\n") + "\n"
	parts := regexp.MustCompile(`;[ \t]*\n`).Split(joined, -1)
	var out []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		p = strings.TrimSuffix(p, ";")
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

// funcs is the closed helper set available to templates (ADR-0003). Anything else
// fails at parse time.
func funcs(t Topology) template.FuncMap {
	return template.FuncMap{
		// db is the target database.
		"db": func() string { return t.Database },
		// cluster is the remote_servers cluster ("" outside the replicated dialect).
		"cluster": func() string { return t.Cluster },
		// dialect reports whether the target uses the named dialect.
		"dialect": func(d string) bool { return string(t.Dialect) == d },
		// sharded reports whether tables are split into local + Distributed tables.
		"sharded": func() bool { return t.Sharded() },
		// onCluster renders the ON CLUSTER clause where DDL must fan out.
		"onCluster": func() string {
			if t.Dialect == Replicated {
				return fmt.Sprintf("ON CLUSTER '%s'", t.Cluster)
			}
			return ""
		},
		// table is the physical (shard-local when sharded) name of a logical table.
		"table": func(name string) string { return physical(t, name) },
		// engine renders a MergeTree-family engine; on the replicated dialect it becomes
		// Replicated* with a per-shard Keeper path. Extra args are engine parameters
		// (e.g. the ReplacingMergeTree version column).
		"engine": func(name string, args ...string) string { return engine(t, name, false, args) },
		// globalEngine is like engine but replicates across ALL shards (one Keeper path
		// for the whole cluster): for small state/lookup tables that must be identical
		// everywhere, such as the migrations table.
		"globalEngine": func(name string, args ...string) string { return engine(t, name, true, args) },
		// distributed renders the Distributed table over a sharded logical table, or
		// nothing when the target is not sharded.
		"distributed": func(name, key string) string {
			if !t.Sharded() {
				return ""
			}
			return fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s.%s ON CLUSTER '%s' AS %s.%s ENGINE = Distributed('%s', '%s', '%s', %s)",
				t.Database, name, t.Cluster, t.Database, physical(t, name), t.Cluster, t.Database, physical(t, name), key)
		},
		// settings renders a SETTINGS clause from a comma-separated list, appending the
		// target's storage policy when one is configured; renders nothing when both are empty.
		"settings": func(list ...string) string {
			var items []string
			for _, l := range list {
				if strings.TrimSpace(l) != "" {
					items = append(items, strings.TrimSpace(l))
				}
			}
			if t.StoragePolicy != "" {
				items = append(items, fmt.Sprintf("storage_policy = '%s'", t.StoragePolicy))
			}
			if len(items) == 0 {
				return ""
			}
			return "SETTINGS " + strings.Join(items, ", ")
		},
		// dropDistributed renders the matching DROP for down migrations.
		"dropDistributed": func(name string) string {
			if !t.Sharded() {
				return ""
			}
			return fmt.Sprintf("DROP TABLE IF EXISTS %s.%s ON CLUSTER '%s' SYNC", t.Database, name, t.Cluster)
		},
	}
}

func physical(t Topology, name string) string {
	if t.Sharded() {
		return name + "_local"
	}
	return name
}

func engine(t Topology, name string, global bool, args []string) string {
	if !strings.HasSuffix(name, "MergeTree") || strings.HasPrefix(name, "Replicated") || strings.HasPrefix(name, "Shared") {
		// Only plain MergeTree-family names are templated; anything else is a mistake.
		return fmt.Sprintf("/* invalid engine %q: use the plain MergeTree-family name */ %s", name, name)
	}
	switch t.Dialect {
	case Replicated:
		shard := "{shard}"
		if global {
			shard = "all"
		}
		params := append([]string{fmt.Sprintf("'/clickhouse/tables/%s/{database}/{table}'", shard), "'{replica}'"}, args...)
		return fmt.Sprintf("Replicated%s(%s)", name, strings.Join(params, ", "))
	default: // single, cloud (Cloud substitutes Shared* for plain names)
		if len(args) == 0 {
			return name
		}
		return fmt.Sprintf("%s(%s)", name, strings.Join(args, ", "))
	}
}

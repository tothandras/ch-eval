// Package probe runs feature probes against a target and renders the feature
// matrix from docs/feature-matrix.yaml plus the latest probe results
// (feature-probes spec).
//
// A SQL probe is a file under probes/ with header comments:
//
//	-- probe: <title>
//	-- expect: <observation>            default for every target
//	-- expect.<selector>: <observation> override; selectors, most specific first:
//	                                    tier-<tier>, <platform>, storage-<storage>, dialect-<dialect>
//	-- requires: cluster | replicated | sharded   (skipped elsewhere, recorded as skipped)
//	-- cleanup: <statement>             run after the probe whatever happened
//
// followed by statements separated by ';'. Every statement but the last is setup;
// the last one's outcome is the observation: "ok", "value:<first row, tab-joined>",
// or "error:<code>". Expectations are "ok" (no error), "value:<exact>",
// "regex:<re>", "error" (any error) or "error:<code>", alternatives joined by " || ".
// Placeholders: {db} scratch database, {cluster} the target's cluster, {nodes}
// reachable node count.
package probe

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"

	"github.com/tothandras/ch-eval/bench/target"
	"github.com/tothandras/ch-eval/migrations/migrate"
)

// Probe is one parsed SQL probe or a registered Go probe.
type Probe struct {
	Path       string
	Title      string
	Expect     map[string]string // selector ("" = default) -> expectation
	Requires   []string
	Cleanup    []string
	Statements []string
	Go         GoProbe
}

// GoProbe implements a probe that SQL cannot express (client behaviour, TLS).
type GoProbe func(ctx context.Context, env Env) Observation

// Env is what a probe runs against.
type Env struct {
	Conn     driver.Conn
	Options  *clickhouse.Options
	Manifest target.Manifest
	DB       string // scratch database
	Nodes    int
}

var goProbes = map[string]GoProbe{}

func registerGo(path, title string, expect map[string]string, f GoProbe) {
	goProbes[path] = f
	goMeta[path] = Probe{Path: path, Title: title, Expect: expect, Go: f}
}

var goMeta = map[string]Probe{}

// Load reads every probes/**/*.sql file plus the registered Go probes.
func Load(root string) ([]Probe, error) {
	var out []Probe
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(path, ".sql") {
			return err
		}
		p, err := parse(path)
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(filepath.Dir(root), path)
		p.Path = filepath.ToSlash(rel)
		out = append(out, p)
		return nil
	})
	if err != nil {
		return nil, err
	}
	for _, p := range goMeta {
		out = append(out, p)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Path < out[j].Path })
	return out, nil
}

func parse(path string) (Probe, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return Probe{}, err
	}
	p := Probe{Expect: map[string]string{}}
	var body []string
	for _, line := range strings.Split(string(b), "\n") {
		t := strings.TrimSpace(line)
		if strings.HasPrefix(t, "--") {
			kv := strings.TrimSpace(strings.TrimPrefix(t, "--"))
			key, val, ok := strings.Cut(kv, ":")
			if ok && !strings.Contains(key, " ") {
				val = strings.TrimSpace(val)
				switch {
				case key == "probe":
					p.Title = val
				case key == "expect":
					p.Expect[""] = val
				case strings.HasPrefix(key, "expect."):
					p.Expect[strings.TrimPrefix(key, "expect.")] = val
				case key == "requires":
					p.Requires = append(p.Requires, strings.Fields(strings.ReplaceAll(val, ",", " "))...)
				case key == "cleanup":
					p.Cleanup = append(p.Cleanup, val)
				}
			}
			continue
		}
		body = append(body, line)
	}
	p.Statements = migrate.SplitStatements(strings.Join(body, "\n"))
	if p.Title == "" || p.Expect[""] == "" && len(p.Expect) == 0 || len(p.Statements) == 0 {
		return p, fmt.Errorf("%s: needs -- probe:, at least one -- expect and a statement", path)
	}
	return p, nil
}

// Expectation picks the expectation for a target.
func (p Probe) Expectation(m target.Manifest) string {
	for _, sel := range []string{"tier-" + m.Tier, m.Platform, "storage-" + m.Storage, "dialect-" + m.Dialect, ""} {
		if e, ok := p.Expect[sel]; ok {
			return e
		}
	}
	return ""
}

// Skip reports why a probe does not apply to a target, or "".
func (p Probe) Skip(m target.Manifest) string {
	for _, r := range p.Requires {
		switch r {
		case "cluster":
			if m.SystemCluster() == "" {
				return "needs a cluster"
			}
		case "replicated":
			if m.Dialect == "single" {
				return "needs replicated tables (Keeper)"
			}
		case "sharded":
			if !(m.Dialect == "replicated" && m.Shards > 1) {
				return "needs a sharded topology"
			}
		}
	}
	return ""
}

func (e Env) render(s string) string {
	r := strings.NewReplacer("{db}", e.DB, "{cluster}", e.Manifest.SystemCluster(), "{nodes}", fmt.Sprint(e.Nodes))
	return r.Replace(s)
}

var selectRe = regexp.MustCompile(`(?is)^\s*(SELECT|WITH)\b`)

func firstRowTSV(ctx context.Context, env Env, sql string, i int) Observation {
	rows, err := env.Conn.Query(ctx, "SELECT formatRow('TabSeparated', *) FROM ("+sql+")")
	if err != nil {
		return errObs(err, i)
	}
	defer rows.Close()
	var v string
	if rows.Next() {
		if err := rows.Scan(&v); err != nil {
			return errObs(err, i)
		}
		return Observation{Kind: "value", Value: strings.TrimSuffix(v, "\n")}
	}
	if err := rows.Err(); err != nil {
		return errObs(err, i)
	}
	return Observation{Kind: "ok", Message: "no rows"}
}

var queryRe = regexp.MustCompile(`(?is)^\s*(SELECT|WITH|SHOW|DESCRIBE|DESC|EXISTS|EXPLAIN)\b`)

// Run executes one probe.
func Run(ctx context.Context, env Env, p Probe) Result {
	start := time.Now()
	res := Result{Path: p.Path, Title: p.Title, Expect: env.render(p.Expectation(env.Manifest))}
	if why := p.Skip(env.Manifest); why != "" {
		res.Observed = Observation{Kind: "skipped", Message: why}
		res.Pass = true
		return res
	}
	if p.Go != nil {
		res.Observed = p.Go(ctx, env)
	} else {
		res.Observed = runSQL(ctx, env, p.Statements)
	}
	for _, c := range p.Cleanup {
		_ = env.Conn.Exec(ctx, env.render(c))
	}
	res.Pass = Matches(res.Expect, res.Observed)
	res.DurationMs = float64(time.Since(start).Microseconds()) / 1000
	return res
}

func runSQL(ctx context.Context, env Env, stmts []string) Observation {
	for i, s := range stmts {
		sql := env.render(s)
		last := i == len(stmts)-1
		if queryRe.MatchString(sql) {
			if last && selectRe.MatchString(sql) {
				// Let the server render the row (TabSeparated), so observations read
				// exactly as ClickHouse prints them on every target and version.
				return firstRowTSV(ctx, env, sql, i)
			}
			rows, err := env.Conn.Query(ctx, sql)
			if err != nil {
				return errObs(err, i)
			}
			var first []string
			for rows.Next() {
				if first != nil {
					continue
				}
				cols := rows.ColumnTypes()
				ptrs := make([]any, len(cols))
				for k, c := range cols {
					ptrs[k] = reflect.New(c.ScanType()).Interface()
				}
				if err := rows.Scan(ptrs...); err != nil {
					rows.Close()
					return errObs(err, i)
				}
				first = make([]string, len(cols))
				for k, p := range ptrs {
					first[k] = fmt.Sprint(reflect.ValueOf(p).Elem().Interface())
				}
			}
			err = rows.Err()
			rows.Close()
			if err != nil {
				return errObs(err, i)
			}
			if last {
				if first == nil {
					return Observation{Kind: "ok", Message: "no rows"}
				}
				return Observation{Kind: "value", Value: strings.Join(first, "\t")}
			}
			continue
		}
		if err := env.Conn.Exec(ctx, sql); err != nil {
			return errObs(err, i)
		}
	}
	return Observation{Kind: "ok"}
}

func errObs(err error, stmt int) Observation {
	var ex *clickhouse.Exception
	if errors.As(err, &ex) {
		msg := ex.Message
		if len(msg) > 300 {
			msg = msg[:300]
		}
		return Observation{Kind: "error", ErrorCode: ex.Code, Message: fmt.Sprintf("statement %d: %s", stmt+1, msg)}
	}
	return Observation{Kind: "error", ErrorCode: -1, Message: fmt.Sprintf("statement %d: %v", stmt+1, err)}
}

// Matches reports whether an observation satisfies an expectation.
func Matches(expect string, o Observation) bool {
	if o.Kind == "skipped" {
		return true
	}
	for _, alt := range strings.Split(expect, "||") {
		alt = strings.TrimSpace(alt)
		kind, arg, _ := strings.Cut(alt, ":")
		switch kind {
		case "ok":
			if o.Kind != "error" {
				return true
			}
		case "value":
			if o.Kind == "value" && o.Value == arg {
				return true
			}
		case "regex":
			if o.Kind == "value" {
				if re, err := regexp.Compile(arg); err == nil && re.MatchString(o.Value) {
					return true
				}
			}
		case "error":
			if o.Kind == "error" && (arg == "" || arg == fmt.Sprint(o.ErrorCode)) {
				return true
			}
		}
	}
	return false
}

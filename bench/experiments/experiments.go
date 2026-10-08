// Package experiments implements the local correctness experiments of design D13
// (replication and sharding): S3 unavailable shard, S4 insert routing and
// time-to-queryable, R6 DDL with a replica down, R3 Keeper quorum loss. They stop
// and start containers, so they run on local tiers; results describe behaviour
// (errors, partial results, catch-up), never performance.
package experiments

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"

	"github.com/tothandras/ch-eval/bench/target"
)

// Env is a running multi-node local target.
type Env struct {
	M       target.Manifest
	Opts    *clickhouse.Options
	Control Control
	Logf    func(string, ...any)
}

// Control stops and starts the target's processes.
type Control interface {
	Nodes() []string
	StopNode(i int) error
	StartNode(i int) error
	StopKeeper(k int) error
	StartKeeper(k int) error
}

// Compose controls a compose deployment by container name.
type Compose struct {
	Project string
	nodes   []string
}

// NewCompose reads the node list written by targets/compose/gen.py.
func NewCompose(m target.Manifest) (*Compose, error) {
	b, err := os.ReadFile(filepath.Join("targets", ".state", m.Platform+"-"+m.Deployment, "nodes"))
	if err != nil {
		return nil, err
	}
	return &Compose{Project: "cheval-" + m.Platform + "-" + m.Deployment, nodes: strings.Fields(string(b))}, nil
}

func (c *Compose) Nodes() []string { return c.nodes }
func (c *Compose) docker(args ...string) error {
	out, err := exec.Command("docker", args...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("docker %v: %v: %s", args, err, out)
	}
	return nil
}
func (c *Compose) StopNode(i int) error {
	return c.docker("stop", "-t", "5", c.Project+"-"+c.nodes[i]+"-1")
}
func (c *Compose) StartNode(i int) error { return c.docker("start", c.Project+"-"+c.nodes[i]+"-1") }
func (c *Compose) StopKeeper(k int) error {
	return c.docker("stop", "-t", "5", fmt.Sprintf("%s-keeper%d-1", c.Project, k))
}
func (c *Compose) StartKeeper(k int) error {
	return c.docker("start", fmt.Sprintf("%s-keeper%d-1", c.Project, k))
}

// node opens a connection to node i (native_base + i).
func (e Env) node(ctx context.Context, i int) (driver.Conn, error) {
	host, _, _ := net.SplitHostPort(e.Opts.Addr[0])
	o := *e.Opts
	o.Addr = []string{net.JoinHostPort(host, strconv.Itoa(e.M.Ports.NativeBase+i))}
	o.DialTimeout = 2 * time.Second
	c, err := clickhouse.Open(&o)
	if err != nil {
		return nil, err
	}
	return c, c.Ping(ctx)
}

// waitNode polls until node i answers, returning how long it took.
func (e Env) waitNode(ctx context.Context, i int, timeout time.Duration) (time.Duration, error) {
	start := time.Now()
	for time.Since(start) < timeout {
		if c, err := e.node(ctx, i); err == nil {
			c.Close()
			return time.Since(start), nil
		}
		time.Sleep(500 * time.Millisecond)
	}
	return time.Since(start), fmt.Errorf("node %d not back after %s", i, timeout)
}

// Step is one observed action.
type Step struct {
	Name    string  `json:"name"`
	Outcome string  `json:"outcome"` // ok | error:<code> | value:<v>
	Detail  string  `json:"detail,omitempty"`
	Millis  float64 `json:"ms"`
}

// Result is one experiment's record.
type Result struct {
	ID        string    `json:"id"`
	Question  string    `json:"question"`
	Target    string    `json:"target"`
	Version   string    `json:"version"`
	StartedAt time.Time `json:"started_at"`
	Steps     []Step    `json:"steps"`
	Findings  []string  `json:"findings"`
}

func (r *Result) step(name string, start time.Time, err error, value string) Step {
	s := Step{Name: name, Millis: float64(time.Since(start).Microseconds()) / 1000}
	switch {
	case err != nil:
		s.Outcome = outcome(err)
		s.Detail = errDetail(err)
	case value != "":
		s.Outcome = "value:" + value
	default:
		s.Outcome = "ok"
	}
	r.Steps = append(r.Steps, s)
	return s
}

func outcome(err error) string {
	var ex *clickhouse.Exception
	if errors.As(err, &ex) {
		return fmt.Sprintf("error:%d", ex.Code)
	}
	return "error"
}

func errDetail(err error) string {
	s := err.Error()
	if i := strings.Index(s, "Stack trace"); i > 0 {
		s = s[:i]
	}
	if len(s) > 240 {
		s = s[:240] + "…"
	}
	return s
}

func count(ctx context.Context, c driver.Conn, q string, settings clickhouse.Settings) (uint64, error) {
	if settings != nil {
		ctx = clickhouse.Context(ctx, clickhouse.WithSettings(settings))
	}
	var n uint64
	err := c.QueryRow(ctx, q).Scan(&n)
	return n, err
}

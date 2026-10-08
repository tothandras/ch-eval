package otel

import (
	"fmt"
	"math"
	"math/rand/v2"
	"strings"
)

// GraphConfig shapes the synthetic system that emits telemetry: services, their
// instances and operations, and who calls whom. Traces, logs and metrics share one
// graph, so logs can correlate with traces and metric series match span routes.
type GraphConfig struct {
	Seed           uint64
	Services       int // default 30
	OpsMin, OpsMax int // operations per service, default 5..40
	PodsPerService int // default 3
	MaxFanout      int // max downstream calls per operation, default 3
	Users          int // user.id cardinality (a high-cardinality attribute), default 100_000
	ErrorRateMax   float64
}

// DefaultGraph is the default graph configuration.
func DefaultGraph(seed uint64) GraphConfig {
	return GraphConfig{Seed: seed, Services: 30, OpsMin: 5, OpsMax: 40, PodsPerService: 3, MaxFanout: 3, Users: 100_000, ErrorRateMax: 0.05}
}

var serviceNames = []string{
	"frontend", "api-gateway", "auth", "cart", "checkout", "payment", "inventory", "catalog",
	"search", "recommendation", "shipping", "email", "notification", "ledger", "pricing",
	"currency", "user-profile", "session", "ad", "review", "quote", "fraud-detection",
	"accounting", "reporting", "media", "upload", "thumbnail", "geo", "billing", "metering",
}

var nouns = []string{
	"items", "orders", "users", "carts", "products", "payments", "invoices", "sessions",
	"accounts", "reviews", "shipments", "quotes", "reports", "images", "addresses", "tokens",
	"prices", "events", "messages", "plans",
}

var langs = []string{"go", "java", "python", "nodejs", "dotnet", "rust"}

// Instrumentation scopes per language (open-source instrumentation library names).
var traceScopes = map[string][2]string{
	"go":     {"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp", "0.63.0"},
	"java":   {"io.opentelemetry.spring-webmvc-6.0", "2.20.0-alpha"},
	"python": {"opentelemetry.instrumentation.flask", "0.58b0"},
	"nodejs": {"@opentelemetry/instrumentation-http", "0.205.0"},
	"dotnet": {"OpenTelemetry.Instrumentation.AspNetCore", "1.12.0"},
	"rust":   {"tracing-opentelemetry", "0.31.0"},
}

var logScopes = map[string][2]string{
	"go":     {"go.opentelemetry.io/contrib/bridges/otelslog", "0.13.0"},
	"java":   {"io.opentelemetry.logback-appender-1.0", "2.20.0-alpha"},
	"python": {"opentelemetry.sdk._logs", "1.37.0"},
	"nodejs": {"@opentelemetry/winston-transport", "0.16.0"},
	"dotnet": {"OpenTelemetry.Logs", "1.12.0"},
	"rust":   {"opentelemetry-appender-tracing", "0.31.0"},
}

var dbSystems = []string{"postgresql", "mysql", "redis", "mongodb"}

const (
	callService = iota
	callDB
	callInternal
	callPublish
)

type call struct {
	kind   int
	svc    int // callService: callee service
	op     int // callService: callee entry op
	name   string
	prob   float64 // probability the call happens in a given trace
	baseMs float64 // callDB/callInternal/callPublish self time
}

type operation struct {
	name      string // server span name
	route     string // http route, or rpc method
	method    string // http method, or rpc service
	mu, sigma float64
	errRate   float64
	calls     []call
}

type service struct {
	name, lang, version, proto string
	pods                       []string
	nodes                      []string
	ops                        []operation
	db                         string // db system, "" if none
	dbName                     string
	topic                      string
}

// Graph is the generated system. It is deterministic in GraphConfig.
type Graph struct {
	cfg        GraphConfig
	svcs       []service
	resources  []Resource // one per (service, pod)
	scopes     []Scope    // trace scope per service
	logScopes  []Scope    // log scope per service
	edge       []int      // services that receive external traffic
	edgeWeight []float64
}

// NewGraph builds the graph for a configuration.
func NewGraph(cfg GraphConfig) *Graph {
	if cfg.Services <= 0 {
		cfg.Services = 30
	}
	if cfg.OpsMin <= 0 {
		cfg.OpsMin = 5
	}
	if cfg.OpsMax < cfg.OpsMin {
		cfg.OpsMax = cfg.OpsMin
	}
	if cfg.PodsPerService <= 0 {
		cfg.PodsPerService = 3
	}
	if cfg.Users <= 0 {
		cfg.Users = 100_000
	}
	r := newRNG(cfg.Seed, streamGraph, 0)
	g := &Graph{cfg: cfg}
	for i := 0; i < cfg.Services; i++ {
		name := serviceNames[i%len(serviceNames)]
		if i >= len(serviceNames) {
			name = fmt.Sprintf("%s-%d", name, i/len(serviceNames))
		}
		s := service{name: name, lang: pick(r, langs), version: fmt.Sprintf("1.%d.%d", r.IntN(20), r.IntN(10)), proto: "http"}
		if i >= 2 && r.Float64() < 0.4 {
			s.proto = "grpc"
		}
		for p := 0; p < cfg.PodsPerService; p++ {
			s.pods = append(s.pods, fmt.Sprintf("%s-%07x-%05x", name, r.Uint32()&0xfffffff, r.Uint32()&0xfffff))
			s.nodes = append(s.nodes, fmt.Sprintf("node-%02d", r.IntN(24)))
		}
		if r.Float64() < 0.6 {
			s.db = pick(r, dbSystems)
			s.dbName = strings.ReplaceAll(name, "-", "_")
		}
		if r.Float64() < 0.3 {
			s.topic = name + "." + pick(r, nouns)
		}
		nOps := cfg.OpsMin + r.IntN(cfg.OpsMax-cfg.OpsMin+1)
		seen := map[string]bool{}
		for o := 0; len(s.ops) < nOps && o < nOps*4; o++ {
			noun := pick(r, nouns)
			var op operation
			if s.proto == "http" {
				method := pick(r, []string{"GET", "GET", "GET", "POST", "PUT", "DELETE"})
				route := "/api/v1/" + noun
				if r.IntN(2) == 0 {
					route += "/{id}"
				}
				op = operation{name: method + " " + route, route: route, method: method}
			} else {
				svcName := "shop." + camel(name) + "Service"
				verb := pick(r, []string{"Get", "List", "Create", "Update", "Delete"})
				m := verb + camel(strings.TrimSuffix(noun, "s"))
				op = operation{name: svcName + "/" + m, route: m, method: svcName}
			}
			if seen[op.name] {
				continue
			}
			seen[op.name] = true
			op.mu = math.Log(1 + 30*r.Float64())
			op.sigma = 0.3 + 0.6*r.Float64()
			op.errRate = cfg.ErrorRateMax * r.Float64() * r.Float64()
			s.ops = append(s.ops, op)
		}
		g.svcs = append(g.svcs, s)
	}
	// Calls follow service order (a DAG: service i only calls j > i), and the
	// deeper two thirds of the services make no downstream calls.
	for i := range g.svcs {
		s := &g.svcs[i]
		for o := range s.ops {
			op := &s.ops[o]
			if i < len(g.svcs)*2/3 && i+1 < len(g.svcs) {
				for c := r.IntN(cfg.MaxFanout + 1); c > 0; c-- {
					j := i + 1 + r.IntN(min(len(g.svcs)-i-1, 8))
					op.calls = append(op.calls, call{kind: callService, svc: j, op: r.IntN(len(g.svcs[j].ops)), prob: 0.6 + 0.4*r.Float64()})
				}
			}
			if s.db != "" && r.Float64() < 0.6 {
				for c := 1 + r.IntN(2); c > 0; c-- {
					verb := pick(r, []string{"SELECT", "SELECT", "INSERT", "UPDATE"})
					op.calls = append(op.calls, call{kind: callDB, name: verb + " " + s.dbName + "." + pick(r, nouns), prob: 0.9, baseMs: 0.5 + 8*r.Float64()})
				}
			}
			if r.Float64() < 0.3 {
				op.calls = append(op.calls, call{kind: callInternal, name: strings.ReplaceAll(s.name, "-", "_") + "." + pick(r, []string{"validate", "transform", "authorize", "enrich", "render"}) + camel(pick(r, nouns)), prob: 0.95, baseMs: 0.2 + 5*r.Float64()})
			}
			if s.topic != "" && r.Float64() < 0.2 {
				op.calls = append(op.calls, call{kind: callPublish, name: s.topic + " publish", prob: 0.8, baseMs: 0.3 + 2*r.Float64()})
			}
		}
	}
	for i, s := range g.svcs {
		ts, lsc := traceScopes[s.lang], logScopes[s.lang]
		g.scopes = append(g.scopes, Scope{Name: ts[0], Version: ts[1]})
		g.logScopes = append(g.logScopes, Scope{Name: lsc[0], Version: lsc[1]})
		for p := range s.pods {
			g.resources = append(g.resources, Resource{
				SchemaURL: "https://opentelemetry.io/schemas/1.37.0",
				Attrs: []Attr{
					S("deployment.environment.name", "bench"),
					S("host.name", s.pods[p]),
					S("k8s.cluster.name", "bench-cluster"),
					S("k8s.deployment.name", s.name),
					S("k8s.namespace.name", "shop"),
					S("k8s.node.name", s.nodes[p]),
					S("k8s.pod.name", s.pods[p]),
					S("service.name", s.name),
					S("service.namespace", "shop"),
					S("service.version", s.version),
					S("telemetry.sdk.language", s.lang),
					S("telemetry.sdk.name", "opentelemetry"),
				},
			})
		}
		if i < 2 {
			g.edge = append(g.edge, i)
			g.edgeWeight = append(g.edgeWeight, float64(2-i))
		}
	}
	return g
}

// Resources lists every emitting resource (service instance).
func (g *Graph) Resources() []Resource { return g.resources }

// Scopes lists the trace instrumentation scope of each service.
func (g *Graph) Scopes() []Scope { return g.scopes }

// LogScopes lists the logging scope of each service.
func (g *Graph) LogScopes() []Scope { return g.logScopes }

// ServiceNames lists the service names in graph order.
func (g *Graph) ServiceNames() []string {
	out := make([]string, len(g.svcs))
	for i, s := range g.svcs {
		out[i] = s.name
	}
	return out
}

func (g *Graph) resource(svc, pod int) int { return svc*g.cfg.PodsPerService + pod }

func (g *Graph) pickEdge(r *rand.Rand) int {
	total := 0.0
	for _, w := range g.edgeWeight {
		total += w
	}
	x := r.Float64() * total
	for i, w := range g.edgeWeight {
		if x < w {
			return g.edge[i]
		}
		x -= w
	}
	return g.edge[len(g.edge)-1]
}

func camel(s string) string {
	parts := strings.FieldsFunc(s, func(r rune) bool { return r == '-' || r == '_' || r == '.' })
	for i, p := range parts {
		if p != "" {
			parts[i] = strings.ToUpper(p[:1]) + p[1:]
		}
	}
	return strings.Join(parts, "")
}

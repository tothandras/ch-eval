package otel

import (
	"fmt"
	"math/rand/v2"
	"time"
)

// TracesConfig controls trace generation.
type TracesConfig struct {
	Graph    GraphConfig
	TimeEnd  time.Time     // newest trace start; traces start uniformly in [TimeEnd-TimeSpan, TimeEnd)
	TimeSpan time.Duration // default 72h
	MaxDepth int           // default 6
	MaxSpans int           // per-trace cap, default 256
}

// TraceGen generates traces as a pure function of (seed, trace index).
type TraceGen struct {
	cfg TracesConfig
	g   *Graph
}

// NewTraceGen builds a generator; the graph is shared with logs and metrics.
func NewTraceGen(cfg TracesConfig, g *Graph) *TraceGen {
	if cfg.TimeSpan <= 0 {
		cfg.TimeSpan = 72 * time.Hour
	}
	if cfg.MaxDepth <= 0 {
		cfg.MaxDepth = 6
	}
	if cfg.MaxSpans <= 0 {
		cfg.MaxSpans = 256
	}
	return &TraceGen{cfg: cfg, g: g}
}

// Graph returns the shared graph.
func (t *TraceGen) Graph() *Graph { return t.g }

// Window is the span of trace start times.
func (t *TraceGen) Window() (time.Time, time.Time) {
	return t.cfg.TimeEnd.Add(-t.cfg.TimeSpan), t.cfg.TimeEnd
}

type traceBuild struct {
	r     *rand.Rand
	seed  uint64
	ti    uint64
	id    [16]byte
	spans []Span
	max   int
}

func (b *traceBuild) newSpan(parent int, s Span) int {
	s.TraceID = b.id
	s.SpanID = SpanID(b.seed, b.ti, uint64(len(b.spans)))
	if parent >= 0 {
		s.ParentSpanID = b.spans[parent].SpanID
	}
	b.spans = append(b.spans, s)
	return len(b.spans) - 1
}

// Trace returns the spans of trace i, root first.
func (t *TraceGen) Trace(i uint64) []Span {
	seed := t.cfg.Graph.Seed
	r := newRNG(seed, streamTrace, i)
	b := &traceBuild{r: r, seed: seed, ti: i, id: TraceID(seed, i), max: t.cfg.MaxSpans}
	from := t.cfg.TimeEnd.Add(-t.cfg.TimeSpan)
	start := uint64(from.UnixNano()) + uint64(r.Int64N(int64(t.cfg.TimeSpan)))
	svc := t.g.pickEdge(r)
	op := r.IntN(len(t.g.svcs[svc].ops))
	user := r.IntN(t.g.cfg.Users)
	root := t.server(b, -1, svc, op, start, 0, user)
	if i > 0 && r.Float64() < 0.01 {
		b.spans[root].Links = []Link{{TraceID: TraceID(seed, i-1), SpanID: SpanID(seed, i-1, 0), Attrs: []Attr{S("link.type", "follows_from")}}}
	}
	return b.spans
}

// server emits the server span of (svc, op) and its subtree; returns the span index.
func (t *TraceGen) server(b *traceBuild, parent, svc, op int, start uint64, depth, user int) int {
	r := b.r
	s := &t.g.svcs[svc]
	o := &s.ops[op]
	pod := r.IntN(len(s.pods))
	idx := b.newSpan(parent, Span{Name: o.name, Kind: KindServer, StartNs: start, Resource: t.g.resource(svc, pod), Scope: svc})
	self := uint64(logNormal(r, o.mu, o.sigma) * 1e6)
	cursor := start + self*3/10
	childErr := false
	if depth < t.cfg.MaxDepth {
		for _, c := range o.calls {
			if len(b.spans) >= b.max || r.Float64() >= c.prob {
				continue
			}
			var end uint64
			var failed bool
			switch c.kind {
			case callService:
				end, failed = t.client(b, idx, svc, pod, c, cursor, depth, user)
			default:
				end, failed = t.leaf(b, idx, svc, pod, c, cursor)
			}
			childErr = childErr || failed
			cursor = end + uint64(r.IntN(200_000))
		}
	}
	end := cursor + self*7/10
	sp := &b.spans[idx]
	sp.EndNs = end
	failed := r.Float64() < o.errRate
	origin := failed
	if childErr && r.Float64() < 0.5 {
		failed = true
	}
	status := 200
	if s.proto == "http" {
		switch {
		case failed:
			status = pick(r, []int{500, 500, 502, 503})
		case o.method == "POST":
			status = 201
		case o.method == "DELETE":
			status = 204
		case r.Float64() < 0.03:
			status = 404
		}
		path := o.route
		if len(path) > 5 && path[len(path)-5:] == "/{id}" {
			path = path[:len(path)-5] + fmt.Sprintf("/%d", r.IntN(1_000_000))
		}
		sp.Attrs = []Attr{
			S("http.request.method", o.method),
			I("http.response.status_code", int64(status)),
			S("http.route", o.route),
			S("url.path", path),
			S("url.scheme", "https"),
		}
		if parent < 0 {
			sp.Attrs = append(sp.Attrs,
				S("client.address", fmt.Sprintf("203.0.113.%d", r.IntN(254)+1)),
				S("server.address", "shop.example.com"),
				S("user.id", fmt.Sprintf("user-%d", user)),
				S("user_agent.original", pick(r, []string{"Mozilla/5.0 (X11; Linux x86_64)", "Mozilla/5.0 (Macintosh)", "okhttp/4.12.0", "curl/8.9.1"})))
		}
	} else {
		code := int64(0)
		if failed {
			code = pick(r, []int64{2, 13, 14})
		}
		sp.Attrs = []Attr{
			I("rpc.grpc.status_code", code),
			S("rpc.method", o.route),
			S("rpc.service", o.method),
			S("rpc.system", "grpc"),
		}
	}
	if failed {
		sp.Status = StatusError
		if origin {
			sp.StatusMessage = pick(r, errMessages)
			sp.Events = []Event{exception(r, end-self/10)}
		}
	}
	return idx
}

var errMessages = []string{"connection reset by peer", "context deadline exceeded", "invalid argument", "permission denied", "resource exhausted", "internal error"}

func exception(r *rand.Rand, at uint64) Event {
	typ := pick(r, []string{"TimeoutError", "ConnectionError", "ValueError", "PermissionError", "RuntimeError"})
	return Event{TimeNs: at, Name: "exception", Attrs: []Attr{S("exception.message", pick(r, errMessages)), S("exception.type", typ)}}
}

// client emits the client span for a downstream call and the callee's server span.
func (t *TraceGen) client(b *traceBuild, parent, svc, pod int, c call, start uint64, depth, user int) (uint64, bool) {
	r := b.r
	callee := &t.g.svcs[c.svc]
	o := &callee.ops[c.op]
	idx := b.newSpan(parent, Span{Name: o.name, Kind: KindClient, StartNs: start, Resource: t.g.resource(svc, pod), Scope: svc})
	net := uint64(100_000 + r.IntN(1_500_000))
	srv := t.server(b, idx, c.svc, c.op, start+net/2, depth+1, user)
	end := b.spans[srv].EndNs + net/2
	failed := b.spans[srv].Status == StatusError
	sp := &b.spans[idx]
	sp.EndNs = end
	host := callee.name
	if callee.proto == "http" {
		var status int64 = 200
		for _, a := range b.spans[srv].Attrs {
			if a.Key == "http.response.status_code" {
				status = a.Int
			}
		}
		url := fmt.Sprintf("http://%s:8080%s", host, o.route)
		for _, a := range b.spans[srv].Attrs {
			if a.Key == "url.path" {
				url = fmt.Sprintf("http://%s:8080%s", host, a.Str)
			}
		}
		sp.Attrs = []Attr{
			S("http.request.method", o.method),
			I("http.response.status_code", status),
			S("server.address", host),
			I("server.port", 8080),
			S("url.full", url),
		}
	} else {
		var code int64
		for _, a := range b.spans[srv].Attrs {
			if a.Key == "rpc.grpc.status_code" {
				code = a.Int
			}
		}
		sp.Attrs = []Attr{
			I("rpc.grpc.status_code", code),
			S("rpc.method", o.route),
			S("rpc.service", o.method),
			S("rpc.system", "grpc"),
			S("server.address", host),
		}
	}
	if failed {
		sp.Status = StatusError
	}
	return end, failed
}

// leaf emits a database, internal or messaging span.
func (t *TraceGen) leaf(b *traceBuild, parent, svc, pod int, c call, start uint64) (uint64, bool) {
	r := b.r
	s := &t.g.svcs[svc]
	dur := uint64(logNormal(r, 0, 0.6) * c.baseMs * 1e6)
	sp := Span{Name: c.name, StartNs: start, EndNs: start + dur, Resource: t.g.resource(svc, pod), Scope: svc}
	failed := false
	switch c.kind {
	case callDB:
		sp.Kind = KindClient
		op, coll := c.name, ""
		for k := 0; k < len(c.name); k++ {
			if c.name[k] == ' ' {
				op, coll = c.name[:k], c.name[k+1:]
				break
			}
		}
		sp.Attrs = []Attr{
			S("db.collection.name", coll),
			S("db.namespace", s.dbName),
			S("db.operation.name", op),
			S("db.system.name", s.db),
			S("server.address", s.name+"-db"),
		}
		if s.db != "redis" && s.db != "mongodb" {
			sp.Attrs = append(sp.Attrs, S("db.query.text", queryText(op, coll)))
		}
		failed = r.Float64() < 0.002
	case callInternal:
		sp.Kind = KindInternal
		sp.Attrs = []Attr{S("code.function.name", c.name)}
	case callPublish:
		sp.Kind = KindProducer
		sp.Attrs = []Attr{
			S("messaging.destination.name", s.topic),
			S("messaging.operation.type", "send"),
			S("messaging.system", "kafka"),
		}
	}
	if failed {
		sp.Status = StatusError
		sp.StatusMessage = pick(r, errMessages)
		sp.Events = []Event{exception(r, start+dur/2)}
	}
	b.newSpan(parent, sp)
	return start + dur, failed
}

func queryText(op, coll string) string {
	switch op {
	case "INSERT":
		return "INSERT INTO " + coll + " (id, payload) VALUES (?, ?)"
	case "UPDATE":
		return "UPDATE " + coll + " SET payload = ? WHERE id = ?"
	}
	return "SELECT * FROM " + coll + " WHERE id = ?"
}

package otel

import (
	"fmt"
	"math/rand/v2"
	"strings"
	"time"
)

// LogsConfig controls log generation.
type LogsConfig struct {
	TimeEnd  time.Time
	TimeSpan time.Duration // default 72h
	// Correlated is the fraction of records emitted inside a generated span
	// (TraceId/SpanId set, service and time taken from the span). Default 0.3.
	Correlated float64
	// TraceSpace is the trace index range correlated records draw from; it should
	// equal the number of traces loaded by the otel-traces workload. Default 100_000.
	TraceSpace uint64
	// NeedleEvery plants a rare token in one record out of NeedleEvery. Default 100_000.
	NeedleEvery uint64
}

// LogGen generates log records as a pure function of (seed, record index).
type LogGen struct {
	cfg LogsConfig
	tg  *TraceGen
}

// NewLogGen builds a generator correlated with tg (same graph and seed).
func NewLogGen(cfg LogsConfig, tg *TraceGen) *LogGen {
	if cfg.TimeSpan <= 0 {
		cfg.TimeSpan = 72 * time.Hour
	}
	if cfg.Correlated == 0 {
		cfg.Correlated = 0.3
	}
	if cfg.TraceSpace == 0 {
		cfg.TraceSpace = 100_000
	}
	if cfg.NeedleEvery == 0 {
		cfg.NeedleEvery = 100_000
	}
	return &LogGen{cfg: cfg, tg: tg}
}

// Window is the span of uncorrelated record times (correlated records follow
// their trace, which starts in the trace generator's window).
func (l *LogGen) Window() (time.Time, time.Time) {
	return l.cfg.TimeEnd.Add(-l.cfg.TimeSpan), l.cfg.TimeEnd
}

type severity struct {
	text string
	num  int32
	w    float64
}

var severities = []severity{
	{"TRACE", 1, 0.01}, {"DEBUG", 5, 0.14}, {"INFO", 9, 0.65}, {"WARN", 13, 0.13}, {"ERROR", 17, 0.065}, {"FATAL", 21, 0.005},
}

// Needle returns the planted token of record j, or "" when j carries none.
func (l *LogGen) Needle(j uint64) string {
	seed := l.tg.g.cfg.Seed
	if hash(seed, streamLog, 0x6e65, j)%l.cfg.NeedleEvery != 0 {
		return ""
	}
	return fmt.Sprintf("ndl%08x", hash(seed, streamLog, 0x6e66, j)&0xffffffff)
}

// Record returns log record j.
func (l *LogGen) Record(j uint64) LogRecord {
	g := l.tg.g
	r := newRNG(g.cfg.Seed, streamLog, j)
	var rec LogRecord
	var svc, pod int
	sev := pickSeverity(r)
	if r.Float64() < l.cfg.Correlated {
		ti := r.Uint64N(l.cfg.TraceSpace)
		spans := l.tg.Trace(ti)
		sp := spans[r.IntN(len(spans))]
		rec.TraceID, rec.SpanID, rec.Flags = sp.TraceID, sp.SpanID, 1
		rec.TimeNs = sp.StartNs + uint64(r.Int64N(int64(sp.EndNs-sp.StartNs)+1))
		svc = sp.Scope
		pod = sp.Resource - g.resource(svc, 0)
		if sp.Status == StatusError {
			sev = severities[4]
		}
	} else {
		svc = r.IntN(len(g.svcs))
		pod = r.IntN(len(g.svcs[svc].pods))
		from, _ := l.Window()
		rec.TimeNs = uint64(from.UnixNano()) + uint64(r.Int64N(int64(l.cfg.TimeSpan)))
	}
	rec.ObservedTimeNs = rec.TimeNs + uint64(r.IntN(500_000_000))
	rec.Resource, rec.Scope = g.resource(svc, pod), svc
	rec.SeverityText, rec.SeverityNumber = sev.text, sev.num
	rec.Body, rec.Attrs = l.body(r, svc, sev.text)
	if n := l.Needle(j); n != "" {
		rec.Body += " token=" + n
	}
	return rec
}

func pickSeverity(r *rand.Rand) severity {
	x := r.Float64()
	for _, s := range severities {
		if x < s.w {
			return s
		}
		x -= s.w
	}
	return severities[2]
}

// body renders a per-severity message template with variable parts and returns
// the record attributes the template sets (keys sorted, no duplicates).
func (l *LogGen) body(r *rand.Rand, svc int, sev string) (string, []Attr) {
	g := l.tg.g
	s := &g.svcs[svc]
	op := &s.ops[r.IntN(len(s.ops))]
	other := g.svcs[r.IntN(len(g.svcs))].name
	noun := pick(r, nouns)
	logger := S("log.logger", strings.ReplaceAll(s.name, "-", ".")+".app")
	user := fmt.Sprintf("user-%d", r.IntN(g.cfg.Users))
	ms := int(logNormal(r, 3, 1))
	switch sev {
	case "TRACE":
		fn := strings.ReplaceAll(s.name, "-", "_") + ".handle" + camel(noun)
		return fmt.Sprintf("entering %s args=%d", fn, r.IntN(6)), []Attr{S("code.function.name", fn), logger}
	case "DEBUG":
		switch r.IntN(3) {
		case 0:
			return fmt.Sprintf("cache miss key=%s:%d", noun, r.IntN(1_000_000)), []Attr{logger}
		case 1:
			return fmt.Sprintf("loaded config revision %d", r.IntN(500)), []Attr{logger}
		}
		return fmt.Sprintf("sql query took %dms rows=%d", ms, r.IntN(1000)), []Attr{S("db.system.name", firstNonEmpty(s.db, "postgresql")), logger}
	case "WARN":
		switch r.IntN(3) {
		case 0:
			return fmt.Sprintf("slow request route=%s duration_ms=%d", op.route, 1000+ms*10), []Attr{S("http.route", op.route), logger}
		case 1:
			return fmt.Sprintf("retrying call to %s attempt=%d", other, 1+r.IntN(3)), []Attr{logger, S("peer.service", other)}
		}
		return fmt.Sprintf("connection pool near capacity used=%d/%d", 45+r.IntN(5), 50), []Attr{logger}
	case "ERROR":
		switch r.IntN(3) {
		case 0:
			return fmt.Sprintf("failed to process %s %d: %s", noun, r.IntN(1_000_000), pick(r, errMessages)), []Attr{S("exception.type", "ProcessingError"), logger}
		case 1:
			return fmt.Sprintf("upstream %s returned status %d", other, pick(r, []int{500, 502, 503})), []Attr{logger, S("peer.service", other)}
		}
		return fmt.Sprintf("timeout after %dms calling %s", 1000+ms*10, other), []Attr{logger, S("peer.service", other)}
	case "FATAL":
		return fmt.Sprintf("unrecoverable error in %s worker: %s", noun, pick(r, errMessages)), []Attr{logger}
	}
	switch r.IntN(4) {
	case 0:
		code := pick(r, []int64{200, 200, 200, 201, 204, 404})
		return fmt.Sprintf("request completed method=%s route=%s status=%d duration_ms=%d", op.method, op.route, code, ms),
			[]Attr{I("http.response.status_code", code), S("http.route", op.route), logger}
	case 1:
		return fmt.Sprintf("user %s signed in", user), []Attr{logger, S("user.id", user)}
	case 2:
		return fmt.Sprintf("processed %d %s in %dms", r.IntN(500), noun, ms), []Attr{logger}
	}
	return fmt.Sprintf("order %d accepted for user %s", r.IntN(10_000_000), user), []Attr{logger, S("user.id", user)}
}

func firstNonEmpty(a, b string) string {
	if a != "" {
		return a
	}
	return b
}

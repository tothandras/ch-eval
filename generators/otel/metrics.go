package otel

import (
	"fmt"
	"math"
	"time"
)

// MetricType is the OTLP data type, which selects the exporter table.
type MetricType int

const (
	Gauge MetricType = iota
	Sum
	Histogram
	ExpHistogram
	Summary
)

// MetricTypes lists every type in table order.
var MetricTypes = []MetricType{Gauge, Sum, Histogram, ExpHistogram, Summary}

// Table is the exporter's default table name for the type.
func (t MetricType) Table() string {
	return [...]string{"otel_metrics_gauge", "otel_metrics_sum", "otel_metrics_histogram", "otel_metrics_exponential_histogram", "otel_metrics_summary"}[t]
}

// MetricsConfig controls metric generation.
type MetricsConfig struct {
	TimeEnd  time.Time
	TimeSpan time.Duration // default 72h
	Interval time.Duration // collection interval, default 60s
	// MaxPoints caps the points generated; the window then covers only the last
	// MaxPoints/series intervals before TimeEnd. 0 = no cap.
	MaxPoints uint64
	// HighCardUsers adds app.user.requests, a monotonic sum with one series per
	// user.id (the high-cardinality profile); 0 disables it.
	HighCardUsers int
}

type metricDef struct {
	name, desc, unit string
	typ              MetricType
	monotonic        bool
	isInt            bool
}

// MetricPoint is one data point. Fields not used by the point's type are zero.
type MetricPoint struct {
	Type          MetricType
	Name          string
	Description   string
	Unit          string
	Monotonic     bool
	IsInt         bool // gauge/sum value is an OTLP asInt
	Temporality   int32
	Attrs         []Attr
	Resource      int
	Scope         int
	StartNs       uint64 // 0 for gauges
	TimeNs        uint64
	Value         float64 // gauge/sum
	Count         uint64
	SumValue      float64
	Bounds        []float64 // histogram
	BucketCounts  []uint64  // histogram, or exp histogram positive buckets
	Scale         int32
	ZeroCount     uint64
	PosOffset     int32
	Quantiles     []float64 // summary
	QuantileVals  []float64
}

type series struct {
	def      int
	resource int
	scope    int
	attrs    []Attr
	rate     float64 // events per second (cumulative types) or level (gauges)
	mu, sig  float64 // value distribution (histograms, summaries)
	period   uint64  // reset period in intervals; 0 = never resets
	phase    uint64
	startOff uint64 // process start before the window, in intervals
}

// MetricGen generates points as a pure function of (seed, point index). Points are
// ordered interval-major: point i is series i%S at interval i/S.
type MetricGen struct {
	cfg    MetricsConfig
	g      *Graph
	defs   []metricDef
	series []series
	slots  uint64
	start  time.Time
}

var httpBounds = []float64{0.005, 0.01, 0.025, 0.05, 0.075, 0.1, 0.25, 0.5, 0.75, 1, 2.5, 5, 7.5, 10}

var summaryQuantiles = []float64{0, 0.25, 0.5, 0.75, 1}

// NewMetricGen enumerates the series of the graph.
func NewMetricGen(cfg MetricsConfig, g *Graph) *MetricGen {
	if cfg.TimeSpan <= 0 {
		cfg.TimeSpan = 72 * time.Hour
	}
	if cfg.Interval <= 0 {
		cfg.Interval = time.Minute
	}
	m := &MetricGen{cfg: cfg, g: g}
	m.defs = []metricDef{
		{"system.cpu.utilization", "Difference in system.cpu.time since the last measurement, divided by the elapsed time and number of logical CPUs.", "1", Gauge, false, false},
		{"process.memory.usage", "The amount of physical memory in use.", "By", Gauge, false, true},
		{"http.server.request.count", "Number of HTTP server requests.", "{request}", Sum, true, true},
		{"db.client.connection.count", "The number of connections that are currently in state described by the state attribute.", "{connection}", Sum, false, true},
		{"http.server.request.duration", "Duration of HTTP server requests.", "s", Histogram, false, false},
		{"rpc.server.duration", "Measures the duration of inbound RPC.", "ms", ExpHistogram, false, false},
		{"go.gc.duration", "A summary of the wall-time pause (stop-the-world) duration in garbage collection cycles.", "s", Summary, false, false},
		{"app.user.requests", "Requests per end user.", "{request}", Sum, true, true},
	}
	seed := g.cfg.Seed
	add := func(def, res, scope int, attrs []Attr) {
		k := uint64(len(m.series))
		h := hash(seed, streamMet, 0x5e, k)
		s := series{def: def, resource: res, scope: scope, attrs: attrs,
			rate: 0.05 + 20*unit(hash(h, 1))*unit(hash(h, 2)),
			mu:   math.Log(0.002 + 0.2*unit(hash(h, 3))), sig: 0.4 + 0.8*unit(hash(h, 4)),
			phase: hash(h, 5) % 10_000, startOff: 1 + hash(h, 6)%1440}
		if unit(hash(h, 7)) < 0.3 {
			s.period = 500 + hash(h, 8)%2500
		}
		m.series = append(m.series, s)
	}
	for si, s := range g.svcs {
		for p := range s.pods {
			res := g.resource(si, p)
			for _, mode := range []string{"idle", "iowait", "system", "user"} {
				add(0, res, si, []Attr{S("cpu.mode", mode)})
			}
			add(1, res, si, nil)
			if s.db != "" {
				for _, st := range []string{"idle", "used"} {
					add(3, res, si, []Attr{S("db.client.connection.pool.name", s.dbName+"-pool"), S("db.client.connection.state", st)})
				}
			}
			for _, op := range s.ops {
				if s.proto == "http" {
					for _, code := range []int64{200, 500} {
						add(2, res, si, []Attr{S("http.request.method", op.method), I("http.response.status_code", code), S("http.route", op.route)})
					}
					add(4, res, si, []Attr{S("http.request.method", op.method), S("http.route", op.route)})
				} else {
					add(5, res, si, []Attr{S("rpc.method", op.route), S("rpc.service", op.method), S("rpc.system", "grpc")})
				}
			}
			if s.lang == "go" {
				add(6, res, si, nil)
			}
		}
	}
	for u := 0; u < cfg.HighCardUsers; u++ {
		p := u % g.cfg.PodsPerService
		add(7, g.resource(0, p), 0, []Attr{S("user.id", fmt.Sprintf("user-%d", u))})
	}
	m.slots = uint64(cfg.TimeSpan / cfg.Interval)
	if cfg.MaxPoints > 0 {
		m.slots = min(m.slots, max(1, cfg.MaxPoints/uint64(len(m.series))))
	}
	m.start = cfg.TimeEnd.Truncate(cfg.Interval).Add(-time.Duration(m.slots) * cfg.Interval)
	return m
}

// Series is the number of distinct series.
func (m *MetricGen) Series() int { return len(m.series) }

// Points is the number of points the generator emits.
func (m *MetricGen) Points() uint64 { return m.slots * uint64(len(m.series)) }

// Window is [first point time, TimeEnd].
func (m *MetricGen) Window() (time.Time, time.Time) {
	return m.start.Add(m.cfg.Interval), m.start.Add(time.Duration(m.slots) * m.cfg.Interval)
}

// Point returns point i.
func (m *MetricGen) Point(i uint64) MetricPoint {
	n := uint64(len(m.series))
	slot, si := i/n+1, i%n
	s := &m.series[si]
	d := &m.defs[s.def]
	iv := m.cfg.Interval
	t := m.start.Add(time.Duration(slot) * iv)
	seed := m.g.cfg.Seed
	h := hash(seed, streamMet, si, slot)
	p := MetricPoint{Type: d.typ, Name: d.name, Description: d.desc, Unit: d.unit, Monotonic: d.monotonic, IsInt: d.isInt,
		Attrs: s.attrs, Resource: s.resource, Scope: s.scope, TimeNs: uint64(t.UnixNano())}
	// Cumulative types: slots since the last reset (k) and the reset's start time.
	abs := slot + s.startOff
	k := abs
	if s.period > 0 {
		k = (abs+s.phase)%s.period + 1
		if k > abs {
			k = abs
		}
	}
	p.StartNs = uint64(t.Add(-time.Duration(k) * iv).UnixNano())
	sec := iv.Seconds()
	// cum is non-decreasing within a reset epoch: rate*sec*(k + u_k), u in [0,1).
	cum := func(rate float64) float64 { return rate * sec * (float64(k) + unit(hash(h, 11))) }
	switch d.typ {
	case Gauge:
		p.StartNs = 0
		day := 2 * math.Pi * float64(t.Unix()%86400) / 86400
		if d.isInt {
			p.Value = math.Floor(5e7 + 4e8*unit(hash(seed, streamMet, 0x6d, si)) + 5e7*math.Sin(day+float64(s.phase)) + 1e6*unit(h))
		} else {
			base := 0.05 + 0.4*unit(hash(seed, streamMet, 0x63, si))
			p.Value = math.Round(clamp(base+0.2*math.Sin(day+float64(s.phase))+0.05*unit(h), 0, 1)*1e4) / 1e4
		}
	case Sum:
		p.Temporality = 2
		if d.monotonic {
			p.Value = math.Floor(cum(s.rate))
		} else {
			p.StartNs = uint64(m.start.Add(-time.Duration(s.startOff) * iv).UnixNano())
			p.Value = float64(hash(h, 12) % 50)
		}
	case Histogram:
		p.Temporality = 2
		p.Count = uint64(cum(s.rate))
		p.SumValue = math.Round(float64(p.Count)*math.Exp(s.mu+s.sig*s.sig/2)*1e6) / 1e6
		p.Bounds = httpBounds
		p.BucketCounts = bucketize(p.Count, s.mu, s.sig, httpBounds)
	case ExpHistogram:
		p.Temporality = 2
		p.Count = uint64(cum(s.rate))
		mu := s.mu + math.Log(1000) // milliseconds
		p.SumValue = math.Round(float64(p.Count)*math.Exp(mu+s.sig*s.sig/2)*1e3) / 1e3
		p.Scale = 2
		base := math.Pow(2, math.Pow(2, -float64(p.Scale)))
		lo := int32(math.Floor((mu - 3*s.sig) / math.Log(base)))
		var bounds []float64
		for b := lo; b < lo+24; b++ {
			bounds = append(bounds, math.Pow(base, float64(b+1)))
		}
		p.PosOffset = lo
		p.BucketCounts = bucketize(p.Count, mu, s.sig, bounds[:len(bounds)-1])
	case Summary:
		p.Count = uint64(cum(s.rate))
		p.SumValue = math.Round(float64(p.Count)*math.Exp(s.mu)*1e6) / 1e6
		p.Quantiles = summaryQuantiles
		z := []float64{-2.5, -0.674, 0, 0.674, 2.5}
		for qi := range summaryQuantiles {
			v := math.Exp(s.mu+s.sig*z[qi]) * (1 + 0.1*unit(hash(h, uint64(20+qi))))
			p.QuantileVals = append(p.QuantileVals, math.Round(v*1e6)/1e6)
		}
	}
	return p
}

// bucketize splits count over buckets bounded by bounds (len(bounds)+1 buckets)
// with a log-normal CDF. Cumulative counts are floor(count*CDF), so each bucket is
// non-decreasing as count grows and the buckets sum to count.
func bucketize(count uint64, mu, sigma float64, bounds []float64) []uint64 {
	out := make([]uint64, len(bounds)+1)
	var prev uint64
	for b, ub := range bounds {
		cdf := 0.5 * math.Erfc(-(math.Log(ub)-mu)/(sigma*math.Sqrt2))
		c := uint64(math.Floor(float64(count) * cdf))
		if c < prev {
			c = prev
		}
		out[b] = c - prev
		prev = c
	}
	out[len(bounds)] = count - prev
	return out
}

func clamp(x, lo, hi float64) float64 { return math.Max(lo, math.Min(hi, x)) }

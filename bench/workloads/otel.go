package workloads

import (
	"context"
	"fmt"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"

	"github.com/tothandras/ch-eval/bench/runner"
	"github.com/tothandras/ch-eval/generators/otel"
)

// OTel generators (design D6). --rows means spans (otel-traces), log records
// (otel-logs) or metric points across the five tables (otel-metrics). The shared
// graph uses the default shape; the window is 72h ending at --time-end.

const otelSpan = 72 * time.Hour

// LogTraceSpace is the trace index range otel-logs correlates with. It equals the
// number of traces otel-traces writes at its default 1M spans (about 50k traces),
// so correlated records point at traces that exist when both workloads are loaded
// with defaults.
const LogTraceSpace = 50_000

// LogNeedleEvery plants one rare token per this many log records.
const LogNeedleEvery = 10_000

func graph(o Opts) *otel.Graph { return otel.NewGraph(otel.DefaultGraph(o.Seed)) }

func traceGen(o Opts) *otel.TraceGen {
	g := graph(o)
	return otel.NewTraceGen(otel.TracesConfig{Graph: otel.DefaultGraph(o.Seed), TimeEnd: o.TimeEnd, TimeSpan: otelSpan}, g)
}

func logGen(o Opts) *otel.LogGen {
	return otel.NewLogGen(otel.LogsConfig{TimeEnd: o.TimeEnd, TimeSpan: otelSpan, TraceSpace: LogTraceSpace, NeedleEvery: LogNeedleEvery}, traceGen(o))
}

func metricGen(o Opts) *otel.MetricGen {
	cfg := otel.MetricsConfig{TimeEnd: o.TimeEnd, TimeSpan: otelSpan, MaxPoints: uint64(o.Rows)}
	if o.Profile == "high-card" {
		cfg.HighCardUsers = 100_000
	}
	return otel.NewMetricGen(cfg, graph(o))
}

func batchSize(o Opts) int {
	if o.BatchSize > 0 {
		return o.BatchSize
	}
	return 10_000
}

// sink writes one batch through the selected ingest path.
type sink struct {
	direct *otel.Direct
	sender *otel.Sender
	g      *otel.Graph
}

func newSink(conn driver.Conn, db string, g *otel.Graph, o Opts) (*sink, error) {
	switch o.IngestPath {
	case "", "direct":
		return &sink{direct: &otel.Direct{Conn: conn, Database: db, Graph: g, Settings: o.Settings}, g: g}, nil
	case "collector":
		if o.CollectorURL == "" {
			return nil, fmt.Errorf("collector ingest path needs a collector URL")
		}
		return &sink{sender: &otel.Sender{Endpoint: o.CollectorURL}, g: g}, nil
	}
	return nil, fmt.Errorf("unknown ingest path %q", o.IngestPath)
}

func ingestResult(path string, rows uint64, batch int, start time.Time) *runner.IngestResult {
	d := time.Since(start)
	src := "otel-direct"
	if path == "collector" {
		src = "otel-collector"
	}
	return &runner.IngestResult{Source: src, Rows: int(rows), DurationSeconds: d.Seconds(), EventsPerSecond: float64(rows) / d.Seconds(), BatchSize: batch}
}

// ---- traces

type otelTraces struct{}

func init() { register("otel-traces", otelTraces{}) }

func (otelTraces) Seed(ctx context.Context, conn driver.Conn, db string, tables []string, o Opts) (Result, error) {
	tg := traceGen(o)
	s, err := newSink(conn, db, tg.Graph(), o)
	if err != nil {
		return Result{}, err
	}
	bs := batchSize(o)
	start := time.Now()
	var buf []otel.Span
	var spans, traces uint64
	flush := func() error {
		if len(buf) == 0 {
			return nil
		}
		var err error
		if s.direct != nil {
			err = s.direct.Spans(ctx, tables[0], buf)
		} else {
			var body []byte
			if body, err = otel.EncodeTraces(s.g, buf); err == nil {
				err = s.sender.Post(ctx, "traces", body)
			}
		}
		buf = buf[:0]
		return err
	}
	// Whole traces per batch: the exporter-default trace_id_ts view aggregates per
	// insert block, so a trace never straddles two direct batches.
	for spans < uint64(o.Rows) {
		tr := tg.Trace(traces)
		traces++
		spans += uint64(len(tr))
		buf = append(buf, tr...)
		if len(buf) >= bs {
			if err := flush(); err != nil {
				return Result{}, err
			}
		}
	}
	if err := flush(); err != nil {
		return Result{}, err
	}
	return Result{Ingest: ingestResult(o.IngestPath, spans, bs, start), Rows: map[string]uint64{tables[0]: spans},
		Notes: []string{fmt.Sprintf("%d traces", traces)}}, nil
}

func (otelTraces) Window(o Opts) (time.Time, time.Time) { return o.TimeEnd.Add(-otelSpan), o.TimeEnd }

// Params derive from trace 7 (present whenever at least ~200 spans are loaded).
func (w otelTraces) Params(o Opts) (map[string]string, error) {
	tg := traceGen(o)
	tr := tg.Trace(7)
	root := tr[0]
	user := ""
	for _, a := range root.Attrs {
		if a.Key == "user.id" {
			user = a.String()
		}
	}
	if user == "" {
		user = "user-1"
	}
	g := tg.Graph()
	res := g.Resources()[root.Resource]
	from, to := w.Window(o)
	route := root.Name
	for _, a := range root.Attrs {
		if a.Key == "http.route" {
			route = a.String()
		}
	}
	return map[string]string{
		"from":      unix(from),
		"to":        unix(to),
		"from_1h":   unix(to.Add(-time.Hour)),
		"from_24h":  unix(to.Add(-24 * time.Hour)),
		"trace_id":  sqlString(otel.TraceIDHex(o.Seed, 7)),
		"service":   sqlString(res.ServiceName()),
		"span_name": sqlString(root.Name),
		"route":     sqlString(route),
		"user_id":   sqlString(user),
	}, nil
}

// ---- logs

type otelLogs struct{}

func init() { register("otel-logs", otelLogs{}) }

func (otelLogs) Seed(ctx context.Context, conn driver.Conn, db string, tables []string, o Opts) (Result, error) {
	lg := logGen(o)
	g := traceGen(o).Graph()
	s, err := newSink(conn, db, g, o)
	if err != nil {
		return Result{}, err
	}
	bs := batchSize(o)
	start := time.Now()
	buf := make([]otel.LogRecord, 0, bs)
	flush := func() error {
		if len(buf) == 0 {
			return nil
		}
		var err error
		if s.direct != nil {
			err = s.direct.Logs(ctx, tables[0], buf)
		} else {
			var body []byte
			if body, err = otel.EncodeLogs(g, buf); err == nil {
				err = s.sender.Post(ctx, "logs", body)
			}
		}
		buf = buf[:0]
		return err
	}
	for j := uint64(0); j < uint64(o.Rows); j++ {
		buf = append(buf, lg.Record(j))
		if len(buf) >= bs {
			if err := flush(); err != nil {
				return Result{}, err
			}
		}
	}
	if err := flush(); err != nil {
		return Result{}, err
	}
	return Result{Ingest: ingestResult(o.IngestPath, uint64(o.Rows), bs, start), Rows: map[string]uint64{tables[0]: uint64(o.Rows)}}, nil
}

func (otelLogs) Window(o Opts) (time.Time, time.Time) { return o.TimeEnd.Add(-otelSpan), o.TimeEnd }

// Params: the first planted needle, the first correlated trace, and a user and
// service that occur early in the record stream.
func (w otelLogs) Params(o Opts) (map[string]string, error) {
	lg := logGen(o)
	g := traceGen(o).Graph()
	needle, traceID, user, service := "", "", "", ""
	limit := uint64(max(o.Rows, 1))
	for j := uint64(0); j < limit && (needle == "" || traceID == "" || user == ""); j++ {
		if needle == "" {
			needle = lg.Needle(j)
		}
		if traceID == "" || user == "" {
			r := lg.Record(j)
			if service == "" {
				service = g.Resources()[r.Resource].ServiceName()
			}
			if traceID == "" && r.TraceID != ([16]byte{}) {
				traceID = fmt.Sprintf("%x", r.TraceID)
			}
			for _, a := range r.Attrs {
				if user == "" && a.Key == "user.id" {
					user = a.String()
				}
			}
		}
	}
	if needle == "" {
		needle = "ndl00000000" // fewer rows than LogNeedleEvery: the search finds nothing, identically everywhere
	}
	from, to := w.Window(o)
	return map[string]string{
		"from":     unix(from),
		"to":       unix(to),
		"from_1h":  unix(to.Add(-time.Hour)),
		"from_24h": unix(to.Add(-24 * time.Hour)),
		"needle":   sqlString(needle),
		"trace_id": sqlString(traceID),
		"user_id":  sqlString(user),
		"service":  sqlString(service),
	}, nil
}

// ---- metrics

type otelMetrics struct{}

func init() { register("otel-metrics", otelMetrics{}) }

func (otelMetrics) Seed(ctx context.Context, conn driver.Conn, db string, tables []string, o Opts) (Result, error) {
	mg := metricGen(o)
	g := graph(o)
	s, err := newSink(conn, db, g, o)
	if err != nil {
		return Result{}, err
	}
	tableOf := map[otel.MetricType]string{}
	for i, typ := range otel.MetricTypes {
		if i < len(tables) {
			tableOf[typ] = tables[i]
		}
	}
	bs := batchSize(o)
	start := time.Now()
	buf := make([]otel.MetricPoint, 0, bs)
	rows := map[string]uint64{}
	flush := func() error {
		if len(buf) == 0 {
			return nil
		}
		var err error
		if s.direct != nil {
			err = s.direct.Metrics(ctx, tableOf, buf)
		} else {
			var body []byte
			if body, err = otel.EncodeMetrics(g, buf); err == nil {
				err = s.sender.Post(ctx, "metrics", body)
			}
		}
		buf = buf[:0]
		return err
	}
	n := mg.Points()
	for i := uint64(0); i < n; i++ {
		p := mg.Point(i)
		rows[tableOf[p.Type]]++
		buf = append(buf, p)
		if len(buf) >= bs {
			if err := flush(); err != nil {
				return Result{}, err
			}
		}
	}
	if err := flush(); err != nil {
		return Result{}, err
	}
	return Result{Ingest: ingestResult(o.IngestPath, n, bs, start), Rows: rows,
		Notes: []string{fmt.Sprintf("%d series", mg.Series())}}, nil
}

func (otelMetrics) Window(o Opts) (time.Time, time.Time) {
	from, to := metricGen(o).Window()
	return from, to.Add(time.Second)
}

func (w otelMetrics) Params(o Opts) (map[string]string, error) {
	g := graph(o)
	from, to := w.Window(o)
	svc := g.ServiceNames()[0]
	return map[string]string{
		"from":     unix(from),
		"to":       unix(to),
		"from_1h":  unix(to.Add(-time.Hour)),
		"service":  sqlString(svc),
		"user_id":  sqlString("user-42"),
		"counter":  sqlString("http.server.request.count"),
		"gauge":    sqlString("system.cpu.utilization"),
		"histo":    sqlString("http.server.request.duration"),
		"exphisto": sqlString("rpc.server.duration"),
		"summary":  sqlString("go.gc.duration"),
	}, nil
}

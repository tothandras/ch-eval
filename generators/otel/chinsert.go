package otel

import (
	"context"
	"encoding/hex"
	"fmt"
	"sort"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/column"
	"github.com/ClickHouse/clickhouse-go/v2/lib/column/orderedmap"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
)

// The INSERT column lists are the exporter's (sqltemplates/*_insert.sql at
// v0.162.0; docs/research/otel-ddl/exporter_inserts.sql). Both schema variants
// accept them: ClickStack's extra columns are MATERIALIZED or ALIAS, and both
// logs tables carry EventName, which the exporter then appends.
const (
	tracesColumns = `Timestamp, TraceId, SpanId, ParentSpanId, TraceState, SpanName, SpanKind, ServiceName,
    ResourceAttributes, ScopeName, ScopeVersion, SpanAttributes, Duration, StatusCode, StatusMessage,
    Events.Timestamp, Events.Name, Events.Attributes, Links.TraceId, Links.SpanId, Links.TraceState, Links.Attributes`
	logsColumns = `Timestamp, TraceId, SpanId, TraceFlags, SeverityText, SeverityNumber, ServiceName, Body,
    ResourceSchemaUrl, ResourceAttributes, ScopeSchemaUrl, ScopeName, ScopeVersion, ScopeAttributes, LogAttributes, EventName`
	metricsCommon = `ResourceAttributes, ResourceSchemaUrl, ScopeName, ScopeVersion, ScopeAttributes, ScopeDroppedAttrCount,
    ScopeSchemaUrl, ServiceName, MetricName, MetricDescription, MetricUnit, Attributes, StartTimeUnix, TimeUnix`
	exemplarCols = `Exemplars.FilteredAttributes, Exemplars.TimeUnix, Exemplars.Value, Exemplars.SpanId, Exemplars.TraceId`
)

var metricColumns = map[MetricType]string{
	Gauge:        metricsCommon + `, Value, Flags, ` + exemplarCols,
	Sum:          metricsCommon + `, Value, Flags, ` + exemplarCols + `, AggregationTemporality, IsMonotonic`,
	Histogram:    metricsCommon + `, Count, Sum, BucketCounts, ExplicitBounds, ` + exemplarCols + `, Flags, Min, Max, AggregationTemporality`,
	ExpHistogram: metricsCommon + `, Count, Sum, Scale, ZeroCount, PositiveOffset, PositiveBucketCounts, NegativeOffset, NegativeBucketCounts, ` + exemplarCols + `, Flags, Min, Max, AggregationTemporality`,
	Summary:      metricsCommon + `, Count, Sum, ValueAtQuantiles.Quantile, ValueAtQuantiles.Value, Flags`,
}

// attrMap mirrors the exporter's AttributesToMap: values rendered with AsString,
// entries sorted by key (orderedmap.CollectN sorts).
func attrMap(attrs []Attr) column.IterableOrderedMap {
	return orderedmap.CollectN(func(yield func(string, string) bool) {
		for _, a := range attrs {
			if !yield(a.Key, a.String()) {
				return
			}
		}
	}, len(attrs))
}

func nsTime(ns uint64) time.Time { return time.Unix(0, int64(ns)).UTC() }

// hexOrEmpty mirrors traceutil.TraceIDToHexOrEmptyString / SpanIDToHexOrEmptyString.
func hexOrEmpty(b []byte) string {
	for _, x := range b {
		if x != 0 {
			return hex.EncodeToString(b)
		}
	}
	return ""
}

// Direct writes generated records with the exporter's column mapping.
type Direct struct {
	Conn     driver.Conn
	Database string
	Graph    *Graph
	// Settings are applied to every INSERT (for example async_insert=0).
	Settings map[string]any
}

func (d *Direct) batch(ctx context.Context, table, cols string) (driver.Batch, error) {
	if len(d.Settings) > 0 {
		ctx = clickhouse.Context(ctx, clickhouse.WithSettings(clickhouse.Settings(d.Settings)))
	}
	return d.Conn.PrepareBatch(ctx, fmt.Sprintf("INSERT INTO %s.%s (%s)", d.Database, table, cols))
}

// Spans inserts spans into table (otel_traces).
func (d *Direct) Spans(ctx context.Context, table string, spans []Span) error {
	b, err := d.batch(ctx, table, tracesColumns)
	if err != nil {
		return err
	}
	defer b.Close()
	res := d.Graph.Resources()
	scopes := d.Graph.Scopes()
	resMaps := make(map[int]column.IterableOrderedMap)
	for i := range spans {
		s := &spans[i]
		rm, ok := resMaps[s.Resource]
		if !ok {
			rm = attrMap(res[s.Resource].Attrs)
			resMaps[s.Resource] = rm
		}
		var evT []time.Time
		var evN []string
		var evA []column.IterableOrderedMap
		for _, e := range s.Events {
			evT, evN, evA = append(evT, nsTime(e.TimeNs)), append(evN, e.Name), append(evA, attrMap(e.Attrs))
		}
		var lT, lS, lState []string
		var lA []column.IterableOrderedMap
		for _, l := range s.Links {
			lT, lS, lState, lA = append(lT, hexOrEmpty(l.TraceID[:])), append(lS, hexOrEmpty(l.SpanID[:])), append(lState, l.TraceState), append(lA, attrMap(l.Attrs))
		}
		sc := scopes[s.Scope]
		if err := b.Append(nsTime(s.StartNs), hexOrEmpty(s.TraceID[:]), hexOrEmpty(s.SpanID[:]), hexOrEmpty(s.ParentSpanID[:]),
			s.TraceState, s.Name, s.Kind.String(), res[s.Resource].ServiceName(), rm, sc.Name, sc.Version, attrMap(s.Attrs),
			s.EndNs-s.StartNs, s.Status.String(), s.StatusMessage, evT, evN, evA, lT, lS, lState, lA); err != nil {
			return fmt.Errorf("append span: %w", err)
		}
	}
	return b.Send()
}

// Logs inserts log records into table (otel_logs).
func (d *Direct) Logs(ctx context.Context, table string, recs []LogRecord) error {
	b, err := d.batch(ctx, table, logsColumns)
	if err != nil {
		return err
	}
	defer b.Close()
	res := d.Graph.Resources()
	scopes := d.Graph.LogScopes()
	for i := range recs {
		r := &recs[i]
		ts := r.TimeNs
		if ts == 0 {
			ts = r.ObservedTimeNs
		}
		rs, sc := res[r.Resource], scopes[r.Scope]
		if err := b.Append(nsTime(ts), hexOrEmpty(r.TraceID[:]), hexOrEmpty(r.SpanID[:]), uint8(r.Flags), r.SeverityText, uint8(r.SeverityNumber),
			rs.ServiceName(), r.Body, rs.SchemaURL, attrMap(rs.Attrs), sc.SchemaURL, sc.Name, sc.Version, attrMap(sc.Attrs), attrMap(r.Attrs), r.EventName); err != nil {
			return fmt.Errorf("append log: %w", err)
		}
	}
	return b.Send()
}

func arraySet[T any](xs []T) clickhouse.ArraySet {
	out := make(clickhouse.ArraySet, 0, len(xs))
	for _, x := range xs {
		out = append(out, x)
	}
	return out
}

// Metrics inserts points into the per-type tables (tables[type], default
// MetricType.Table()). Types are written one batch each, in MetricTypes order.
func (d *Direct) Metrics(ctx context.Context, tables map[MetricType]string, pts []MetricPoint) error {
	byType := map[MetricType][]*MetricPoint{}
	for i := range pts {
		byType[pts[i].Type] = append(byType[pts[i].Type], &pts[i])
	}
	res := d.Graph.Resources()
	scopes := d.Graph.Scopes()
	for _, typ := range MetricTypes {
		ps := byType[typ]
		if len(ps) == 0 {
			continue
		}
		table := tables[typ]
		if table == "" {
			table = typ.Table()
		}
		b, err := d.batch(ctx, table, metricColumns[typ])
		if err != nil {
			return err
		}
		for _, p := range ps {
			rs, sc := res[p.Resource], scopes[p.Scope]
			common := []any{attrMap(rs.Attrs), rs.SchemaURL, sc.Name, sc.Version, attrMap(sc.Attrs), uint32(0), sc.SchemaURL,
				rs.ServiceName(), p.Name, p.Description, p.Unit, attrMap(p.Attrs), nsTime(p.StartNs), nsTime(p.TimeNs)}
			var noEx clickhouse.ArraySet
			var row []any
			switch typ {
			case Gauge:
				row = append(common, p.Value, uint32(0), noEx, noEx, noEx, noEx, noEx)
			case Sum:
				row = append(common, p.Value, uint32(0), noEx, noEx, noEx, noEx, noEx, p.Temporality, p.Monotonic)
			case Histogram:
				row = append(common, p.Count, p.SumValue, arraySet(p.BucketCounts), arraySet(p.Bounds), noEx, noEx, noEx, noEx, noEx, uint32(0), 0.0, 0.0, p.Temporality)
			case ExpHistogram:
				row = append(common, p.Count, p.SumValue, p.Scale, p.ZeroCount, p.PosOffset, arraySet(p.BucketCounts), int32(0), arraySet([]uint64{}),
					noEx, noEx, noEx, noEx, noEx, uint32(0), 0.0, 0.0, p.Temporality)
			case Summary:
				row = append(common, p.Count, p.SumValue, arraySet(p.Quantiles), arraySet(p.QuantileVals), uint32(0))
			}
			if err := b.Append(row...); err != nil {
				b.Close()
				return fmt.Errorf("append %s point: %w", typ.Table(), err)
			}
		}
		if err := b.Send(); err != nil {
			return err
		}
	}
	return nil
}

// sortedAttrs returns a copy sorted by key (used by the OTLP encoder so that both
// paths present attributes in the same order; the exporter sorts regardless).
func sortedAttrs(a []Attr) []Attr {
	out := append([]Attr(nil), a...)
	sort.SliceStable(out, func(i, j int) bool { return out[i].Key < out[j].Key })
	return out
}

package otel

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
)

// OTLP/HTTP with the JSON encoding (opentelemetry-proto JSON mapping: ids as hex
// strings, 64-bit integers as decimal strings, enums as numbers). It needs no
// protobuf or gRPC dependency and reaches the same collector receiver, decoder
// and exporter as OTLP/gRPC.

type jAny struct {
	StringValue *string `json:"stringValue,omitempty"`
	IntValue    *string `json:"intValue,omitempty"`
}

type jKV struct {
	Key   string `json:"key"`
	Value jAny   `json:"value"`
}

func jAttrs(a []Attr) []jKV {
	out := make([]jKV, 0, len(a))
	for _, x := range sortedAttrs(a) {
		kv := jKV{Key: x.Key}
		if x.IsInt {
			v := strconv.FormatInt(x.Int, 10)
			kv.Value.IntValue = &v
		} else {
			v := x.Str
			kv.Value.StringValue = &v
		}
		out = append(out, kv)
	}
	return out
}

type jResource struct {
	Attributes []jKV `json:"attributes"`
}

type jScope struct {
	Name       string `json:"name"`
	Version    string `json:"version,omitempty"`
	Attributes []jKV  `json:"attributes,omitempty"`
}

func u64(v uint64) string { return strconv.FormatUint(v, 10) }

// ---- traces

type jEvent struct {
	TimeUnixNano string `json:"timeUnixNano"`
	Name         string `json:"name"`
	Attributes   []jKV  `json:"attributes,omitempty"`
}

type jLink struct {
	TraceID    string `json:"traceId"`
	SpanID     string `json:"spanId"`
	TraceState string `json:"traceState,omitempty"`
	Attributes []jKV  `json:"attributes,omitempty"`
}

type jStatus struct {
	Code    int32  `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
}

type jSpan struct {
	TraceID           string   `json:"traceId"`
	SpanID            string   `json:"spanId"`
	ParentSpanID      string   `json:"parentSpanId,omitempty"`
	TraceState        string   `json:"traceState,omitempty"`
	Name              string   `json:"name"`
	Kind              int32    `json:"kind"`
	StartTimeUnixNano string   `json:"startTimeUnixNano"`
	EndTimeUnixNano   string   `json:"endTimeUnixNano"`
	Attributes        []jKV    `json:"attributes,omitempty"`
	Events            []jEvent `json:"events,omitempty"`
	Links             []jLink  `json:"links,omitempty"`
	Status            jStatus  `json:"status"`
}

type jScopeSpans struct {
	Scope jScope  `json:"scope"`
	Spans []jSpan `json:"spans"`
}

type jResourceSpans struct {
	Resource   jResource     `json:"resource"`
	SchemaURL  string        `json:"schemaUrl,omitempty"`
	ScopeSpans []jScopeSpans `json:"scopeSpans"`
}

// EncodeTraces renders spans as an ExportTraceServiceRequest, grouped by resource
// and scope in first-seen order.
func EncodeTraces(g *Graph, spans []Span) ([]byte, error) {
	var out struct {
		ResourceSpans []*jResourceSpans `json:"resourceSpans"`
	}
	type key struct{ res, scope int }
	rsIdx := map[int]*jResourceSpans{}
	ssIdx := map[key]int{}
	for _, s := range spans {
		rs, ok := rsIdx[s.Resource]
		if !ok {
			r := g.resources[s.Resource]
			rs = &jResourceSpans{Resource: jResource{Attributes: jAttrs(r.Attrs)}, SchemaURL: r.SchemaURL}
			rsIdx[s.Resource] = rs
			out.ResourceSpans = append(out.ResourceSpans, rs)
		}
		k := key{s.Resource, s.Scope}
		si, ok := ssIdx[k]
		if !ok {
			sc := g.scopes[s.Scope]
			rs.ScopeSpans = append(rs.ScopeSpans, jScopeSpans{Scope: jScope{Name: sc.Name, Version: sc.Version, Attributes: jAttrs(sc.Attrs)}})
			si = len(rs.ScopeSpans) - 1
			ssIdx[k] = si
		}
		js := jSpan{TraceID: hexOrEmpty(s.TraceID[:]), SpanID: hexOrEmpty(s.SpanID[:]), ParentSpanID: hexOrEmpty(s.ParentSpanID[:]),
			TraceState: s.TraceState, Name: s.Name, Kind: int32(s.Kind), StartTimeUnixNano: u64(s.StartNs), EndTimeUnixNano: u64(s.EndNs),
			Attributes: jAttrs(s.Attrs), Status: jStatus{Code: int32(s.Status), Message: s.StatusMessage}}
		for _, e := range s.Events {
			js.Events = append(js.Events, jEvent{TimeUnixNano: u64(e.TimeNs), Name: e.Name, Attributes: jAttrs(e.Attrs)})
		}
		for _, l := range s.Links {
			js.Links = append(js.Links, jLink{TraceID: hexOrEmpty(l.TraceID[:]), SpanID: hexOrEmpty(l.SpanID[:]), TraceState: l.TraceState, Attributes: jAttrs(l.Attrs)})
		}
		rs.ScopeSpans[si].Spans = append(rs.ScopeSpans[si].Spans, js)
	}
	return json.Marshal(out)
}

// ---- logs

type jLog struct {
	TimeUnixNano         string `json:"timeUnixNano"`
	ObservedTimeUnixNano string `json:"observedTimeUnixNano"`
	SeverityNumber       int32  `json:"severityNumber"`
	SeverityText         string `json:"severityText"`
	Body                 jAny   `json:"body"`
	Attributes           []jKV  `json:"attributes,omitempty"`
	Flags                uint32 `json:"flags,omitempty"`
	TraceID              string `json:"traceId,omitempty"`
	SpanID               string `json:"spanId,omitempty"`
	EventName            string `json:"eventName,omitempty"`
}

type jScopeLogs struct {
	Scope      jScope `json:"scope"`
	SchemaURL  string `json:"schemaUrl,omitempty"`
	LogRecords []jLog `json:"logRecords"`
}

type jResourceLogs struct {
	Resource  jResource    `json:"resource"`
	SchemaURL string       `json:"schemaUrl,omitempty"`
	ScopeLogs []jScopeLogs `json:"scopeLogs"`
}

// EncodeLogs renders records as an ExportLogsServiceRequest.
func EncodeLogs(g *Graph, recs []LogRecord) ([]byte, error) {
	var out struct {
		ResourceLogs []*jResourceLogs `json:"resourceLogs"`
	}
	type key struct{ res, scope int }
	rlIdx := map[int]*jResourceLogs{}
	slIdx := map[key]int{}
	for _, r := range recs {
		rl, ok := rlIdx[r.Resource]
		if !ok {
			res := g.resources[r.Resource]
			rl = &jResourceLogs{Resource: jResource{Attributes: jAttrs(res.Attrs)}, SchemaURL: res.SchemaURL}
			rlIdx[r.Resource] = rl
			out.ResourceLogs = append(out.ResourceLogs, rl)
		}
		k := key{r.Resource, r.Scope}
		si, ok := slIdx[k]
		if !ok {
			sc := g.logScopes[r.Scope]
			rl.ScopeLogs = append(rl.ScopeLogs, jScopeLogs{Scope: jScope{Name: sc.Name, Version: sc.Version, Attributes: jAttrs(sc.Attrs)}, SchemaURL: sc.SchemaURL})
			si = len(rl.ScopeLogs) - 1
			slIdx[k] = si
		}
		body := r.Body
		rl.ScopeLogs[si].LogRecords = append(rl.ScopeLogs[si].LogRecords, jLog{
			TimeUnixNano: u64(r.TimeNs), ObservedTimeUnixNano: u64(r.ObservedTimeNs), SeverityNumber: r.SeverityNumber,
			SeverityText: r.SeverityText, Body: jAny{StringValue: &body}, Attributes: jAttrs(r.Attrs), Flags: r.Flags,
			TraceID: hexOrEmpty(r.TraceID[:]), SpanID: hexOrEmpty(r.SpanID[:]), EventName: r.EventName,
		})
	}
	return json.Marshal(out)
}

// ---- metrics

type jNumberDP struct {
	Attributes        []jKV    `json:"attributes,omitempty"`
	StartTimeUnixNano string   `json:"startTimeUnixNano,omitempty"`
	TimeUnixNano      string   `json:"timeUnixNano"`
	AsDouble          *float64 `json:"asDouble,omitempty"`
	AsInt             *string  `json:"asInt,omitempty"`
}

type jHistDP struct {
	Attributes        []jKV     `json:"attributes,omitempty"`
	StartTimeUnixNano string    `json:"startTimeUnixNano"`
	TimeUnixNano      string    `json:"timeUnixNano"`
	Count             string    `json:"count"`
	Sum               float64   `json:"sum"`
	BucketCounts      []string  `json:"bucketCounts"`
	ExplicitBounds    []float64 `json:"explicitBounds"`
}

type jBuckets struct {
	Offset       int32    `json:"offset"`
	BucketCounts []string `json:"bucketCounts"`
}

type jExpDP struct {
	Attributes        []jKV    `json:"attributes,omitempty"`
	StartTimeUnixNano string   `json:"startTimeUnixNano"`
	TimeUnixNano      string   `json:"timeUnixNano"`
	Count             string   `json:"count"`
	Sum               float64  `json:"sum"`
	Scale             int32    `json:"scale"`
	ZeroCount         string   `json:"zeroCount"`
	Positive          jBuckets `json:"positive"`
}

type jQuantile struct {
	Quantile float64 `json:"quantile"`
	Value    float64 `json:"value"`
}

type jSummaryDP struct {
	Attributes        []jKV       `json:"attributes,omitempty"`
	StartTimeUnixNano string      `json:"startTimeUnixNano"`
	TimeUnixNano      string      `json:"timeUnixNano"`
	Count             string      `json:"count"`
	Sum               float64     `json:"sum"`
	QuantileValues    []jQuantile `json:"quantileValues"`
}

type jMetric struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Unit        string `json:"unit,omitempty"`
	Gauge       *struct {
		DataPoints []jNumberDP `json:"dataPoints"`
	} `json:"gauge,omitempty"`
	Sum *struct {
		DataPoints             []jNumberDP `json:"dataPoints"`
		AggregationTemporality int32       `json:"aggregationTemporality"`
		IsMonotonic            bool        `json:"isMonotonic,omitempty"`
	} `json:"sum,omitempty"`
	Histogram *struct {
		DataPoints             []jHistDP `json:"dataPoints"`
		AggregationTemporality int32     `json:"aggregationTemporality"`
	} `json:"histogram,omitempty"`
	ExponentialHistogram *struct {
		DataPoints             []jExpDP `json:"dataPoints"`
		AggregationTemporality int32    `json:"aggregationTemporality"`
	} `json:"exponentialHistogram,omitempty"`
	Summary *struct {
		DataPoints []jSummaryDP `json:"dataPoints"`
	} `json:"summary,omitempty"`
}

type jScopeMetrics struct {
	Scope   jScope     `json:"scope"`
	Metrics []*jMetric `json:"metrics"`
}

type jResourceMetrics struct {
	Resource     jResource        `json:"resource"`
	SchemaURL    string           `json:"schemaUrl,omitempty"`
	ScopeMetrics []*jScopeMetrics `json:"scopeMetrics"`
}

func strs(xs []uint64) []string {
	out := make([]string, len(xs))
	for i, x := range xs {
		out[i] = u64(x)
	}
	return out
}

// EncodeMetrics renders points as an ExportMetricsServiceRequest, one Metric per
// (resource, scope, metric name).
func EncodeMetrics(g *Graph, pts []MetricPoint) ([]byte, error) {
	var out struct {
		ResourceMetrics []*jResourceMetrics `json:"resourceMetrics"`
	}
	type skey struct{ res, scope int }
	type mkey struct {
		res, scope int
		name       string
	}
	rmIdx := map[int]*jResourceMetrics{}
	smIdx := map[skey]*jScopeMetrics{}
	mIdx := map[mkey]*jMetric{}
	for _, p := range pts {
		rm, ok := rmIdx[p.Resource]
		if !ok {
			res := g.resources[p.Resource]
			rm = &jResourceMetrics{Resource: jResource{Attributes: jAttrs(res.Attrs)}, SchemaURL: res.SchemaURL}
			rmIdx[p.Resource] = rm
			out.ResourceMetrics = append(out.ResourceMetrics, rm)
		}
		sm, ok := smIdx[skey{p.Resource, p.Scope}]
		if !ok {
			sc := g.scopes[p.Scope]
			sm = &jScopeMetrics{Scope: jScope{Name: sc.Name, Version: sc.Version, Attributes: jAttrs(sc.Attrs)}}
			smIdx[skey{p.Resource, p.Scope}] = sm
			rm.ScopeMetrics = append(rm.ScopeMetrics, sm)
		}
		m, ok := mIdx[mkey{p.Resource, p.Scope, p.Name}]
		if !ok {
			m = &jMetric{Name: p.Name, Description: p.Description, Unit: p.Unit}
			mIdx[mkey{p.Resource, p.Scope, p.Name}] = m
			sm.Metrics = append(sm.Metrics, m)
		}
		attrs := jAttrs(p.Attrs)
		start := ""
		if p.StartNs != 0 {
			start = u64(p.StartNs)
		}
		number := func() jNumberDP {
			dp := jNumberDP{Attributes: attrs, StartTimeUnixNano: start, TimeUnixNano: u64(p.TimeNs)}
			if p.IsInt {
				v := strconv.FormatInt(int64(p.Value), 10)
				dp.AsInt = &v
			} else {
				v := p.Value
				dp.AsDouble = &v
			}
			return dp
		}
		switch p.Type {
		case Gauge:
			if m.Gauge == nil {
				m.Gauge = &struct {
					DataPoints []jNumberDP `json:"dataPoints"`
				}{}
			}
			m.Gauge.DataPoints = append(m.Gauge.DataPoints, number())
		case Sum:
			if m.Sum == nil {
				m.Sum = &struct {
					DataPoints             []jNumberDP `json:"dataPoints"`
					AggregationTemporality int32       `json:"aggregationTemporality"`
					IsMonotonic            bool        `json:"isMonotonic,omitempty"`
				}{AggregationTemporality: p.Temporality, IsMonotonic: p.Monotonic}
			}
			m.Sum.DataPoints = append(m.Sum.DataPoints, number())
		case Histogram:
			if m.Histogram == nil {
				m.Histogram = &struct {
					DataPoints             []jHistDP `json:"dataPoints"`
					AggregationTemporality int32     `json:"aggregationTemporality"`
				}{AggregationTemporality: p.Temporality}
			}
			m.Histogram.DataPoints = append(m.Histogram.DataPoints, jHistDP{Attributes: attrs, StartTimeUnixNano: start, TimeUnixNano: u64(p.TimeNs),
				Count: u64(p.Count), Sum: p.SumValue, BucketCounts: strs(p.BucketCounts), ExplicitBounds: p.Bounds})
		case ExpHistogram:
			if m.ExponentialHistogram == nil {
				m.ExponentialHistogram = &struct {
					DataPoints             []jExpDP `json:"dataPoints"`
					AggregationTemporality int32    `json:"aggregationTemporality"`
				}{AggregationTemporality: p.Temporality}
			}
			m.ExponentialHistogram.DataPoints = append(m.ExponentialHistogram.DataPoints, jExpDP{Attributes: attrs, StartTimeUnixNano: start,
				TimeUnixNano: u64(p.TimeNs), Count: u64(p.Count), Sum: p.SumValue, Scale: p.Scale, ZeroCount: u64(p.ZeroCount),
				Positive: jBuckets{Offset: p.PosOffset, BucketCounts: strs(p.BucketCounts)}})
		case Summary:
			if m.Summary == nil {
				m.Summary = &struct {
					DataPoints []jSummaryDP `json:"dataPoints"`
				}{}
			}
			dp := jSummaryDP{Attributes: attrs, StartTimeUnixNano: start, TimeUnixNano: u64(p.TimeNs), Count: u64(p.Count), Sum: p.SumValue}
			for qi, q := range p.Quantiles {
				dp.QuantileValues = append(dp.QuantileValues, jQuantile{Quantile: q, Value: p.QuantileVals[qi]})
			}
			m.Summary.DataPoints = append(m.Summary.DataPoints, dp)
		}
	}
	return json.Marshal(out)
}

// Sender posts OTLP/HTTP JSON requests to a collector.
type Sender struct {
	Endpoint string // for example http://127.0.0.1:4318
	Client   *http.Client
}

// Post sends one request body to /v1/<signal> and fails on a non-2xx answer or a
// partial-success rejection.
func (s *Sender) Post(ctx context.Context, signal string, body []byte) error {
	c := s.Client
	if c == nil {
		c = &http.Client{Timeout: 60 * time.Second}
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.Endpoint+"/v1/"+signal, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	msg, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
	if resp.StatusCode/100 != 2 {
		return fmt.Errorf("otlp %s: HTTP %d: %s", signal, resp.StatusCode, msg)
	}
	var ps struct {
		PartialSuccess map[string]any `json:"partialSuccess"`
	}
	if json.Unmarshal(msg, &ps) == nil && len(ps.PartialSuccess) > 0 {
		for k, v := range ps.PartialSuccess {
			if (k == "rejectedSpans" || k == "rejectedLogRecords" || k == "rejectedDataPoints") && fmt.Sprint(v) != "0" {
				return fmt.Errorf("otlp %s: partial success: %v", signal, ps.PartialSuccess)
			}
		}
	}
	return nil
}

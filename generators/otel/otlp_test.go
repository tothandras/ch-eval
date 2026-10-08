package otel

import (
	"encoding/json"
	"strings"
	"testing"
)

// The OTLP JSON mapping: ids are lowercase hex, 64-bit integers are decimal
// strings, enums are numbers, attributes are sorted by key.
func TestEncodeTracesShape(t *testing.T) {
	tg := testTraceGen(42)
	spans := tg.Trace(7)
	body, err := EncodeTraces(tg.Graph(), spans)
	if err != nil {
		t.Fatal(err)
	}
	var req struct {
		ResourceSpans []struct {
			ScopeSpans []struct {
				Spans []map[string]any `json:"spans"`
			} `json:"scopeSpans"`
		} `json:"resourceSpans"`
	}
	if err := json.Unmarshal(body, &req); err != nil {
		t.Fatal(err)
	}
	n := 0
	for _, rs := range req.ResourceSpans {
		for _, ss := range rs.ScopeSpans {
			for _, s := range ss.Spans {
				n++
				if id, _ := s["traceId"].(string); id != TraceIDHex(42, 7) {
					t.Fatalf("traceId %q, want %q", id, TraceIDHex(42, 7))
				}
				if _, ok := s["startTimeUnixNano"].(string); !ok {
					t.Fatal("startTimeUnixNano must be a decimal string")
				}
				if _, ok := s["kind"].(float64); !ok {
					t.Fatal("kind must be a number")
				}
			}
		}
	}
	if n != len(spans) {
		t.Fatalf("encoded %d spans, want %d", n, len(spans))
	}
	if !strings.Contains(string(body), `"intValue":"`) {
		t.Fatal("int attributes must be encoded as intValue strings")
	}
}

func TestAttrMapSortedLikeExporter(t *testing.T) {
	m := attrMap([]Attr{S("b", "2"), I("a", 1), S("c", "x")})
	var keys, vals []string
	it := m.Iterator()
	for it.Next() {
		keys = append(keys, it.Key().(string))
		vals = append(vals, it.Value().(string))
	}
	if strings.Join(keys, ",") != "a,b,c" || strings.Join(vals, ",") != "1,2,x" {
		t.Fatalf("got keys %v values %v; the exporter sorts by key and renders ints in decimal", keys, vals)
	}
	if hexOrEmpty(make([]byte, 8)) != "" || hexOrEmpty([]byte{0, 1}) != "0001" {
		t.Fatal("hexOrEmpty must mirror traceutil (empty for all-zero ids)")
	}
}

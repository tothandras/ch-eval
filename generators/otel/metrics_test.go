package otel

import (
	"reflect"
	"testing"
	"time"
)

func TestMetricsDeterministicMonotonic(t *testing.T) {
	g := NewGraph(DefaultGraph(42))
	cfg := MetricsConfig{TimeEnd: time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC), MaxPoints: 2_000_000}
	m, m2 := NewMetricGen(cfg, g), NewMetricGen(cfg, NewGraph(DefaultGraph(42)))
	n := uint64(m.Series())
	t.Logf("series %d, slots %d, points %d", m.Series(), m.slots, m.Points())
	byType := map[MetricType]int{}
	for i := uint64(0); i < m.Points(); i += 7 {
		p := m.Point(i)
		byType[p.Type]++
		if i%1001 == 0 && !reflect.DeepEqual(p, m2.Point(i)) {
			t.Fatalf("point %d not deterministic", i)
		}
		if p.Type == Histogram || p.Type == ExpHistogram {
			var s uint64
			for _, c := range p.BucketCounts {
				s += c
			}
			if s != p.Count {
				t.Fatalf("point %d: buckets sum %d != count %d", i, s, p.Count)
			}
		}
	}
	for _, typ := range MetricTypes {
		if byType[typ] == 0 {
			t.Fatalf("no %s points", typ.Table())
		}
	}
	// cumulative sums never decrease within a reset epoch
	for si := uint64(0); si < n; si += 13 {
		var prev MetricPoint
		for slot := uint64(0); slot < m.slots; slot++ {
			p := m.Point(slot*n + si)
			if p.Type == Sum && p.Monotonic && slot > 0 && p.StartNs == prev.StartNs && p.Value < prev.Value {
				t.Fatalf("series %d decreases at slot %d without a reset", si, slot)
			}
			prev = p
		}
	}
	hc := NewMetricGen(MetricsConfig{TimeEnd: cfg.TimeEnd, MaxPoints: 1, HighCardUsers: 100_000}, g)
	if hc.Series() < 100_000 {
		t.Fatalf("high-cardinality profile has %d series", hc.Series())
	}
}

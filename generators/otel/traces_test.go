package otel

import (
	"reflect"
	"testing"
	"time"
)

func testTraceGen(seed uint64) *TraceGen {
	g := NewGraph(DefaultGraph(seed))
	return NewTraceGen(TracesConfig{Graph: DefaultGraph(seed), TimeEnd: time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)}, g)
}

func TestTraceDeterministic(t *testing.T) {
	a, b := testTraceGen(42), testTraceGen(42)
	for i := uint64(0); i < 200; i++ {
		if !reflect.DeepEqual(a.Trace(i), b.Trace(i)) {
			t.Fatalf("trace %d differs between two generators with the same seed", i)
		}
	}
	// order independence: generating trace 7 alone equals generating it after others
	if !reflect.DeepEqual(testTraceGen(42).Trace(7), a.Trace(7)) {
		t.Fatal("trace depends on generation order")
	}
	if reflect.DeepEqual(testTraceGen(43).Trace(0), a.Trace(0)) {
		t.Fatal("different seeds produced the same trace")
	}
}

func TestTraceShape(t *testing.T) {
	tg := testTraceGen(42)
	from, to := tg.Window()
	var spans, errs, maxSpans int
	kinds := map[SpanKind]int{}
	for i := uint64(0); i < 2000; i++ {
		tr := tg.Trace(i)
		spans += len(tr)
		maxSpans = max(maxSpans, len(tr))
		ids := map[[8]byte]bool{}
		for k, s := range tr {
			kinds[s.Kind]++
			if s.Status == StatusError {
				errs++
			}
			if s.TraceID != tr[0].TraceID {
				t.Fatalf("trace %d span %d: wrong trace id", i, k)
			}
			if ids[s.SpanID] {
				t.Fatalf("trace %d: duplicate span id", i)
			}
			ids[s.SpanID] = true
			if k > 0 && !ids[s.ParentSpanID] {
				t.Fatalf("trace %d span %d: parent not emitted before child", i, k)
			}
			if s.EndNs < s.StartNs {
				t.Fatalf("trace %d span %d: ends before it starts", i, k)
			}
			seen := map[string]bool{}
			for _, a := range s.Attrs {
				if seen[a.Key] {
					t.Fatalf("trace %d span %d: duplicate attribute %s", i, k, a.Key)
				}
				seen[a.Key] = true
			}
		}
		if (tr[0].ParentSpanID != [8]byte{}) || tr[0].Kind != KindServer {
			t.Fatalf("trace %d: root must be a server span without parent", i)
		}
		st := time.Unix(0, int64(tr[0].StartNs))
		if st.Before(from) || !st.Before(to) {
			t.Fatalf("trace %d starts outside the window: %s", i, st)
		}
	}
	avg := float64(spans) / 2000
	t.Logf("avg spans/trace %.1f, max %d, error spans %.2f%%, kinds %v", avg, maxSpans, 100*float64(errs)/float64(spans), kinds)
	if avg < 5 || avg > 120 {
		t.Fatalf("average spans per trace %.1f outside the intended 5..120", avg)
	}
	for _, k := range []SpanKind{KindServer, KindClient, KindInternal} {
		if kinds[k] == 0 {
			t.Fatalf("no %s spans generated", k)
		}
	}
}

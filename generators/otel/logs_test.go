package otel

import (
	"reflect"
	"testing"
)

func TestLogsDeterministicAndCorrelated(t *testing.T) {
	tg := testTraceGen(42)
	lg := NewLogGen(LogsConfig{TimeEnd: tg.cfg.TimeEnd, TraceSpace: 1000, NeedleEvery: 1000}, tg)
	lg2 := NewLogGen(LogsConfig{TimeEnd: tg.cfg.TimeEnd, TraceSpace: 1000, NeedleEvery: 1000}, testTraceGen(42))
	var corr, needles int
	sev := map[string]int{}
	for j := uint64(0); j < 20000; j++ {
		a := lg.Record(j)
		if j%97 == 0 && !reflect.DeepEqual(a, lg2.Record(j)) {
			t.Fatalf("record %d not deterministic", j)
		}
		sev[a.SeverityText]++
		if a.TraceID != ([16]byte{}) {
			corr++
			found := false
			for i := uint64(0); i < 1000 && !found; i++ {
				if TraceID(42, i) == a.TraceID {
					for _, s := range tg.Trace(i) {
						found = found || (s.SpanID == a.SpanID && a.TimeNs >= s.StartNs && a.TimeNs <= s.EndNs)
					}
				}
			}
			if !found && corr < 50 {
				t.Fatalf("record %d: correlated ids/time do not match a generated span", j)
			}
		}
		if lg.Needle(j) != "" {
			needles++
		}
	}
	t.Logf("correlated %.1f%%, needles %d, severities %v", 100*float64(corr)/20000, needles, sev)
	if corr < 5000 || corr > 7000 {
		t.Fatalf("correlated fraction off: %d of 20000", corr)
	}
	if needles == 0 {
		t.Fatal("no needles planted")
	}
}

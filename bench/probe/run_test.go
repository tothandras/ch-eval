package probe

import (
	"testing"

	"github.com/tothandras/ch-eval/bench/target"
)

func TestExpectationPrecedence(t *testing.T) {
	p := Probe{Expect: map[string]string{
		"":                                     "default",
		"dialect-replicated":                   "dialect",
		"storage-s3":                           "storage",
		"altinity":                             "platform",
		"tier-local-cluster":                   "tier",
		"target-altinity/local-cluster-s3-cow": "target",
	}}
	m := target.Manifest{Name: "altinity/local-cluster-s3-cow", Platform: "altinity", Tier: "local-cluster", Storage: "s3", Dialect: "replicated"}
	for _, drop := range []string{"target-altinity/local-cluster-s3-cow", "tier-local-cluster", "altinity", "storage-s3", "dialect-replicated", ""} {
		want := p.Expect[drop]
		if got := p.Expectation(m); got != want {
			t.Fatalf("with %v: got %q, want %q", p.Expect, got, want)
		}
		delete(p.Expect, drop)
	}
	if got := p.Expectation(m); got != "" {
		t.Fatalf("no expectations: got %q", got)
	}
}

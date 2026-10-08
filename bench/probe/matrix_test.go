package probe

import (
	"os"
	"strings"
	"testing"
)

// The YAML must carry every row of the matrix: rendering it without probe results
// reproduces every table row of the Phase 1 draft byte for byte, except the last
// (probe/evidence) column, which Phase 2 extends.
func TestMatrixRoundTrip(t *testing.T) {
	m, err := LoadMatrix("../../docs/feature-matrix.yaml")
	if err != nil {
		t.Fatal(err)
	}
	for si := range m.Sections { // observations are appended to cells; compare the hypotheses
		for ri := range m.Sections[si].Rows {
			m.Sections[si].Rows[ri].Observations = nil
		}
	}
	out := m.Render(nil, "> **Status:** test")
	want, err := os.ReadFile("testdata/feature-matrix.phase1.md")
	if err != nil {
		t.Fatal(err)
	}
	for _, line := range strings.Split(string(want), "\n") {
		if i := strings.LastIndex(line, " | "); strings.HasPrefix(line, "| ") && i > 0 && !strings.Contains(out, line[:i]) {
			t.Errorf("table line missing from the rendered matrix:\n%s", line)
		}
	}
	rows := 0
	for _, s := range m.Sections {
		rows += len(s.Rows)
	}
	if rows < 85 {
		t.Fatalf("matrix has %d rows, the Phase 1 draft had 85", rows)
	}
}

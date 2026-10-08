package migrate

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	createRE      = regexp.MustCompile(`(?is)^\s*CREATE\s+(TABLE|MATERIALIZED\s+VIEW|VIEW|DATABASE|DICTIONARY|FUNCTION|ROLE|USER|ROW\s+POLICY|SETTINGS\s+PROFILE|QUOTA)\b`)
	createOKRE    = regexp.MustCompile(`(?is)^\s*CREATE\s+(OR\s+REPLACE\s+)?(TABLE|MATERIALIZED\s+VIEW|VIEW|DATABASE|DICTIONARY|FUNCTION|ROLE|USER|ROW\s+POLICY|SETTINGS\s+PROFILE|QUOTA)\s+(IF\s+NOT\s+EXISTS\b|.*\bOR\s+REPLACE\b)`)
	createOrRepl  = regexp.MustCompile(`(?is)^\s*CREATE\s+OR\s+REPLACE\b`)
	dropRE        = regexp.MustCompile(`(?is)^\s*DROP\s+(TABLE|VIEW|DATABASE|DICTIONARY|FUNCTION|ROLE|USER|ROW\s+POLICY|SETTINGS\s+PROFILE|QUOTA)\b`)
	dropOKRE      = regexp.MustCompile(`(?is)^\s*DROP\s+(TABLE|VIEW|DATABASE|DICTIONARY|FUNCTION|ROLE|USER|ROW\s+POLICY|SETTINGS\s+PROFILE|QUOTA)\s+IF\s+EXISTS\b`)
	alterAddRE    = regexp.MustCompile(`(?is)\bADD\s+(COLUMN|INDEX|PROJECTION)\s+`)
	alterAddOKRE  = regexp.MustCompile(`(?is)\bADD\s+(COLUMN|INDEX|PROJECTION)\s+IF\s+NOT\s+EXISTS\b`)
	alterDropRE   = regexp.MustCompile(`(?is)\bDROP\s+(COLUMN|INDEX|PROJECTION)\s+`)
	alterDropOKRE = regexp.MustCompile(`(?is)\bDROP\s+(COLUMN|INDEX|PROJECTION)\s+IF\s+EXISTS\b`)
	alterRE       = regexp.MustCompile(`(?is)^\s*ALTER\s+TABLE\b`)
	onClusterRE   = regexp.MustCompile(`(?i)\bON\s+CLUSTER\b`)
)

// Lint checks rendered statements for idempotency rules: CREATE ... IF NOT EXISTS
// (or OR REPLACE), DROP ... IF EXISTS, ALTER ADD/DROP ... IF [NOT] EXISTS, and no
// hand-written ON CLUSTER outside the replicated dialect (it must come from the
// onCluster helper). It returns one message per violation.
func Lint(stmts []string, topo Topology) []string {
	var out []string
	for i, s := range stmts {
		head := firstLine(s)
		switch {
		case createRE.MatchString(s) && !createOKRE.MatchString(s) && !createOrRepl.MatchString(s):
			out = append(out, fmt.Sprintf("statement %d: CREATE without IF NOT EXISTS / OR REPLACE: %s", i+1, head))
		case dropRE.MatchString(s) && !dropOKRE.MatchString(s):
			out = append(out, fmt.Sprintf("statement %d: DROP without IF EXISTS: %s", i+1, head))
		}
		if alterRE.MatchString(s) {
			if alterAddRE.MatchString(s) && !alterAddOKRE.MatchString(s) {
				out = append(out, fmt.Sprintf("statement %d: ALTER ADD without IF NOT EXISTS: %s", i+1, head))
			}
			if alterDropRE.MatchString(s) && !alterDropOKRE.MatchString(s) {
				out = append(out, fmt.Sprintf("statement %d: ALTER DROP without IF EXISTS: %s", i+1, head))
			}
		}
		if topo.Dialect != Replicated && onClusterRE.MatchString(s) {
			out = append(out, fmt.Sprintf("statement %d: ON CLUSTER outside the replicated dialect (use {{ onCluster }}): %s", i+1, head))
		}
	}
	return out
}

func firstLine(s string) string {
	l := strings.SplitN(strings.TrimSpace(s), "\n", 2)[0]
	if len(l) > 100 {
		l = l[:100] + "..."
	}
	return l
}

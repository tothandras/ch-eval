package scenario

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"

	"github.com/tothandras/ch-eval/bench/target"
)

// TestMigrationLifecycle runs up → verify → up (no-op) → down → up → verify for
// every workload scenario on each target named in CH_EVAL_IT_TARGETS (comma-
// separated <platform>/<tier>; the targets must be up). It works in a throwaway
// database so it never touches benchmark data. Skipped when the variable is unset.
func TestMigrationLifecycle(t *testing.T) {
	names := os.Getenv("CH_EVAL_IT_TARGETS")
	if names == "" {
		t.Skip("set CH_EVAL_IT_TARGETS=altinity/local-single,... to run against live targets")
	}
	ctx := context.Background()
	for _, name := range strings.Split(names, ",") {
		m, err := target.Load("../../targets", name)
		if err != nil {
			t.Fatal(err)
		}
		dsn, _, err := m.DSN()
		if err != nil {
			t.Fatal(err)
		}
		opts, err := clickhouse.ParseDSN(dsn)
		if err != nil {
			t.Fatal(err)
		}
		conn, err := clickhouse.Open(opts)
		if err != nil {
			t.Fatal(err)
		}
		scs, err := Discover("../../scenarios", "openmeter", nil)
		if err != nil {
			t.Fatal(err)
		}
		for _, sc := range scs {
			sc := sc
			sc.Config.Database = fmt.Sprintf("it_migrate_%d", time.Now().UnixNano()%1_000_000)
			t.Run(name+"/"+sc.Name, func(t *testing.T) {
				defer dropDatabase(ctx, t, conn, m, sc.Config.Database)
				runID := "it-" + sc.Config.Database
				applied, err := sc.Prepare(ctx, conn, m, false, runID, t.Logf)
				if err != nil || len(applied) != 1 {
					t.Fatalf("first up: applied=%d err=%v", len(applied), err)
				}
				mustVerify(ctx, t, sc, conn, m, true)
				if applied, err = sc.Prepare(ctx, conn, m, false, runID, t.Logf); err != nil || len(applied) != 0 {
					t.Fatalf("second up must be a no-op: applied=%d err=%v", len(applied), err)
				}
				// fresh = down everything, then up again
				applied, err = sc.Prepare(ctx, conn, m, true, runID, t.Logf)
				if err != nil || len(applied) != 2 || applied[0].Direction != "down" || applied[1].Direction != "up" {
					t.Fatalf("fresh: want [down up], got %d steps, err=%v", len(applied), err)
				}
				mustVerify(ctx, t, sc, conn, m, true)
			})
		}
		conn.Close()
	}
}

func mustVerify(ctx context.Context, t *testing.T, sc Scenario, conn clickhouse.Conn, m target.Manifest, ok bool) {
	t.Helper()
	problems, err := sc.Verify(ctx, conn, m)
	if err != nil {
		t.Fatal(err)
	}
	if ok && len(problems) > 0 {
		t.Fatalf("verify: %v", problems)
	}
}

func dropDatabase(ctx context.Context, t *testing.T, conn clickhouse.Conn, m target.Manifest, db string) {
	q := "DROP DATABASE IF EXISTS " + db + " SYNC"
	if m.Dialect == "replicated" {
		q = fmt.Sprintf("DROP DATABASE IF EXISTS %s ON CLUSTER '%s' SYNC", db, m.Cluster)
	}
	if err := conn.Exec(ctx, q); err != nil {
		t.Errorf("cleanup %s: %v", db, err)
	}
}

package experiments

import (
	"context"
	"fmt"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
)

// R6: ON CLUSTER DDL with a replica down, against the Replicated database engine.
func R6(ctx context.Context, e Env, admin driver.Conn) (Result, error) {
	r := Result{ID: "R6", Question: "How does DDL propagate when a replica is down (ON CLUSTER vs Replicated database)?", Target: e.M.Name, StartedAt: time.Now().UTC()}
	cl := e.M.Cluster
	for _, q := range []string{
		fmt.Sprintf("DROP DATABASE IF EXISTS %s ON CLUSTER '%s' SYNC", db, cl),
		fmt.Sprintf("DROP DATABASE IF EXISTS ch_eval_exp_repl ON CLUSTER '%s' SYNC", cl),
		fmt.Sprintf("CREATE DATABASE %s ON CLUSTER '%s'", db, cl),
		fmt.Sprintf("CREATE DATABASE ch_eval_exp_repl ON CLUSTER '%s' ENGINE = Replicated('/clickhouse/databases/ch_eval_exp_repl', '{shard}', '{replica}')", cl),
	} {
		if err := admin.Exec(ctx, q); err != nil {
			return r, fmt.Errorf("%s: %w", q, err)
		}
	}
	defer func() {
		_ = admin.Exec(context.Background(), fmt.Sprintf("DROP DATABASE IF EXISTS %s ON CLUSTER '%s' SYNC", db, cl))
		_ = admin.Exec(context.Background(), fmt.Sprintf("DROP DATABASE IF EXISTS ch_eval_exp_repl ON CLUSTER '%s' SYNC", cl))
	}()
	down := len(e.Control.Nodes()) - 1
	t := time.Now()
	err := e.Control.StopNode(down)
	r.step("stop node "+e.Control.Nodes()[down], t, err, "")
	t = time.Now()
	err = admin.Exec(clickhouse.Context(ctx, clickhouse.WithSettings(clickhouse.Settings{"distributed_ddl_task_timeout": 10})),
		fmt.Sprintf("CREATE TABLE %s.ddl_t ON CLUSTER '%s' (k UInt64) ENGINE = MergeTree ORDER BY k", db, cl))
	oc := r.step("CREATE TABLE ... ON CLUSTER with distributed_ddl_task_timeout=10", t, err, "")
	t = time.Now()
	err = admin.Exec(clickhouse.Context(ctx, clickhouse.WithSettings(clickhouse.Settings{"distributed_ddl_task_timeout": 10})),
		"CREATE TABLE ch_eval_exp_repl.ddl_t (k UInt64) ENGINE = ReplicatedMergeTree ORDER BY k")
	rd := r.step("CREATE TABLE in a Replicated database (no ON CLUSTER), distributed_ddl_task_timeout=10", t, err, "")
	t = time.Now()
	_ = e.Control.StartNode(down)
	back, err := e.waitNode(ctx, down, 2*time.Minute)
	r.step("restart node", t, err, fmt.Sprintf("%.0fms", float64(back.Microseconds())/1000))
	c, err := e.node(ctx, down)
	if err != nil {
		return r, err
	}
	defer c.Close()
	for _, name := range []string{db + ".ddl_t", "ch_eval_exp_repl.ddl_t"} {
		start := time.Now()
		seen := false
		for time.Since(start) < 2*time.Minute {
			var n uint8
			if err := c.QueryRow(ctx, "EXISTS TABLE "+name).Scan(&n); err == nil && n == 1 {
				seen = true
				break
			}
			time.Sleep(200 * time.Millisecond)
		}
		v := "never (2 min)"
		if seen {
			v = fmt.Sprintf("%.0fms", float64(time.Since(start).Microseconds())/1000)
		}
		r.step("time after restart until "+name+" exists on the restarted node", start, nil, v)
	}
	r.Findings = append(r.Findings,
		fmt.Sprintf("ON CLUSTER with a replica down returned %s after %.0f ms; see the steps for when the restarted replica executed the queued task.", oc.Outcome, oc.Millis),
		fmt.Sprintf("In a Replicated database the same DDL returned %s after %.0f ms; see the steps for when the restarted replica replayed it from the database log.", rd.Outcome, rd.Millis))
	return r, nil
}

// R3: Keeper quorum loss on a replicated table.
func R3(ctx context.Context, e Env, admin driver.Conn) (Result, error) {
	r := Result{ID: "R3", Question: "What happens when Keeper loses quorum?", Target: e.M.Name, StartedAt: time.Now().UTC()}
	cl := e.M.Cluster
	for _, q := range []string{
		fmt.Sprintf("DROP DATABASE IF EXISTS %s ON CLUSTER '%s' SYNC", db, cl),
		fmt.Sprintf("CREATE DATABASE %s ON CLUSTER '%s'", db, cl),
		fmt.Sprintf("CREATE TABLE %s.kq ON CLUSTER '%s' (k UInt64) ENGINE = ReplicatedMergeTree('/clickhouse/tables/{shard}/ch_eval_exp/kq', '{replica}') ORDER BY k", db, cl),
	} {
		if err := admin.Exec(ctx, q); err != nil {
			return r, err
		}
	}
	defer func() {
		_ = admin.Exec(context.Background(), fmt.Sprintf("DROP DATABASE IF EXISTS %s ON CLUSTER '%s' SYNC", db, cl))
	}()
	ins := func(name string, k int) Step {
		t := time.Now()
		ictx, cancel := context.WithTimeout(ctx, 60*time.Second)
		defer cancel()
		err := admin.Exec(clickhouse.Context(ictx, clickhouse.WithSettings(clickhouse.Settings{"async_insert": 0, "insert_keeper_max_retries": 2})),
			fmt.Sprintf("INSERT INTO %s.kq VALUES (%d)", db, k))
		return r.step(name, t, err, "")
	}
	ins("insert, 3/3 Keepers up", 1)
	_ = e.Control.StopKeeper(3)
	one := ins("insert, 2/3 Keepers up (one lost)", 2)
	_ = e.Control.StopKeeper(2)
	time.Sleep(2 * time.Second)
	lost := ins("insert, 1/3 Keepers up (quorum lost)", 3)
	t := time.Now()
	n, err := count(ctx, admin, fmt.Sprintf("SELECT count() FROM %s.kq", db), nil)
	sel := r.step("SELECT during quorum loss", t, err, fmt.Sprint(n))
	var ro uint8
	_ = admin.QueryRow(ctx, fmt.Sprintf("SELECT is_readonly FROM system.replicas WHERE database = '%s' AND table = 'kq'", db)).Scan(&ro)
	r.step("system.replicas.is_readonly during quorum loss", time.Now(), nil, fmt.Sprint(ro))
	_ = e.Control.StartKeeper(2)
	_ = e.Control.StartKeeper(3)
	t = time.Now()
	recovered := "never (3 min)"
	for time.Since(t) < 3*time.Minute {
		ictx, cancel := context.WithTimeout(ctx, 5*time.Second)
		err := admin.Exec(clickhouse.Context(ictx, clickhouse.WithSettings(clickhouse.Settings{"async_insert": 0})), fmt.Sprintf("INSERT INTO %s.kq VALUES (4)", db))
		cancel()
		if err == nil {
			recovered = fmt.Sprintf("%.0fms", float64(time.Since(t).Microseconds())/1000)
			break
		}
		time.Sleep(500 * time.Millisecond)
	}
	r.step("time from Keeper restart until inserts succeed", t, nil, recovered)
	r.Findings = append(r.Findings,
		fmt.Sprintf("Losing one of three Keepers: insert %s.", one.Outcome),
		fmt.Sprintf("Losing quorum: insert %s after %.0f ms; SELECT %s; tables went read-only (is_readonly=%d). Inserts recovered %s after the Keepers restarted.", lost.Outcome, lost.Millis, sel.Outcome, ro, recovered))
	return r, nil
}

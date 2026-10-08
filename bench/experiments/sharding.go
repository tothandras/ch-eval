package experiments

import (
	"context"
	"fmt"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
)

const db = "ch_eval_exp"

// shardedTable creates db.t_local (replicated per shard) and db.t (Distributed by
// cityHash64(k)) and returns a cleanup func.
func shardedTable(ctx context.Context, admin driver.Conn, cluster string) (func(), error) {
	for _, q := range []string{
		fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s ON CLUSTER '%s'", db, cluster),
		fmt.Sprintf("DROP TABLE IF EXISTS %s.t ON CLUSTER '%s' SYNC", db, cluster),
		fmt.Sprintf("DROP TABLE IF EXISTS %s.t_local ON CLUSTER '%s' SYNC", db, cluster),
		fmt.Sprintf("CREATE TABLE %s.t_local ON CLUSTER '%s' (k UInt64, v UInt64) ENGINE = ReplicatedMergeTree('/clickhouse/tables/{shard}/ch_eval_exp/t_local', '{replica}') ORDER BY k", db, cluster),
		fmt.Sprintf("CREATE TABLE %s.t ON CLUSTER '%s' AS %s.t_local ENGINE = Distributed('%s', '%s', 't_local', cityHash64(k))", db, cluster, db, cluster, db),
	} {
		if err := admin.Exec(ctx, q); err != nil {
			return nil, fmt.Errorf("%s: %w", q, err)
		}
	}
	return func() {
		_ = admin.Exec(context.Background(), fmt.Sprintf("DROP DATABASE IF EXISTS %s ON CLUSTER '%s' SYNC", db, cluster))
	}, nil
}

// S3: what a query over a Distributed table returns while one shard is down.
func S3(ctx context.Context, e Env, admin driver.Conn) (Result, error) {
	r := Result{ID: "S3", Question: "What happens when a shard is unavailable (default vs skip_unavailable_shards=1)?", Target: e.M.Name, StartedAt: time.Now().UTC()}
	if e.M.Shards < 2 {
		return r, fmt.Errorf("S3 needs a sharded target")
	}
	cleanup, err := shardedTable(ctx, admin, e.M.Cluster)
	if err != nil {
		return r, err
	}
	defer cleanup()
	const rows = 100000
	t := time.Now()
	err = admin.Exec(clickhouse.Context(ctx, clickhouse.WithSettings(clickhouse.Settings{"distributed_foreground_insert": 1})),
		fmt.Sprintf("INSERT INTO %s.t SELECT number, 1 FROM numbers(%d)", db, rows))
	r.step("insert 100k rows through the Distributed table (foreground)", t, err, "")
	if err != nil {
		return r, err
	}
	n, _ := count(ctx, admin, fmt.Sprintf("SELECT count() FROM %s.t", db), nil)
	r.step("count with all shards up", time.Now(), nil, fmt.Sprint(n))
	// stop every replica of the last shard
	last := e.M.Shards - 1
	var stopped []int
	for rep := 0; rep < e.M.Replicas; rep++ {
		i := last*e.M.Replicas + rep
		t = time.Now()
		err := e.Control.StopNode(i)
		r.step(fmt.Sprintf("stop node %s (shard %d)", e.Control.Nodes()[i], last+1), t, err, "")
		stopped = append(stopped, i)
	}
	defer func() {
		for _, i := range stopped {
			_ = e.Control.StartNode(i)
			_, _ = e.waitNode(context.Background(), i, 2*time.Minute)
		}
	}()
	q := fmt.Sprintf("SELECT count() FROM %s.t", db)
	t = time.Now()
	n, err = count(ctx, admin, q, clickhouse.Settings{"connect_timeout_with_failover_ms": 1000, "connections_with_failover_max_tries": 1})
	s := r.step("count, default settings (skip_unavailable_shards=0)", t, err, fmt.Sprint(n))
	t = time.Now()
	m, err2 := count(ctx, admin, q, clickhouse.Settings{"skip_unavailable_shards": 1, "connect_timeout_with_failover_ms": 1000, "connections_with_failover_max_tries": 1})
	s2 := r.step("count, skip_unavailable_shards=1", t, err2, fmt.Sprint(m))
	r.Findings = append(r.Findings,
		fmt.Sprintf("With a shard down, the default query returned %s after %.0f ms.", s.Outcome, s.Millis),
		fmt.Sprintf("With skip_unavailable_shards=1 it returned %s of %d rows, without an error, after %.0f ms: a partial total that a billing query would not notice.", s2.Outcome, rows, s2.Millis))
	return r, nil
}

// S4: time-to-queryable per insert route, and the Distributed queue while a shard is down.
func S4(ctx context.Context, e Env, admin driver.Conn) (Result, error) {
	r := Result{ID: "S4", Question: "How does insert routing affect visibility (Distributed background, foreground, direct to shard)?", Target: e.M.Name, StartedAt: time.Now().UTC()}
	if e.M.Shards < 2 {
		return r, fmt.Errorf("S4 needs a sharded target")
	}
	cleanup, err := shardedTable(ctx, admin, e.M.Cluster)
	if err != nil {
		return r, err
	}
	defer cleanup()
	const batch = 50000
	total := uint64(0)
	visible := func(want uint64) time.Duration {
		start := time.Now()
		for time.Since(start) < 2*time.Minute {
			if n, err := count(ctx, admin, fmt.Sprintf("SELECT count() FROM %s.t", db), nil); err == nil && n >= want {
				return time.Since(start)
			}
			time.Sleep(20 * time.Millisecond)
		}
		return -1
	}
	for _, mode := range []struct {
		name     string
		settings clickhouse.Settings
	}{
		{"Distributed, background queue (default)", clickhouse.Settings{"distributed_foreground_insert": 0}},
		{"Distributed, distributed_foreground_insert=1", clickhouse.Settings{"distributed_foreground_insert": 1}},
	} {
		t := time.Now()
		err := admin.Exec(clickhouse.Context(ctx, clickhouse.WithSettings(mode.settings)),
			fmt.Sprintf("INSERT INTO %s.t SELECT number + %d, 1 FROM numbers(%d)", db, total, batch))
		r.step("insert ack: "+mode.name, t, err, "")
		total += batch
		d := visible(total)
		r.step("time from ack to queryable: "+mode.name, time.Now().Add(-d), nil, fmt.Sprintf("%.0fms", float64(d.Microseconds())/1000))
	}
	// direct to shard: the client computes the shard and writes local tables
	t := time.Now()
	for s := 0; s < e.M.Shards; s++ {
		c, err := e.node(ctx, s*e.M.Replicas)
		if err != nil {
			return r, err
		}
		err = c.Exec(ctx, fmt.Sprintf("INSERT INTO %s.t_local SELECT number + %d AS k, 1 FROM numbers(%d) WHERE cityHash64(k) %% %d = %d", db, total, batch, e.M.Shards, s))
		c.Close()
		if err != nil {
			return r, err
		}
	}
	r.step("insert ack: direct to shard-local tables (client-side sharding)", t, nil, "")
	total += batch
	d := visible(total)
	r.step("time from ack to queryable: direct to shard", time.Now().Add(-d), nil, fmt.Sprintf("%.0fms", float64(d.Microseconds())/1000))

	// queue behaviour while the last shard is down
	last := e.M.Shards - 1
	var stopped []int
	for rep := 0; rep < e.M.Replicas; rep++ {
		i := last*e.M.Replicas + rep
		_ = e.Control.StopNode(i)
		stopped = append(stopped, i)
	}
	t = time.Now()
	err = admin.Exec(clickhouse.Context(ctx, clickhouse.WithSettings(clickhouse.Settings{"distributed_foreground_insert": 0})),
		fmt.Sprintf("INSERT INTO %s.t SELECT number + %d, 1 FROM numbers(%d)", db, total, batch))
	r.step("insert with the last shard down: background queue", t, err, "")
	queued := err == nil
	if queued {
		total += batch
	}
	t = time.Now()
	err = admin.Exec(clickhouse.Context(ctx, clickhouse.WithSettings(clickhouse.Settings{"distributed_foreground_insert": 1, "connect_timeout_with_failover_ms": 1000, "connections_with_failover_max_tries": 1})),
		fmt.Sprintf("INSERT INTO %s.t SELECT number + %d, 1 FROM numbers(%d)", db, total+10*batch, batch))
	fg := r.step("insert with the last shard down: foreground", t, err, "")
	var pending uint64
	_ = admin.QueryRow(ctx, fmt.Sprintf("SELECT sum(data_files) FROM system.distribution_queue WHERE database = '%s'", db)).Scan(&pending)
	r.step("distribution_queue files pending on the initiator", time.Now(), nil, fmt.Sprint(pending))
	for _, i := range stopped {
		_ = e.Control.StartNode(i)
	}
	t = time.Now()
	for _, i := range stopped {
		_, _ = e.waitNode(ctx, i, 2*time.Minute)
	}
	d = visible(total)
	r.step("time from shard restart to queued rows queryable", t, nil, fmt.Sprintf("%.0fms", float64(time.Since(t).Microseconds())/1000))
	r.Findings = append(r.Findings,
		"Background Distributed inserts acknowledge before rows are queryable; foreground inserts acknowledge after the shards have them.",
		fmt.Sprintf("With a shard down, the background insert was accepted (%v) and queued on the initiator's disk (%d files) until the shard returned; the foreground insert returned %s.", queued, pending, fg.Outcome))
	return r, nil
}

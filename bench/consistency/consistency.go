// Package consistency implements the replica-consistency experiment (design D7,
// replica-consistency spec): how soon an acknowledged insert is visible on each
// replica, and what concurrent readers behind a load balancer observe (stale
// totals, totals that go backwards) under each consistency policy.
package consistency

import (
	"context"
	"fmt"
	"math/rand/v2"
	"os/exec"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
)

// Policy is one consistency policy under test.
type Policy string

const (
	Default     Policy = "default"      // engine defaults
	Quorum      Policy = "quorum"       // RMT: insert_quorum=<replicas>, insert_quorum_parallel=0, select_sequential_consistency=1
	SyncReplica Policy = "sync-replica" // RMT: SYSTEM SYNC REPLICA ... LIGHTWEIGHT before each read
	SeqConsist  Policy = "seq"          // SMT: select_sequential_consistency=1
)

// Config is one experiment run.
type Config struct {
	Database   string
	Table      string
	Cluster    string // cluster for clusterAllReplicas and ON CLUSTER ("" on Cloud: no ON CLUSTER)
	Cloud      bool   // SharedMergeTree target: no ON CLUSTER, Cloud converts the engine
	Replicas   int
	Policy     Policy
	AsyncInsert bool
	Duration   time.Duration
	BatchEvery time.Duration // one batch per interval
	BatchRows  int
	Readers    int
	ReadEvery  time.Duration
	PollEvery  time.Duration
	// RestartCmd, when set, runs at half time (for example a docker restart of
	// one replica); the run measures behaviour across the restart.
	RestartCmd string
	Logf       func(string, ...any)
}

// Result is the run outcome.
type Result struct {
	Config        Config                 `json:"config"`
	StartedAt     time.Time              `json:"started_at"`
	Batches       int                    `json:"batches_acked"`
	InsertErrors  int                    `json:"insert_errors"`
	InsertMs      Dist                   `json:"insert_latency_ms"`
	LagMs         map[string]Dist        `json:"ack_to_visible_lag_ms"` // per replica host, plus "all"
	NotVisible    map[string]int         `json:"batches_never_visible"` // per replica, by run end
	Reads         int                    `json:"reads"`
	ReadErrors    int                    `json:"read_errors"`
	ReadMs        Dist                   `json:"read_latency_ms"`
	Stale         int                    `json:"stale_reads"`
	StaleRate     float64                `json:"stale_rate"`
	Violations    int                    `json:"monotonicity_violations"`
	ViolationsPer10k float64             `json:"violations_per_10k_reads"`
	DeficitRows   Dist                   `json:"stale_deficit_rows"`
	DeficitMs     Dist                   `json:"stale_deficit_ms"`
	PerNodeReads  map[string]int         `json:"reads_per_node"`
	RestartAt     *time.Time             `json:"restart_at,omitempty"`
	RestartOutput string                 `json:"restart_output,omitempty"`
	Notes         []string               `json:"notes,omitempty"`
}

// Dist summarises a sample.
type Dist struct {
	N   int     `json:"n"`
	P50 float64 `json:"p50"`
	P99 float64 `json:"p99"`
	Max float64 `json:"max"`
}

func dist(xs []float64) Dist {
	if len(xs) == 0 {
		return Dist{}
	}
	s := append([]float64(nil), xs...)
	sort.Float64s(s)
	q := func(p float64) float64 { return s[min(len(s)-1, int(p*float64(len(s))))] }
	return Dist{N: len(s), P50: q(0.5), P99: q(0.99), Max: s[len(s)-1]}
}

// Nodes are per-replica connections (readers pick one at random, as a load
// balancer would); Admin runs DDL and the visibility poller.
type Nodes struct {
	Admin driver.Conn
	Node  []driver.Conn
	Names []string
}

func (c Config) fq() string { return c.Database + "." + c.Table }

func (c Config) onCluster() string {
	if c.Cloud || c.Cluster == "" {
		return ""
	}
	return fmt.Sprintf(" ON CLUSTER '%s'", c.Cluster)
}

// Setup recreates the table: ReplicatedMergeTree on self-managed clusters,
// MergeTree on Cloud (which runs it as SharedMergeTree).
func Setup(ctx context.Context, n Nodes, c Config) error {
	engine := "ReplicatedMergeTree('/clickhouse/tables/ch_eval_consistency/{database}/{table}', '{replica}')"
	if c.Cloud {
		engine = "MergeTree"
	}
	for _, q := range []string{
		fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s%s", c.Database, c.onCluster()),
		fmt.Sprintf("DROP TABLE IF EXISTS %s%s SYNC", c.fq(), c.onCluster()),
		fmt.Sprintf(`CREATE TABLE %s%s (namespace String, batch_seq UInt64, row UInt32, value UInt64, inserted_at DateTime64(3))
ENGINE = %s ORDER BY (namespace, batch_seq, row)`, c.fq(), c.onCluster(), engine),
	} {
		if err := n.Admin.Exec(ctx, q); err != nil {
			return fmt.Errorf("setup: %w", err)
		}
	}
	return nil
}

func (c Config) insertSettings() clickhouse.Settings {
	s := clickhouse.Settings{"async_insert": 0}
	if c.AsyncInsert {
		s["async_insert"] = 1
		s["wait_for_async_insert"] = 1
	}
	if c.Policy == Quorum {
		s["insert_quorum"] = c.Replicas
		s["insert_quorum_parallel"] = 0
	}
	return s
}

func (c Config) readSettings() clickhouse.Settings {
	if c.Policy == Quorum || c.Policy == SeqConsist {
		return clickhouse.Settings{"select_sequential_consistency": 1}
	}
	return nil
}

type ackLog struct {
	mu   sync.Mutex
	at   []time.Time // at[k] = ack time of batch k+1
	rows int
}

func (a *ackLog) acked() (int, time.Time) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if len(a.at) == 0 {
		return 0, time.Time{}
	}
	return len(a.at), a.at[len(a.at)-1]
}

func (a *ackLog) ackTime(k int) (time.Time, bool) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if k < 1 || k > len(a.at) {
		return time.Time{}, false
	}
	return a.at[k-1], true
}

// Run executes the experiment.
func Run(ctx context.Context, n Nodes, c Config) (Result, error) {
	logf := c.Logf
	if logf == nil {
		logf = func(string, ...any) {}
	}
	res := Result{Config: c, StartedAt: time.Now().UTC(), LagMs: map[string]Dist{}, NotVisible: map[string]int{}, PerNodeReads: map[string]int{}}
	if err := Setup(ctx, n, c); err != nil {
		return res, err
	}
	runCtx, cancel := context.WithTimeout(ctx, c.Duration)
	defer cancel()
	acks := &ackLog{rows: c.BatchRows}
	var wg sync.WaitGroup
	var mu sync.Mutex
	var insertMs, readMs, deficitRows, deficitMs []float64

	// writer
	wg.Add(1)
	go func() {
		defer wg.Done()
		t := time.NewTicker(c.BatchEvery)
		defer t.Stop()
		ictx := clickhouse.Context(runCtx, clickhouse.WithSettings(c.insertSettings()))
		for seq := uint64(1); ; seq++ {
			select {
			case <-runCtx.Done():
				return
			case <-t.C:
			}
			start := time.Now()
			b, err := n.Admin.PrepareBatch(ictx, "INSERT INTO "+c.fq())
			if err == nil {
				for r := 0; r < c.BatchRows; r++ {
					if err = b.Append("ns", seq, uint32(r), uint64(1), start); err != nil {
						break
					}
				}
				if err == nil {
					err = b.Send()
				}
			}
			if err != nil {
				if runCtx.Err() != nil {
					return
				}
				mu.Lock()
				res.InsertErrors++
				mu.Unlock()
				logf("insert batch %d: %v", seq, err)
				seq-- // retry the same sequence number so the stream stays gap-free
				continue
			}
			now := time.Now()
			acks.mu.Lock()
			acks.at = append(acks.at, now)
			acks.mu.Unlock()
			mu.Lock()
			insertMs = append(insertMs, float64(now.Sub(start).Microseconds())/1000)
			mu.Unlock()
		}
	}()

	// per-replica visibility poller
	visible := map[string]int{} // host -> highest batch seen complete
	lags := map[string][]float64{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		q := fmt.Sprintf("SELECT hostName(), max(batch_seq), count() FROM clusterAllReplicas('%s', %s) GROUP BY 1 SETTINGS skip_unavailable_shards = 1", c.Cluster, c.fq())
		for runCtx.Err() == nil {
			now := time.Now()
			rows, err := n.Admin.Query(runCtx, q)
			if err == nil {
				for rows.Next() {
					var host string
					var maxSeq, cnt uint64
					if rows.Scan(&host, &maxSeq, &cnt) != nil {
						continue
					}
					// batch k is visible when every row of batches 1..k is there
					complete := int(min(maxSeq, cnt/uint64(c.BatchRows)))
					for k := visible[host] + 1; k <= complete; k++ {
						if at, ok := acks.ackTime(k); ok {
							lags[host] = append(lags[host], float64(now.Sub(at).Microseconds())/1000)
						}
					}
					if complete > visible[host] {
						visible[host] = complete
					}
				}
				rows.Close()
			}
			time.Sleep(c.PollEvery)
		}
	}()

	// readers through the "load balancer"
	for r := 0; r < c.Readers; r++ {
		wg.Add(1)
		go func(r int) {
			defer wg.Done()
			rng := rand.New(rand.NewPCG(uint64(r), 7))
			prev := uint64(0)
			for runCtx.Err() == nil {
				time.Sleep(c.ReadEvery)
				node := rng.IntN(len(n.Node))
				conn := n.Node[node]
				ackedBefore, _ := acks.acked()
				t0 := time.Now()
				var err error
				if c.Policy == SyncReplica {
					err = conn.Exec(runCtx, fmt.Sprintf("SYSTEM SYNC REPLICA %s LIGHTWEIGHT", c.fq()))
				}
				var total uint64
				if err == nil {
					rctx := runCtx
					if s := c.readSettings(); s != nil {
						rctx = clickhouse.Context(runCtx, clickhouse.WithSettings(s))
					}
					err = conn.QueryRow(rctx, "SELECT count() FROM "+c.fq()+" WHERE namespace = 'ns'").Scan(&total)
				}
				dur := time.Since(t0)
				mu.Lock()
				if err != nil {
					if runCtx.Err() == nil {
						res.ReadErrors++
					}
					mu.Unlock()
					continue
				}
				res.Reads++
				res.PerNodeReads[n.Names[node]]++
				readMs = append(readMs, float64(dur.Microseconds())/1000)
				want := uint64(ackedBefore * c.BatchRows)
				if total < want {
					res.Stale++
					deficitRows = append(deficitRows, float64(want-total))
					// the observed state ends at batch seen; the first missing batch was acked at ackTime(seen+1)
					seen := int(total / uint64(c.BatchRows))
					if at, ok := acks.ackTime(seen + 1); ok {
						deficitMs = append(deficitMs, float64(t0.Sub(at).Microseconds())/1000)
					}
				}
				if total < prev {
					res.Violations++
				}
				prev = total // spec: compared with the same reader's previous observation
				mu.Unlock()
			}
		}(r)
	}

	if c.RestartCmd != "" {
		wg.Add(1)
		go func() {
			defer wg.Done()
			select {
			case <-runCtx.Done():
				return
			case <-time.After(c.Duration / 2):
			}
			now := time.Now().UTC()
			res.RestartAt = &now
			out, err := exec.Command("sh", "-c", c.RestartCmd).CombinedOutput()
			res.RestartOutput = strings.TrimSpace(string(out))
			if err != nil {
				res.RestartOutput += " (" + err.Error() + ")"
			}
		}()
	}
	wg.Wait()

	// final visibility sweep: batches never seen on a replica by the end
	res.Batches, _ = acks.acked()
	var all []float64
	for host, l := range lags {
		res.LagMs[host] = dist(l)
		all = append(all, l...)
		res.NotVisible[host] = res.Batches - visible[host]
	}
	res.LagMs["all"] = dist(all)
	res.InsertMs = dist(insertMs)
	res.ReadMs = dist(readMs)
	res.DeficitRows = dist(deficitRows)
	res.DeficitMs = dist(deficitMs)
	if res.Reads > 0 {
		res.StaleRate = float64(res.Stale) / float64(res.Reads)
		res.ViolationsPer10k = 10000 * float64(res.Violations) / float64(res.Reads)
	}
	if c.AsyncInsert && c.Policy == Quorum {
		res.Notes = append(res.Notes, "insert_quorum with async inserts: see insert_errors")
	}
	return res, nil
}

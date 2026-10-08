package runner

import (
	"context"
	"fmt"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
)

// NewClusterCPUProbe is the cluster-aware variant of NewCPUProbe (ADR-0002
// amendment): on a multi-node target a query's CPU is spread over the initiator and
// the remote shards, and each node's query_log holds only its own threads. The
// log_comment setting is forwarded to remote sub-queries, so CPU is summed per
// initial_query_id across clusterAllReplicas(cluster, system.query_log) and the
// per-iteration totals feed the same p50/p95 statistics as the single-node probe.
func NewClusterCPUProbe(conn driver.Conn, cluster string) CPUProbe {
	return func(ctx context.Context, logComment string) (*CPUStats, error) {
		read := func() (*CPUStats, bool, error) {
			if err := conn.Exec(ctx, fmt.Sprintf("SYSTEM FLUSH LOGS ON CLUSTER '%s'", cluster)); err != nil {
				_ = conn.Exec(ctx, "SYSTEM FLUSH LOGS")
			}
			q := fmt.Sprintf(`SELECT
			    count() AS n,
			    quantile(0.5)(os) AS os_p50, quantile(0.95)(os) AS os_p95, max(os) AS os_max,
			    quantile(0.5)(rt) AS rt_p50, quantile(0.95)(rt) AS rt_p95,
			    avg(mem) AS mem_avg
			  FROM (
			    SELECT initial_query_id,
			      toFloat64(sum(ProfileEvents['OSCPUVirtualTimeMicroseconds'])) AS os,
			      toFloat64(sum(ProfileEvents['RealTimeMicroseconds'])) AS rt,
			      toFloat64(sum(memory_usage)) AS mem
			    FROM clusterAllReplicas('%s', system.query_log)
			    WHERE log_comment = ? AND type = 'QueryFinish'
			    GROUP BY initial_query_id
			  ) SETTINGS skip_unavailable_shards = 1`, cluster)
			var n uint64
			var osP50, osP95, osMax, rtP50, rtP95, memAvg float64
			if err := conn.QueryRow(ctx, q, logComment).Scan(&n, &osP50, &osP95, &osMax, &rtP50, &rtP95, &memAvg); err != nil {
				return nil, false, fmt.Errorf("cluster query_log read: %w", err)
			}
			if n == 0 {
				return nil, false, nil
			}
			if osMax > 0 {
				return &CPUStats{CPUp50Us: osP50, CPUp95Us: osP95, MemAvgBytes: memAvg, Source: "os_cpu_cluster"}, true, nil
			}
			return &CPUStats{CPUp50Us: rtP50, CPUp95Us: rtP95, MemAvgBytes: memAvg, Source: "real_time_cluster"}, true, nil
		}
		for attempt := 0; attempt < 2; attempt++ {
			stats, ok, err := read()
			if err != nil || ok {
				return stats, err
			}
			time.Sleep(300 * time.Millisecond)
		}
		return nil, nil
	}
}

package probe

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
)

// Go probes cover client-side behaviour with this harness's clickhouse-go
// (v2.46.0; OpenMeter pins v2.48.0, recorded as a caveat in the matrix).

func init() {
	registerGo("probes/client/protocols.go", "Native and HTTP interfaces answer (CL-01)",
		map[string]string{"": "value:native,http"}, protocols)
	registerGo("probes/client/compression.go", "lz4 and zstd native compression (CL-02)",
		map[string]string{"": "value:lz4,zstd"}, compression)
	registerGo("probes/client/async_insert.go", "clickhouse-go AsyncInsert(wait=true) makes rows visible on return (CL-03)",
		map[string]string{"": "value:1"}, asyncInsert)
	registerGo("probes/client/json_insert.go", "Batch insert of JSON text into a JSON column (CL-04)",
		map[string]string{"": "value:3"}, jsonInsert)
	registerGo("probes/security/tls.go", "Transport security of the configured connection (CO-06)",
		map[string]string{"": "ok", "cloud": "value:tls-only"}, tlsProbe)
}

func connectWith(ctx context.Context, env Env, mutate func(o *clickhouse.Options)) (clickhouse.Conn, error) {
	o := *env.Options
	o.Auth.Database = env.DB
	mutate(&o)
	c, err := clickhouse.Open(&o)
	if err != nil {
		return nil, err
	}
	return c, c.Ping(ctx)
}

func protocols(ctx context.Context, env Env) Observation {
	var got []string
	if err := env.Conn.Ping(ctx); err == nil {
		got = append(got, "native")
	}
	host, _, _ := net.SplitHostPort(env.Options.Addr[0])
	url := ""
	switch {
	case env.Manifest.Ports != nil:
		url = fmt.Sprintf("http://%s:%d/ping", host, env.Manifest.Ports.HTTPBase)
	case env.Options.TLS != nil:
		url = fmt.Sprintf("https://%s:8443/ping", host)
	}
	if url != "" {
		c := &http.Client{Timeout: 5 * time.Second}
		if resp, err := c.Get(url); err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				got = append(got, "http")
			}
		}
	}
	return Observation{Kind: "value", Value: strings.Join(got, ",")}
}

func compression(ctx context.Context, env Env) Observation {
	var got []string
	for _, m := range []struct {
		name   string
		method clickhouse.CompressionMethod
	}{{"lz4", clickhouse.CompressionLZ4}, {"zstd", clickhouse.CompressionZSTD}} {
		c, err := connectWith(ctx, env, func(o *clickhouse.Options) { o.Compression = &clickhouse.Compression{Method: m.method} })
		if err != nil {
			continue
		}
		var n uint64
		if err := c.QueryRow(ctx, "SELECT count() FROM numbers(100000)").Scan(&n); err == nil && n == 100000 {
			got = append(got, m.name)
		}
		c.Close()
	}
	return Observation{Kind: "value", Value: strings.Join(got, ",")}
}

func asyncInsert(ctx context.Context, env Env) Observation {
	if err := env.Conn.Exec(ctx, fmt.Sprintf("CREATE TABLE %s.async_go_probe (k UInt8) ENGINE = MergeTree ORDER BY k", env.DB)); err != nil {
		return errObs(err, 0)
	}
	if err := env.Conn.AsyncInsert(ctx, fmt.Sprintf("INSERT INTO %s.async_go_probe VALUES (1)", env.DB), true); err != nil {
		return errObs(err, 1)
	}
	var n uint64
	if err := env.Conn.QueryRow(ctx, fmt.Sprintf("SELECT count() FROM %s.async_go_probe", env.DB)).Scan(&n); err != nil {
		return errObs(err, 2)
	}
	return Observation{Kind: "value", Value: fmt.Sprint(n)}
}

func jsonInsert(ctx context.Context, env Env) Observation {
	if err := env.Conn.Exec(ctx, fmt.Sprintf("CREATE TABLE %s.json_go_probe (d JSON) ENGINE = MergeTree ORDER BY tuple()", env.DB)); err != nil {
		return errObs(err, 0)
	}
	b, err := env.Conn.PrepareBatch(ctx, fmt.Sprintf("INSERT INTO %s.json_go_probe (d)", env.DB))
	if err != nil {
		return errObs(err, 1)
	}
	for _, s := range []string{`{"v": 1}`, `{"v": "2"}`, `{"v": 3, "x": {"y": true}}`} {
		if err := b.Append(s); err != nil {
			return errObs(err, 1)
		}
	}
	if err := b.Send(); err != nil {
		return errObs(err, 1)
	}
	var n uint64
	if err := env.Conn.QueryRow(ctx, fmt.Sprintf("SELECT count() FROM %s.json_go_probe WHERE d.v IS NOT NULL", env.DB)).Scan(&n); err != nil {
		return errObs(err, 2)
	}
	return Observation{Kind: "value", Value: fmt.Sprint(n)}
}

// tlsProbe reports the connection's transport and, for a TLS connection, whether
// the same host also accepts plaintext native connections on 9000.
func tlsProbe(ctx context.Context, env Env) Observation {
	if env.Options.TLS == nil {
		return Observation{Kind: "value", Value: "plaintext"}
	}
	host, _, _ := net.SplitHostPort(env.Options.Addr[0])
	o := *env.Options
	o.TLS = nil
	o.Addr = []string{net.JoinHostPort(host, "9000")}
	o.DialTimeout = 3 * time.Second
	if c, err := clickhouse.Open(&o); err == nil {
		defer c.Close()
		if c.Ping(ctx) == nil {
			return Observation{Kind: "value", Value: "tls+plaintext"}
		}
	}
	return Observation{Kind: "value", Value: "tls-only"}
}

// Package otel generates deterministic synthetic OpenTelemetry traces, logs and
// metrics for the otel-traces, otel-logs and otel-metrics workloads (design D6,
// ADR-0004).
//
// Every record is a pure function of (seed, index): any target, run or ingest path
// receives byte-identical data, so value digests are comparable across targets.
// Two sinks consume the generated records:
//
//   - a direct native insert that reproduces the collector-contrib ClickHouse
//     exporter's column mapping (v0.162.0: exporter_traces.go, exporter_logs.go,
//     internal/metrics/*), used for scale; and
//   - OTLP/HTTP JSON sent to an OpenTelemetry Collector, the real ingest path.
//
// Loading the same records both ways must give identical query digests (the
// path-parity check), which proves the direct mapper matches the exporter.
//
// All names, hosts and identifiers are synthetic (example.com, documentation IP
// ranges); nothing is derived from real systems.
package otel

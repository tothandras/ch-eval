package otel

import "strconv"

// Attr is one attribute. OTLP carries ints as intValue; the exporter stores every
// value with pcommon.Value.AsString, which renders ints in decimal. The generator
// uses only string and int values so both paths agree byte for byte.
type Attr struct {
	Key   string
	Str   string
	Int   int64
	IsInt bool
}

// S builds a string attribute.
func S(k, v string) Attr { return Attr{Key: k, Str: v} }

// I builds an int attribute.
func I(k string, v int64) Attr { return Attr{Key: k, Int: v, IsInt: true} }

// String is the exporter's AsString rendering.
func (a Attr) String() string {
	if a.IsInt {
		return strconv.FormatInt(a.Int, 10)
	}
	return a.Str
}

// Resource is one emitting entity (a service instance).
type Resource struct {
	Attrs     []Attr
	SchemaURL string
}

// ServiceName is the resource's service.name, as the exporter extracts it.
func (r Resource) ServiceName() string {
	for _, a := range r.Attrs {
		if a.Key == "service.name" {
			return a.String()
		}
	}
	return ""
}

// Scope is an instrumentation scope.
type Scope struct {
	Name, Version string
	Attrs         []Attr
	SchemaURL     string
}

// SpanKind values are the OTLP enum.
type SpanKind int32

const (
	KindUnspecified SpanKind = 0
	KindInternal    SpanKind = 1
	KindServer      SpanKind = 2
	KindClient      SpanKind = 3
	KindProducer    SpanKind = 4
	KindConsumer    SpanKind = 5
)

// String matches pdata's ptrace.SpanKind.String(), which the exporter stores.
func (k SpanKind) String() string {
	switch k {
	case KindInternal:
		return "Internal"
	case KindServer:
		return "Server"
	case KindClient:
		return "Client"
	case KindProducer:
		return "Producer"
	case KindConsumer:
		return "Consumer"
	}
	return "Unspecified"
}

// StatusCode values are the OTLP enum.
type StatusCode int32

const (
	StatusUnset StatusCode = 0
	StatusOk    StatusCode = 1
	StatusError StatusCode = 2
)

// String matches pdata's ptrace.StatusCode.String().
func (c StatusCode) String() string {
	switch c {
	case StatusOk:
		return "Ok"
	case StatusError:
		return "Error"
	}
	return "Unset"
}

// Event is a span event.
type Event struct {
	TimeNs uint64
	Name   string
	Attrs  []Attr
}

// Link is a span link.
type Link struct {
	TraceID    [16]byte
	SpanID     [8]byte
	TraceState string
	Attrs      []Attr
}

// Span is one generated span.
type Span struct {
	TraceID       [16]byte
	SpanID        [8]byte
	ParentSpanID  [8]byte // zero for a root span
	TraceState    string
	Name          string
	Kind          SpanKind
	StartNs       uint64
	EndNs         uint64
	Attrs         []Attr
	Events        []Event
	Links         []Link
	Status        StatusCode
	StatusMessage string
	Resource      int // index into the generator's resources
	Scope         int // index into the generator's scopes
}

// LogRecord is one generated log record.
type LogRecord struct {
	TimeNs         uint64
	ObservedTimeNs uint64
	TraceID        [16]byte // zero when uncorrelated
	SpanID         [8]byte
	Flags          uint32
	SeverityText   string
	SeverityNumber int32
	Body           string
	Attrs          []Attr
	EventName      string
	Resource       int
	Scope          int
}

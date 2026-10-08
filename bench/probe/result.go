package probe

import "fmt"

// Observation is what a probe's final statement produced.
type Observation struct {
	Kind      string `json:"kind"` // ok | value | error
	Value     string `json:"value,omitempty"`
	ErrorCode int32  `json:"error_code,omitempty"`
	Message   string `json:"message,omitempty"`
}

func (o Observation) String() string {
	switch o.Kind {
	case "error":
		return fmt.Sprintf("error:%d", o.ErrorCode)
	case "value":
		return "value:" + o.Value
	}
	return o.Kind
}

// Result is one probe run on one target.
type Result struct {
	Path       string      `json:"path"`
	Rows       []string    `json:"rows"` // feature-matrix row ids
	Title      string      `json:"title"`
	Expect     string      `json:"expect"`
	Observed   Observation `json:"observed"`
	Pass       bool        `json:"pass"`
	DurationMs float64     `json:"duration_ms"`
	Notes      string      `json:"notes,omitempty"`
}

package lookup

import (
	"context"
	"encoding/json"
	"io"
	"sync"
	"time"
)

// Trace records what a lookup did, one JSON object per line, for the
// output.txt debug report: every query, page, extraction, decision and
// error. Keys are short on purpose; the report is read by a program, not
// by a person.
type Trace struct {
	mu      sync.Mutex
	w       io.Writer
	start   time.Time
	company string
}

// NewTrace writes trace lines to w.
func NewTrace(w io.Writer) *Trace { return &Trace{w: w, start: time.Now()} }

// Company starts the lines of one company.
func (t *Trace) Company(name string) {
	if t == nil {
		return
	}
	t.mu.Lock()
	t.company = name
	t.mu.Unlock()
	t.Ev("company", "raw", name)
}

// Ev writes one event: kind followed by key, value pairs.
func (t *Trace) Ev(kind string, kv ...any) {
	if t == nil || t.w == nil {
		return
	}
	m := map[string]any{"t": time.Since(t.start).Milliseconds(), "e": kind}
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.company != "" {
		m["c"] = t.company
	}
	for i := 0; i+1 < len(kv); i += 2 {
		k, _ := kv[i].(string)
		switch v := kv[i+1].(type) {
		case error:
			if v != nil {
				m[k] = sanitizeErr(v)
			}
		default:
			m[k] = v
		}
	}
	b, err := json.Marshal(m)
	if err == nil {
		_, _ = t.w.Write(append(b, '\n'))
	}
}

type traceKey struct{}

// WithTrace attaches a trace to a context.
func WithTrace(ctx context.Context, t *Trace) context.Context {
	return context.WithValue(ctx, traceKey{}, t)
}

func tr(ctx context.Context) *Trace {
	t, _ := ctx.Value(traceKey{}).(*Trace)
	return t
}

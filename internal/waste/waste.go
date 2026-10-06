// Package waste classifies wasted tokens in Vulcan run traces.
//
// v0 labels are deterministic and come straight from the trace:
//
//	repeat     — same tool + args_hash called again within the run
//	failed     — tool call returned an error
//	oversized  — result larger than the size cap
//
// The "unused result" class needs edit correlation and lands in M2+.
package waste

import (
	"fmt"
	"sort"
	"time"

	"vulcan/internal/vigil"
)

const (
	DefaultSizeCap = 20_000 // bytes
)

// Class is one waste category.
type Class string

const (
	ClassRepeat    Class = "repeat"
	ClassFailed    Class = "failed"
	ClassOversized Class = "oversized"
)

// Step is one tool call extracted from a run.step event.
type Step struct {
	EventID     string
	TS          time.Time
	Tool        string
	ArgsHash    string
	ResultBytes int
	IsError     bool
	DurationMS  int64
	Model       string
}

// Run is one agent run reconstructed from events sharing a trace_id.
type Run struct {
	ID        string
	StartedAt time.Time
	Steps     []Step
	// From run.completed / run.failed / run.aborted attrs, when present.
	TotalTokens int
	Status      string
}

// Finding is a single classified waste occurrence.
type Finding struct {
	Class      Class
	Step       Step
	Bytes      int // addressable bytes attributed to this finding
	Detail     string
}

// Build runs reconstruct runs from raw Vigil events.
func Build(events []vigil.Event) []Run {
	byRun := map[string]*Run{}
	var order []string

	for _, ev := range events {
		id := ev.TraceID
		if id == "" {
			continue
		}
		run := byRun[id]
		if run == nil {
			run = &Run{ID: id}
			byRun[id] = run
			order = append(order, id)
		}
		t := ev.Time()
		if run.StartedAt.IsZero() || t.Before(run.StartedAt) {
			run.StartedAt = t
		}

		attrs := ev.AttrsMap()
		switch {
		case ev.Name == "run.step":
			step := Step{
				EventID:     ev.EventID,
				TS:          t,
				Tool:        str(attrs["tool"]),
				ArgsHash:    str(attrs["args_hash"]),
				ResultBytes: num(attrs["result_bytes"]),
				IsError:     boolOf(attrs["is_error"]),
				DurationMS:  int64(num(attrs["duration_ms"])),
				Model:       str(attrs["model"]),
			}
			run.Steps = append(run.Steps, step)

		case ev.Name == "run.completed":
			run.Status = "completed"
			run.TotalTokens = num(attrs["total_tokens"])
			if len(run.Steps) == 0 || t.After(lastStepTime(run)) {
				// terminal event; keep as status marker
			}

		case ev.Name == "run.failed", ev.Name == "run.aborted":
			run.Status = strAfter(ev.Name, "run.")
			run.TotalTokens = num(attrs["total_tokens"])
		}
	}

	runs := make([]Run, 0, len(order))
	for _, id := range order {
		r := byRun[id]
		if r.Status == "" {
			r.Status = "running"
		}
		sort.Slice(r.Steps, func(i, j int) bool { return r.Steps[i].TS.Before(r.Steps[j].TS) })
		runs = append(runs, *r)
	}
	sort.Slice(runs, func(i, j int) bool { return runs[i].StartedAt.Before(runs[j].StartedAt) })
	return runs
}

// Analyze classifies waste in a run. Addressable bytes are an estimate of the
// token traffic the policy layer could have avoided.
func Analyze(run Run, sizeCap int) []Finding {
	var findings []Finding

	seen := map[string]Step{} // tool+args_hash -> first occurrence
	for _, step := range run.Steps {
		key := step.Tool + "|" + step.ArgsHash

		if step.IsError {
			findings = append(findings, Finding{
				Class: ClassFailed, Step: step,
				Bytes:  step.ResultBytes, // error text itself
				Detail: fmt.Sprintf("%s failed after %dms", step.Tool, step.DurationMS),
			})
		}

		if first, ok := seen[key]; ok {
			_ = first
			findings = append(findings, Finding{
				Class: ClassRepeat, Step: step,
				Bytes:  step.ResultBytes,
				Detail: fmt.Sprintf("%s repeated (%s)", step.Tool, short(step.ArgsHash)),
			})
		} else {
			seen[key] = step
		}

		if step.ResultBytes > sizeCap {
			findings = append(findings, Finding{
				Class: ClassOversized, Step: step,
				Bytes:  step.ResultBytes - sizeCap,
				Detail: fmt.Sprintf("%s returned %d bytes (cap %d)", step.Tool, step.ResultBytes, sizeCap),
			})
		}
	}

	sort.SliceStable(findings, func(i, j int) bool {
		return findings[i].Step.TS.Before(findings[j].Step.TS)
	})
	return findings
}

// Summary aggregates findings for a run.
type Summary struct {
	Runs      int
	Steps     int
	Findings  map[Class]int
	Bytes     map[Class]int
	TotalCap  int // sum of all result bytes seen
	Addressable int
}

func Summarize(runs []Run, findings map[string][]Finding) Summary {
	s := Summary{
		Findings: map[Class]int{},
		Bytes:    map[Class]int{},
	}
	for _, run := range runs {
		s.Runs++
		s.Steps += len(run.Steps)
		for _, f := range findings[run.ID] {
			s.Findings[f.Class]++
			s.Bytes[f.Class] += f.Bytes
			s.Addressable += f.Bytes
		}
		for _, step := range run.Steps {
			s.TotalCap += step.ResultBytes
		}
	}
	return s
}

// helpers

func str(v any) string {
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

func num(v any) int {
	switch n := v.(type) {
	case float64:
		return int(n)
	case int:
		return n
	case int64:
		return int(n)
	}
	return 0
}

func boolOf(v any) bool {
	b, ok := v.(bool)
	return ok && b
}

func strAfter(s, prefix string) string {
	if len(s) > len(prefix) && s[:len(prefix)] == prefix {
		return s[len(prefix):]
	}
	return s
}

func short(s string) string {
	if len(s) > 8 {
		return s[:8]
	}
	return s
}

func lastStepTime(run *Run) time.Time {
	if len(run.Steps) == 0 {
		return time.Time{}
	}
	return run.Steps[len(run.Steps)-1].TS
}

package ui

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/idestis/pipe/internal/model"
)

func steps(ids ...string) []model.Step {
	var out []model.Step
	for _, id := range ids {
		out = append(out, model.Step{
			ID:  id,
			Run: model.RunField{Single: "echo ok"},
		})
	}
	return out
}

func TestNewStatusUI_RowCount(t *testing.T) {
	s := NewStatusUI(&bytes.Buffer{}, steps("a", "b", "c"))
	if len(s.rows) != 3 {
		t.Fatalf("expected 3 rows, got %d", len(s.rows))
	}
}

func TestNewStatusUI_ParallelStrings(t *testing.T) {
	st := []model.Step{{
		ID:  "lint",
		Run: model.RunField{Strings: []string{"cmd1", "cmd2", "cmd3"}},
	}}
	s := NewStatusUI(&bytes.Buffer{}, st)
	if len(s.rows) != 3 {
		t.Fatalf("expected 3 rows for parallel strings, got %d", len(s.rows))
	}
	if s.rows[0].id != "lint/run_0" {
		t.Fatalf("expected lint/run_0, got %s", s.rows[0].id)
	}
}

func TestNewStatusUI_SubRuns(t *testing.T) {
	st := []model.Step{{
		ID: "deploy",
		Run: model.RunField{SubRuns: []model.SubRun{
			{ID: "api", Run: "deploy api"},
			{ID: "web", Run: "deploy web"},
		}},
	}}
	s := NewStatusUI(&bytes.Buffer{}, st)
	if len(s.rows) != 2 {
		t.Fatalf("expected 2 rows for sub-runs, got %d", len(s.rows))
	}
	if s.rows[1].id != "deploy/web" {
		t.Fatalf("expected deploy/web, got %s", s.rows[1].id)
	}
}

func TestSetStatus_Transitions(t *testing.T) {
	var buf bytes.Buffer
	s := NewStatusUI(&buf, steps("a"))

	s.SetStatus("a", Running)
	if s.rows[0].status != Running {
		t.Fatal("expected Running")
	}
	if s.rows[0].startedAt.IsZero() {
		t.Fatal("expected startedAt to be set")
	}

	s.SetStatus("a", Done)
	if s.rows[0].status != Done {
		t.Fatal("expected Done")
	}
	if s.rows[0].duration == 0 {
		t.Fatal("expected duration > 0")
	}
}

func TestSetStatus_UnknownID(t *testing.T) {
	var buf bytes.Buffer
	s := NewStatusUI(&buf, steps("a"))
	// Should not panic
	s.SetStatus("nonexistent", Running)
}

func TestRender_Icons(t *testing.T) {
	var buf bytes.Buffer
	s := NewStatusUI(&buf, steps("build"))

	s.SetStatus("build", Running)
	out := buf.String()
	if !strings.Contains(out, "●") {
		t.Fatalf("expected ● in output, got: %s", out)
	}
	if !strings.Contains(out, "running...") {
		t.Fatalf("expected 'running...' in output, got: %s", out)
	}

	buf.Reset()
	s.SetStatus("build", Done)
	out = buf.String()
	if !strings.Contains(out, "✓") {
		t.Fatalf("expected ✓ in output, got: %s", out)
	}
}

func TestRender_FailedIcon(t *testing.T) {
	var buf bytes.Buffer
	s := NewStatusUI(&buf, steps("test"))
	s.SetStatus("test", Running)
	buf.Reset()
	s.SetStatus("test", Failed)
	out := buf.String()
	if !strings.Contains(out, "✗") {
		t.Fatalf("expected ✗ in output, got: %s", out)
	}
}

func TestRender_WaitingIcon(t *testing.T) {
	var buf bytes.Buffer
	s := NewStatusUI(&buf, steps("push"))
	s.Finish()
	out := buf.String()
	if !strings.Contains(out, "○") {
		t.Fatalf("expected ○ in output, got: %s", out)
	}
	if !strings.Contains(out, "waiting") {
		t.Fatalf("expected 'waiting' in output, got: %s", out)
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		d    time.Duration
		want string
	}{
		{500 * time.Millisecond, "(0.5s)"},
		{2100 * time.Millisecond, "(2.1s)"},
		{59900 * time.Millisecond, "(59.9s)"},
		{90 * time.Second, "(1m 30s)"},
		{125 * time.Second, "(2m 5s)"},
	}
	for _, tt := range tests {
		got := formatDuration(tt.d)
		if got != tt.want {
			t.Errorf("formatDuration(%v) = %q, want %q", tt.d, got, tt.want)
		}
	}
}

func TestMaxWidth_Alignment(t *testing.T) {
	var buf bytes.Buffer
	s := NewStatusUI(&buf, steps("a", "longname"))
	if s.maxWidth != len("longname") {
		t.Fatalf("expected maxWidth=%d, got %d", len("longname"), s.maxWidth)
	}
}

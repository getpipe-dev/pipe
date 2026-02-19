package runner

import (
	"strings"
	"sync"
	"testing"
)

func TestOutputWriter_SingleLine(t *testing.T) {
	var lines []string
	w := newOutputWriter(func(s string) { lines = append(lines, s) })

	_, _ = w.Write([]byte("hello world\n"))

	if len(lines) != 1 {
		t.Fatalf("expected 1 line, got %d", len(lines))
	}
	if lines[0] != "hello world" {
		t.Fatalf("expected %q, got %q", "hello world", lines[0])
	}
}

func TestOutputWriter_MultipleLines(t *testing.T) {
	var lines []string
	w := newOutputWriter(func(s string) { lines = append(lines, s) })

	_, _ = w.Write([]byte("line1\nline2\nline3\n"))

	if len(lines) != 3 {
		t.Fatalf("expected 3 lines, got %d", len(lines))
	}
	if lines[0] != "line1" {
		t.Fatalf("line 0: expected %q, got %q", "line1", lines[0])
	}
	if lines[2] != "line3" {
		t.Fatalf("line 2: expected %q, got %q", "line3", lines[2])
	}
}

func TestOutputWriter_PartialLine(t *testing.T) {
	var lines []string
	w := newOutputWriter(func(s string) { lines = append(lines, s) })

	_, _ = w.Write([]byte("hel"))
	if len(lines) != 0 {
		t.Fatalf("expected 0 lines after partial write, got %d", len(lines))
	}

	_, _ = w.Write([]byte("lo\n"))
	if len(lines) != 1 {
		t.Fatalf("expected 1 line after completing line, got %d", len(lines))
	}
	if lines[0] != "hello" {
		t.Fatalf("expected %q, got %q", "hello", lines[0])
	}
}

func TestOutputWriter_Flush(t *testing.T) {
	var lines []string
	w := newOutputWriter(func(s string) { lines = append(lines, s) })

	_, _ = w.Write([]byte("partial"))
	if len(lines) != 0 {
		t.Fatalf("expected 0 lines before flush, got %d", len(lines))
	}

	w.Flush()
	if len(lines) != 1 {
		t.Fatalf("expected 1 line after flush, got %d", len(lines))
	}
	if lines[0] != "partial" {
		t.Fatalf("expected %q, got %q", "partial", lines[0])
	}
}

func TestOutputWriter_FlushEmpty(t *testing.T) {
	var lines []string
	w := newOutputWriter(func(s string) { lines = append(lines, s) })

	w.Flush()
	if len(lines) != 0 {
		t.Fatalf("expected 0 lines on empty flush, got %d", len(lines))
	}
}

func TestOutputWriter_EmptyLines(t *testing.T) {
	var lines []string
	w := newOutputWriter(func(s string) { lines = append(lines, s) })

	_, _ = w.Write([]byte("a\n\nb\n"))
	if len(lines) != 3 {
		t.Fatalf("expected 3 lines (including empty), got %d", len(lines))
	}
	if lines[1] != "" {
		t.Fatalf("expected empty line %q, got %q", "", lines[1])
	}
}

func TestOutputWriter_ConcurrentWrites(t *testing.T) {
	var mu sync.Mutex
	var lines []string
	w := newOutputWriter(func(s string) {
		mu.Lock()
		lines = append(lines, s)
		mu.Unlock()
	})

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _ = w.Write([]byte("line\n"))
		}()
	}
	wg.Wait()

	mu.Lock()
	defer mu.Unlock()
	if len(lines) != 10 {
		t.Fatalf("expected 10 lines, got %d", len(lines))
	}
	for _, l := range lines {
		if !strings.HasPrefix(l, "line") {
			t.Fatalf("expected 'line' content, got %q", l)
		}
	}
}

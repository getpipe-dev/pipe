package parser

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/destis/pipe/internal/config"
)

// overrideFilesDir points config.FilesDir at a temp directory for the test
// and restores the original value when the test finishes.
func overrideFilesDir(t *testing.T) string {
	t.Helper()
	orig := config.FilesDir
	tmp := t.TempDir()
	config.FilesDir = tmp
	t.Cleanup(func() { config.FilesDir = orig })
	return tmp
}

func writeYAML(t *testing.T, dir, name, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name+".yaml"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestLoadPipeline_Valid(t *testing.T) {
	dir := overrideFilesDir(t)
	writeYAML(t, dir, "deploy", `
name: deploy
steps:
  - id: build
    run: "echo build"
`)
	p, err := LoadPipeline("deploy")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Name != "deploy" {
		t.Fatalf("expected name %q, got %q", "deploy", p.Name)
	}
	if len(p.Steps) != 1 {
		t.Fatalf("expected 1 step, got %d", len(p.Steps))
	}
}

func TestLoadPipeline_NameDefaultsToFilename(t *testing.T) {
	dir := overrideFilesDir(t)
	writeYAML(t, dir, "myfile", `
steps:
  - id: hello
    run: "echo hi"
`)
	p, err := LoadPipeline("myfile")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Name != "myfile" {
		t.Fatalf("expected name %q, got %q", "myfile", p.Name)
	}
}

func TestLoadPipeline_FileNotFound(t *testing.T) {
	overrideFilesDir(t)
	_, err := LoadPipeline("nonexistent")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
	if !strings.Contains(err.Error(), "reading pipeline") {
		t.Fatalf("expected error containing %q, got %q", "reading pipeline", err.Error())
	}
}

func TestLoadPipeline_InvalidYAML(t *testing.T) {
	dir := overrideFilesDir(t)
	writeYAML(t, dir, "bad", `{{{invalid`)
	_, err := LoadPipeline("bad")
	if err == nil {
		t.Fatal("expected error for invalid YAML")
	}
	if !strings.Contains(err.Error(), "parsing pipeline") {
		t.Fatalf("expected error containing %q, got %q", "parsing pipeline", err.Error())
	}
}

func TestValidate_MissingID(t *testing.T) {
	dir := overrideFilesDir(t)
	writeYAML(t, dir, "noid", `
steps:
  - run: "echo hi"
`)
	_, err := LoadPipeline("noid")
	if err == nil {
		t.Fatal("expected error for missing id")
	}
	if !strings.Contains(err.Error(), "missing id") {
		t.Fatalf("expected error containing %q, got %q", "missing id", err.Error())
	}
}

func TestValidate_DuplicateID(t *testing.T) {
	dir := overrideFilesDir(t)
	writeYAML(t, dir, "dupid", `
steps:
  - id: same
    run: "echo a"
  - id: same
    run: "echo b"
`)
	_, err := LoadPipeline("dupid")
	if err == nil {
		t.Fatal("expected error for duplicate id")
	}
	if !strings.Contains(err.Error(), "duplicate id") {
		t.Fatalf("expected error containing %q, got %q", "duplicate id", err.Error())
	}
}

func TestValidate_MissingRun(t *testing.T) {
	dir := overrideFilesDir(t)
	writeYAML(t, dir, "norun", `
steps:
  - id: empty
`)
	_, err := LoadPipeline("norun")
	if err == nil {
		t.Fatal("expected error for missing run field")
	}
	if !strings.Contains(err.Error(), "missing run field") {
		t.Fatalf("expected error containing %q, got %q", "missing run field", err.Error())
	}
}

func TestValidate_Valid(t *testing.T) {
	dir := overrideFilesDir(t)
	writeYAML(t, dir, "ok", `
steps:
  - id: a
    run: "echo a"
  - id: b
    run: ["x", "y"]
`)
	_, err := LoadPipeline("ok")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

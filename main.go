package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/idestis/pipe/internal/config"
	"github.com/idestis/pipe/internal/logging"
	"github.com/idestis/pipe/internal/parser"
	"github.com/idestis/pipe/internal/runner"
	"github.com/idestis/pipe/internal/state"
)

// version is set by goreleaser via ldflags
var version = "dev"

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: pipe <command|pipeline> [args]")
		fmt.Fprintln(os.Stderr, "commands: init, list, validate")
		os.Exit(1)
	}

	if os.Args[1] == "-h" || os.Args[1] == "--help" {
		fmt.Println(`Usage: pipe <command|pipeline> [flags]

Commands:
  init <name>       Create a new pipeline
  list              List all pipelines
  validate <name>   Validate a pipeline

Run a pipeline:
  pipe <name> [--resume <run-id>]

Flags:
  -h, --help       Show help
  -v, --version    Show version`)
		return
	}

	if os.Args[1] == "--version" || os.Args[1] == "-v" {
		fmt.Printf("pipe-%s\n", version)
		return
	}

	switch os.Args[1] {
	case "init":
		cmdInit()
	case "list":
		cmdList()
	case "validate":
		cmdValidate()
	default:
		runPipeline()
	}
}

func cmdInit() {
	if hasFlag(os.Args[2:], "-h", "--help") {
		fmt.Println("Usage: pipe init <name>")
		return
	}
	if len(os.Args) < 3 {
		fmt.Fprintln(os.Stderr, "usage: pipe init <name>")
		os.Exit(1)
	}
	name := os.Args[2]

	switch name {
	case "init", "list", "validate":
		fmt.Fprintf(os.Stderr, "error: %q is a reserved command name\n", name)
		os.Exit(1)
	}

	if !validName(name) {
		fmt.Fprintf(os.Stderr, "error: invalid pipeline name %q — use only letters, digits, hyphens, and underscores\n", name)
		os.Exit(1)
	}

	if err := os.MkdirAll(config.FilesDir, 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", friendlyError(err))
		os.Exit(1)
	}

	path := filepath.Join(config.FilesDir, name+".yaml")
	if _, err := os.Stat(path); err == nil {
		fmt.Fprintf(os.Stderr, "error: pipeline %q already exists at %s\n", name, path)
		os.Exit(1)
	}

	template := fmt.Sprintf(`name: %s
description: ""
steps:
  - id: hello
    run: "echo Hello from %s"
`, name, name)

	if err := os.WriteFile(path, []byte(template), 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", friendlyError(err))
		os.Exit(1)
	}
	fmt.Println(path)
}

func cmdList() {
	if hasFlag(os.Args[2:], "-h", "--help") {
		fmt.Println("Usage: pipe list")
		return
	}
	infos, err := parser.ListPipelines()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	if len(infos) == 0 {
		fmt.Println("no pipelines found — use 'pipe init <name>' to create one")
		return
	}

	// find max name width for alignment
	maxName := len("NAME")
	for _, info := range infos {
		if len(info.Name) > maxName {
			maxName = len(info.Name)
		}
	}

	fmt.Printf("%-*s  %s\n", maxName, "NAME", "DESCRIPTION")
	for _, info := range infos {
		fmt.Printf("%-*s  %s\n", maxName, info.Name, info.Description)
	}
}

func cmdValidate() {
	if hasFlag(os.Args[2:], "-h", "--help") {
		fmt.Println("Usage: pipe validate <name>")
		return
	}
	if len(os.Args) < 3 {
		fmt.Fprintln(os.Stderr, "usage: pipe validate <name>")
		os.Exit(1)
	}
	name := os.Args[2]

	if err := parser.ValidatePipeline(name); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			fmt.Fprintf(os.Stderr, "error: pipeline %q not found\n", name)
			fmt.Fprintf(os.Stderr, "  run \"pipe list\" to see available pipelines, or \"pipe init %s\" to create one\n", name)
		} else if isYAMLError(err) {
			fmt.Fprintf(os.Stderr, "error: invalid YAML in pipeline %q: %v\n", name, unwrapYAMLError(err))
		} else {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
		}
		os.Exit(1)
	}
	fmt.Printf("pipeline %q is valid\n", name)
}

func validName(name string) bool {
	if len(name) == 0 {
		return false
	}
	for i, c := range name {
		switch {
		case c >= 'a' && c <= 'z', c >= 'A' && c <= 'Z', c >= '0' && c <= '9':
			// always allowed
		case c == '-' || c == '_':
			if i == 0 {
				return false
			}
		default:
			return false
		}
	}
	return true
}

func hasFlag(args []string, short, long string) bool {
	for _, a := range args {
		if a == short || a == long {
			return true
		}
	}
	return false
}

func runPipeline() {
	if hasFlag(os.Args[2:], "-h", "--help") {
		fmt.Println("Usage: pipe <pipeline> [--resume <run-id>]")
		return
	}
	pipelineName := os.Args[1]
	var resumeID string

	for i := 2; i < len(os.Args); i++ {
		if os.Args[i] == "--resume" && i+1 < len(os.Args) {
			resumeID = os.Args[i+1]
			i++
		}
	}

	pipeline, err := parser.LoadPipeline(pipelineName)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			fmt.Fprintf(os.Stderr, "error: pipeline %q not found\n", pipelineName)
			fmt.Fprintf(os.Stderr, "  run \"pipe list\" to see available pipelines, or \"pipe init %s\" to create one\n", pipelineName)
		} else if errors.Is(err, os.ErrPermission) {
			fmt.Fprintf(os.Stderr, "error: permission denied reading pipeline %q — check file permissions\n", pipelineName)
		} else if isYAMLError(err) {
			fmt.Fprintf(os.Stderr, "error: invalid YAML in pipeline %q: %v\n", pipelineName, unwrapYAMLError(err))
		} else {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
		}
		os.Exit(1)
	}

	if err := config.EnsureDirs(pipeline.Name); err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", friendlyError(err))
		os.Exit(1)
	}

	var rs *state.RunState
	if resumeID != "" {
		rs, err = state.Load(pipeline.Name, resumeID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		rs.Status = "running"
	} else {
		rs = state.NewRunState(pipeline.Name)
	}

	log, err := logging.New(pipeline.Name, rs.RunID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", friendlyError(err))
		os.Exit(1)
	}
	defer func() { _ = log.Close() }()

	if resumeID != "" {
		log.Log("resuming pipeline %q (run %s)", pipeline.Name, rs.RunID)
	} else {
		log.Log("starting pipeline %q (run %s)", pipeline.Name, rs.RunID)
	}

	if err := state.Save(rs); err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", friendlyError(err))
		os.Exit(1)
	}

	r := runner.New(pipeline, rs, log)
	if resumeID != "" {
		r.RestoreEnvFromState()
	}

	if err := r.Run(); err != nil {
		os.Exit(1)
	}
}

// friendlyError converts common OS errors into user-friendly messages.
func friendlyError(err error) string {
	if errors.Is(err, os.ErrPermission) {
		return "permission denied — check directory permissions for ~/.pipe"
	}
	return err.Error()
}

// isYAMLError returns true if the error originated from YAML parsing.
func isYAMLError(err error) bool {
	return strings.Contains(err.Error(), "parsing pipeline")
}

// unwrapYAMLError extracts the YAML-specific error detail from a wrapped
// "parsing pipeline" error, stripping the redundant prefix.
func unwrapYAMLError(err error) error {
	msg := err.Error()
	// Strip our wrapping prefix "parsing pipeline \"name\": " to get the yaml detail.
	if i := strings.Index(msg, "parsing pipeline"); i >= 0 {
		// Find the ": " after the closing quote of the pipeline name.
		rest := msg[i:]
		if j := strings.Index(rest, ": "); j >= 0 {
			detail := rest[j+2:]
			return errors.New(detail)
		}
	}
	return err
}

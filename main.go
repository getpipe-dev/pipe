package main

import (
	"fmt"
	"os"

	"github.com/destis/pipe/internal/config"
	"github.com/destis/pipe/internal/logging"
	"github.com/destis/pipe/internal/parser"
	"github.com/destis/pipe/internal/runner"
	"github.com/destis/pipe/internal/state"
)

// version is set by goreleaser via ldflags
var version = "dev"

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: pipe <pipeline> [--resume <run-id>]")
		os.Exit(1)
	}

	if os.Args[1] == "--version" || os.Args[1] == "-v" {
		fmt.Printf("pipe-%s\n", version)
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
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	if err := config.EnsureDirs(pipeline.Name); err != nil {
		fmt.Fprintf(os.Stderr, "error creating dirs: %v\n", err)
		os.Exit(1)
	}

	var rs *state.RunState
	if resumeID != "" {
		rs, err = state.Load(pipeline.Name, resumeID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error loading state: %v\n", err)
			os.Exit(1)
		}
		rs.Status = "running"
	} else {
		rs = state.NewRunState(pipeline.Name)
	}

	log, err := logging.New(pipeline.Name, rs.RunID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error creating logger: %v\n", err)
		os.Exit(1)
	}
	defer func() { _ = log.Close() }()

	if resumeID != "" {
		log.Info("resuming pipeline %q (run %s)", pipeline.Name, rs.RunID)
	} else {
		log.Info("starting pipeline %q (run %s)", pipeline.Name, rs.RunID)
	}

	if err := state.Save(rs); err != nil {
		fmt.Fprintf(os.Stderr, "error saving initial state: %v\n", err)
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

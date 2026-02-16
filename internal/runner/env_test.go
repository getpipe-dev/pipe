package runner

import (
	"os"
	"strings"
	"testing"
)

func TestEnvKey_Single(t *testing.T) {
	t.Parallel()
	if got := EnvKey("build"); got != "PIPE_BUILD" {
		t.Fatalf("expected PIPE_BUILD, got %s", got)
	}
}

func TestEnvKey_Multi(t *testing.T) {
	t.Parallel()
	if got := EnvKey("deploy", "east"); got != "PIPE_DEPLOY_EAST" {
		t.Fatalf("expected PIPE_DEPLOY_EAST, got %s", got)
	}
}

func TestEnvKey_Hyphens(t *testing.T) {
	t.Parallel()
	if got := EnvKey("my-step", "sub-run"); got != "PIPE_MY_STEP_SUB_RUN" {
		t.Fatalf("expected PIPE_MY_STEP_SUB_RUN, got %s", got)
	}
}

func TestBuildEnv_IncludesPipeVars(t *testing.T) {
	t.Parallel()
	vars := map[string]string{"PIPE_FOO": "bar"}
	env := BuildEnv(vars)

	// Should contain at least one OS env var (PATH is always present).
	hasPath := false
	hasPipeFoo := false
	for _, e := range env {
		if strings.HasPrefix(e, "PATH=") {
			hasPath = true
		}
		if e == "PIPE_FOO=bar" {
			hasPipeFoo = true
		}
	}
	if !hasPath {
		t.Fatal("expected PATH in env")
	}
	if !hasPipeFoo {
		t.Fatal("expected PIPE_FOO=bar in env")
	}

	// Total should be os.Environ() + 1
	if len(env) != len(os.Environ())+1 {
		t.Fatalf("expected %d entries, got %d", len(os.Environ())+1, len(env))
	}
}

func TestBuildEnv_EmptyMap(t *testing.T) {
	t.Parallel()
	env := BuildEnv(map[string]string{})
	if len(env) != len(os.Environ()) {
		t.Fatalf("expected %d entries, got %d", len(os.Environ()), len(env))
	}
}

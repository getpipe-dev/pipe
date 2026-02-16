package runner

import (
	"os"
	"strings"
)

// EnvKey builds a PIPE_* environment variable name from step/sub-run IDs.
// Hyphens become underscores, everything uppercased.
func EnvKey(parts ...string) string {
	joined := strings.Join(parts, "_")
	joined = strings.ReplaceAll(joined, "-", "_")
	return "PIPE_" + strings.ToUpper(joined)
}

// BuildEnv returns os.Environ() plus all accumulated PIPE_* vars.
func BuildEnv(pipeVars map[string]string) []string {
	env := os.Environ()
	for k, v := range pipeVars {
		env = append(env, k+"="+v)
	}
	return env
}

package config

import (
	"fmt"
	"os"
	"path/filepath"
)

var (
	BaseDir  string
	FilesDir string
	StateDir string
	LogDir   string
	CacheDir string
)

func init() {
	home, err := os.UserHomeDir()
	if err != nil {
		panic("cannot determine home directory: " + err.Error())
	}
	BaseDir = filepath.Join(home, ".pipe")
	FilesDir = filepath.Join(BaseDir, "files")
	StateDir = filepath.Join(BaseDir, "state")
	LogDir = filepath.Join(BaseDir, "logs")
	CacheDir = filepath.Join(BaseDir, "cache")
}

func EnsureDirs(pipelineName string) error {
	dirs := []string{
		FilesDir,
		filepath.Join(StateDir, pipelineName),
		LogDir,
		CacheDir,
	}
	for _, d := range dirs {
		if err := os.MkdirAll(d, 0o755); err != nil {
			return fmt.Errorf("creating directory %s: %w", d, err)
		}
	}
	return nil
}

package logging

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/destis/pipe/internal/config"
)

type Logger struct {
	info *log.Logger
	err  *log.Logger
	file *os.File
}

func New(pipelineName, runID string) (*Logger, error) {
	ts := time.Now().Format("20060102-150405")
	filename := fmt.Sprintf("%s-%s-%s.log", pipelineName, runID[:8], ts)
	path := filepath.Join(config.LogDir, filename)

	f, err := os.Create(path)
	if err != nil {
		return nil, fmt.Errorf("creating log file: %w", err)
	}

	infoW := io.MultiWriter(os.Stderr, f)
	errW := io.MultiWriter(os.Stderr, f)

	return &Logger{
		info: log.New(infoW, "", log.Ltime),
		err:  log.New(errW, "ERROR ", log.Ltime),
		file: f,
	}, nil
}

func (l *Logger) Info(format string, args ...any) {
	l.info.Printf(format, args...)
}

func (l *Logger) Error(format string, args ...any) {
	l.err.Printf(format, args...)
}

func (l *Logger) Close() error {
	return l.file.Close()
}

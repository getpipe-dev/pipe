package runner

import (
	"strings"
	"sync"
)

// outputWriter is an io.Writer that buffers input, splits on newlines,
// and calls an emit callback for each complete line.
type outputWriter struct {
	mu   sync.Mutex
	buf  strings.Builder
	emit func(string)
}

func newOutputWriter(emit func(string)) *outputWriter {
	return &outputWriter{
		emit: emit,
	}
}

func (w *outputWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.buf.Write(p)
	for {
		s := w.buf.String()
		idx := strings.IndexByte(s, '\n')
		if idx < 0 {
			break
		}
		line := s[:idx]
		w.emit(line)
		// Keep the remainder after the newline
		remainder := s[idx+1:]
		w.buf.Reset()
		w.buf.WriteString(remainder)
	}
	return len(p), nil
}

// Flush emits any remaining partial line in the buffer.
func (w *outputWriter) Flush() {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.buf.Len() > 0 {
		w.emit(w.buf.String())
		w.buf.Reset()
	}
}

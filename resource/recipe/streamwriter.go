package recipe

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"sync"
)

// slog attribute key identifying which standard stream a build log line
// originated from.
const logStreamKey = "stream"

// Returns an [io.Writer] that routes each newline-delimited line from written
// bytes to the default slog logger at level.
//
// stream is attached to every log record as the "stream" attribute and
// identifies the source (e.g. "stdout", "stderr"). Safe for concurrent use.
func newStreamWriter(ctx context.Context, level slog.Level, stream string) io.Writer {
	return &streamWriter{ctx: ctx, level: level, stream: stream}
}

// Concrete implementation backed by a line-buffered slog emitter.
type streamWriter struct {
	mu     sync.Mutex
	buf    []byte
	ctx    context.Context
	level  slog.Level
	stream string
}

// Appends p to the internal buffer and emits one log record per complete line.
func (w *streamWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.buf = append(w.buf, p...)
	for {
		i := bytes.IndexByte(w.buf, '\n')
		if i < 0 {
			break
		}
		if i > 0 {
			w.emit(string(w.buf[:i]))
		}
		w.buf = w.buf[i+1:]
	}
	return len(p), nil
}

// Emits a single log record. Caller must hold w.mu.
func (w *streamWriter) emit(line string) {
	slog.Log(w.ctx, w.level, line, logStreamKey, w.stream)
}

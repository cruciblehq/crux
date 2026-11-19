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

// Returns an [io.Writer] that logs each line of the written bytes to the
// default slog logger at level.
//
// stream is attached to every log record as the "stream" attribute and
// identifies the source (e.g. "stdout", "stderr"). Safe for concurrent use.
func newStreamWriter(level slog.Level, stream string) io.Writer {
	return &streamWriter{level: level, stream: stream}
}

// Writer that logs each complete line it receives as a slog record.
type streamWriter struct {
	mu     sync.Mutex // Guards buf.
	buf    []byte     // Bytes received after the last complete line.
	level  slog.Level // Level of each log record.
	stream string     // Value of the stream attribute.
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
//
// [io.Writer] carries no context, so the record gets an empty one.
func (w *streamWriter) emit(line string) {
	slog.Log(context.Background(), w.level, line, logStreamKey, w.stream)
}

package crex

import (
	"context"
	"io"
	"log/slog"
	"os"
	"slices"
	"sync"
)

// Custom implementation of the [slog.Handler] interface that buffers logs and
// can flush them to an output stream using a specified [Formatter].
//
// The handler supports setting log attributes and groups, and allows dynamic
// configuration of the output stream and formatter. The handler buffers log
// records until a formatter is set. At that point, it starts flushing buffered
// records to the output stream. The flush can happen implicitly on the next log
// record or explicitly by calling [Flush].
//
// The handler is safe for concurrent use.
type Handler interface {
	slog.Handler

	// Sets the minimum level for the handler.
	//
	// Only log records with a level equal to or higher than this level will be
	// processed. The default level is [slog.LevelInfo]. The method returns the
	// handler itself to allow method chaining. The current log level can be
	// retrieved using [Level]. Setting the log level only affects new records;
	// it does not retroactively filter already buffered records.
	SetLevel(slog.Level) Handler

	// Returns the current minimum level of the handler.
	//
	// Only log records with a level equal to or higher than this level are
	// processed. The default level is [slog.LevelInfo]. The log level can
	// be changed using [SetLevel].
	Level() slog.Level

	// Sets the output stream for the handler.
	SetStream(io.Writer) Handler

	// Returns the current output stream of the handler.
	Stream() io.Writer

	// Sets the formatter for the handler.
	//
	// After setting the formatter, buffered log records can be flushed to the
	// output stream by calling [Flush], or implicitly on the next log record.
	SetFormatter(Formatter) Handler

	// Returns the current Formatter of the handler.
	Formatter() Formatter

	// Writes all buffered records to the output stream using the set formatter.
	//
	// The bool return value indicates whether a flush was attempted, which
	// happens only if a formatter is set. The error return value indicates
	// whether an error occurred during formatting or writing. After a successful
	// flush, the buffer is cleared. If an error occurs after some records have
	// been written, those records are removed from the buffer, and the rest
	// remain for a future flush attempt.
	Flush() (bool, error)
}

// Holds the mutable state shared between a parent handler and its children
// (created via WithAttrs/WithGroup).
type sharedState struct {
	mux       sync.RWMutex    // Guards all other fields.
	level     slog.Level      // Minimum level for records to be processed.
	buffer    []RecordContext // Buffered records pending a formatter.
	formatter Formatter       // Formatter used to write records; nil means buffer-only.
	stream    io.Writer       // Destination stream; defaults to os.Stderr.
}

// Concrete implementation of [Handler].
type handler struct {
	state  *sharedState // Shared state between parent and child handlers.
	attrs  []slog.Attr  // Attributes appended to every record this handler processes.
	groups []string     // Active group path, accumulated via WithGroup.
}

// Creates a new [Handler] with the default level ([slog.LevelInfo]).
func NewHandler() Handler {
	return NewHandlerWithLevel(slog.LevelInfo)
}

// Creates a new [Handler] with the given minimum log level.
func NewHandlerWithLevel(level slog.Level) Handler {
	return &handler{
		state: &sharedState{
			level:     level,
			buffer:    make([]RecordContext, 0),
			formatter: nil,       // No formatter means we buffer only
			stream:    os.Stderr, // Default to stderr
		},
		attrs:  make([]slog.Attr, 0),
		groups: make([]string, 0),
	}
}

// Acquires the shared write lock and updates the minimum log level.
func (h *handler) SetLevel(level slog.Level) Handler {
	h.state.mux.Lock()
	defer h.state.mux.Unlock()

	h.state.level = level

	return h
}

// Acquires the shared read lock and returns the current minimum log level.
func (h *handler) Level() slog.Level {
	h.state.mux.RLock()
	defer h.state.mux.RUnlock()

	return h.state.level
}

// Acquires the shared write lock and updates the output stream.
func (h *handler) SetStream(stream io.Writer) Handler {
	h.state.mux.Lock()
	defer h.state.mux.Unlock()

	h.state.stream = stream

	return h
}

// Acquires the shared read lock and returns the current output stream. Defaults to [os.Stderr].
func (h *handler) Stream() io.Writer {
	h.state.mux.RLock()
	defer h.state.mux.RUnlock()

	return h.state.stream
}

// Acquires the shared write lock and updates the formatter.
func (h *handler) SetFormatter(formatter Formatter) Handler {
	h.state.mux.Lock()
	defer h.state.mux.Unlock()

	h.state.formatter = formatter

	return h
}

// Acquires the shared read lock and returns the current formatter.
func (h *handler) Formatter() Formatter {
	h.state.mux.RLock()
	defer h.state.mux.RUnlock()

	return h.state.formatter
}

// Acquires the shared write lock and delegates to the internal [handler.flush].
func (h *handler) Flush() (bool, error) {
	h.state.mux.Lock()
	defer h.state.mux.Unlock()

	return h.flush()
}

// Writes all buffered records to the output stream.
//
// Returns (false, nil) if no formatter is set. Returns (true, nil) on success.
// Returns (true, err) if an error occurred; written records are removed from
// buffer, unwritten records remain.
//
// Caller must hold the state mutex.
func (h *handler) flush() (bool, error) {
	if h.state.formatter == nil {
		return false, nil
	}

	var err error
	var idx int

	// Write in order
	for idx = 0; idx < len(h.state.buffer); idx++ {
		err = h.state.formatter.Write(h.state.stream, &h.state.buffer[idx])
		if err != nil {
			break
		}
	}

	h.state.buffer = h.state.buffer[idx:]
	h.state.buffer = slices.Clip(h.state.buffer)

	return true, err
}

// Whether the given level meets the handler's current minimum level.
func (h *handler) Enabled(_ context.Context, level slog.Level) bool {
	h.state.mux.RLock()
	defer h.state.mux.RUnlock()

	return level >= h.state.level
}

// Appends a clone of the record (with local attrs merged) to the shared buffer
// and triggers an implicit flush.
func (h *handler) Handle(_ context.Context, record slog.Record) error {
	newRecord := record.Clone()
	newRecord.AddAttrs(h.attrs...)

	h.state.mux.Lock()
	defer h.state.mux.Unlock()

	h.state.buffer = append(h.state.buffer, RecordContext{
		Record: newRecord,
		Groups: slices.Clone(h.groups),
	})
	_, err := h.flush()
	return err
}

// Returns a new handler sharing state with this one but with the given attrs appended.
func (h *handler) WithAttrs(attrs []slog.Attr) slog.Handler {
	// No lock needed here as we are creating a new struct with copied data
	// and h.attrs and h.groups are immutable after creation.

	if len(attrs) == 0 {
		return h
	}

	return &handler{
		state:  h.state, // Share the state
		attrs:  append(slices.Clip(h.attrs), attrs...),
		groups: slices.Clip(h.groups),
	}
}

// Returns a new handler sharing state with this one but with name appended to groups.
func (h *handler) WithGroup(name string) slog.Handler {
	if name == "" {
		return h
	}

	return &handler{
		state:  h.state, // Shared state
		attrs:  slices.Clip(h.attrs),
		groups: append(slices.Clip(h.groups), name),
	}
}

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

// Provides context for formatting a log record.
//
// This structure is used to preserve log groups along with the record itself
// during formatting.
type RecordContext struct {
	Record slog.Record
	Groups []string
}

// Holds the mutable state shared between a parent handler and its children
// (created via WithAttrs/WithGroup).
type sharedState struct {
	mux       sync.RWMutex
	level     slog.Level
	buffer    []RecordContext
	formatter Formatter
	stream    io.Writer
}

// Concrete implementation of the Handler interface.
//
// It buffers log records and can flush them to an output stream using a specified
// Formatter. The handler supports setting log attributes and groups, and allows
// dynamic configuration of the output stream and formatter.
//
// The handler is safe for concurrent use.
type handler struct {
	state  *sharedState // Pointer to shared state
	attrs  []slog.Attr  // Local attributes
	groups []string     // Local groups
}

// Creates a new [Handler] with the default log level (slog.LevelInfo).
//
// The handler starts with an empty buffer and no formatter. The default log
// level is [slog.LevelInfo]. The default output stream is [os.Stderr], but it
// can be changed using [SetStream]. The handler buffers log records until a
// formatter is set using [SetFormatter], at which point it will start flushing
// buffered records to the output stream. The flush can also be triggered
// manually by calling [Flush].
//
// The handler is safe for concurrent use.
func NewHandler() Handler {
	return NewHandlerWithLevel(slog.LevelInfo)
}

// Creates a new [Handler] with the specified log level.
//
// The handler starts with an empty buffer and no formatter. The default log
// level is [slog.LevelInfo]. The default output stream is [os.Stderr], but it
// can be changed using [SetStream]. The handler buffers log records until a
// formatter is set using [SetFormatter], at which point it will start flushing
// buffered records to the output stream. The flush can also be triggered
// manually by calling [Flush].
//
// The handler is safe for concurrent use.
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

// Sets the minimum level for the handler.
//
// Only log records with a level equal to or higher than this level will be
// processed. The default level is [slog.LevelInfo]. The method returns the
// handler itself to allow method chaining. The current log level can be
// retrieved using [Level]. Setting the log level only affects new records;
// it does not retroactively filter already buffered records.
func (h *handler) SetLevel(level slog.Level) Handler {
	h.state.mux.Lock()
	defer h.state.mux.Unlock()

	h.state.level = level

	return h
}

// Returns the current minimum level of the handler.
//
// Only log records with a level equal to or higher than this level are
// processed. The default level is [slog.LevelInfo]. The log level can be
// changed using [SetLevel].
func (h *handler) Level() slog.Level {
	h.state.mux.RLock()
	defer h.state.mux.RUnlock()

	return h.state.level
}

// Sets the output stream for the handler.
func (h *handler) SetStream(stream io.Writer) Handler {
	h.state.mux.Lock()
	defer h.state.mux.Unlock()

	h.state.stream = stream

	return h
}

// Returns the current output stream of the handler.
//
// The default stream is [os.Stderr].
func (h *handler) Stream() io.Writer {
	h.state.mux.RLock()
	defer h.state.mux.RUnlock()

	return h.state.stream
}

// Sets the formatter for the handler.
//
// After setting the formatter, buffered log records can be flushed to the
// output stream by calling [Flush], or implicitly on the next log record.
func (h *handler) SetFormatter(formatter Formatter) Handler {
	h.state.mux.Lock()
	defer h.state.mux.Unlock()

	h.state.formatter = formatter

	return h
}

// Returns the current Formatter of the handler.
func (h *handler) Formatter() Formatter {
	h.state.mux.RLock()
	defer h.state.mux.RUnlock()

	return h.state.formatter
}

// Writes all buffered records to the output stream using the set formatter.
//
// The bool return value indicates whether a flush was attempted, which happens
// only if a formatter is set. The error return value indicates whether an error
// occurred during formatting or writing. After a successful flush, the buffer
// is cleared. If an error occurs after some records have been written, those
// records are removed from the buffer.
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

// Determines whether a log record with the given level should be processed.
//
// Only records with a level equal to or higher than the handler's current
// level are processed. If the record is not enabled, it is ignored, not even
// being buffered.
//
// Implements [slog.Handler.Enabled].
func (h *handler) Enabled(_ context.Context, level slog.Level) bool {
	h.state.mux.RLock()
	defer h.state.mux.RUnlock()

	return level >= h.state.level
}

// Handles a log record by buffering it and attempting to flush the buffer.
//
// The record is considered only if its level is enabled according to [Enabled].
// The method returns any error encountered during an implicit flush that may
// occur after buffering the record (if a formatter is set).
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

// Creates a new handler with the given attributes added.
//
// The new handler shares the same underlying state (buffer, level, stream,
// formatter) as the original handler, but has its own set of attributes and
// groups. The original handler remains unchanged.
//
// Implements [slog.Handler.WithAttrs].
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

// Creates a new handler with the given group added.
//
// The new handler shares the same underlying state (buffer, level, stream,
// formatter) as the original handler, but has its own set of attributes and
// groups. The original handler remains unchanged.
//
// Implements [slog.Handler.WithGroup].
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

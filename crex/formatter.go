package crex

import (
	"io"
	"log/slog"
)

// Provides context for formatting a log record.
//
// Preserves log groups along with the record itself during formatting.
type RecordContext struct {
	Record slog.Record // The log record to format.
	Groups []string    // Nested group names accumulated via WithGroup.
}

// Formats [slog.Record] entries.
type Formatter interface {

	// Formats a log record and writes it to the provided writer.
	Write(w io.Writer, rctx *RecordContext) error
}

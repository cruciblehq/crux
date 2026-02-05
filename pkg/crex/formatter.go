package crex

import (
	"io"
)

// Formatter formats [slog.Record] entries.
type Formatter interface {

	// Formats a log record and writes it to the provided writer.
	Write(w io.Writer, rctx *RecordContext) error
}

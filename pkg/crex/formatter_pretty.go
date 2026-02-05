package crex

import (
	"io"
	"log/slog"
	"strings"
)

const colorReset = "\033[0m"

// Formats log records for human-readable TTY output.
//
// The formatter outputs log records with optional ANSI color codes for log
// levels. Crex errors are detected automatically and formatted as "message:
// reason. fallback" instead of key=value pairs.
//
// When verbose mode is disabled (default), only the message and crex error
// info are shown. When enabled, all attributes are included.
//
// Example output (non-verbose):
//
//	crux [info]: server started
//	crux [error]: build failed: invalid type. Use widget or service.
//
// Example output (verbose):
//
//	crux [info]: server started port=8080 host=localhost
//	crux [error]: build failed: invalid type. Use widget or service. class=user
type PrettyFormatter struct {
	UseColor bool
	verbose  bool
}

// Creates a new pretty formatter.
//
// If useColor is true, log levels are colored using ANSI escape codes.
// Verbose mode is disabled by default.
func NewPrettyFormatter(useColor bool) *PrettyFormatter {
	return &PrettyFormatter{UseColor: useColor}
}

// SetVerbose enables or disables verbose output.
func (f *PrettyFormatter) SetVerbose(verbose bool) *PrettyFormatter {
	f.verbose = verbose
	return f
}

// Verbose returns whether verbose output is enabled.
func (f *PrettyFormatter) Verbose() bool {
	return f.verbose
}

// Writes a log record as a human-readable line.
func (f *PrettyFormatter) Write(w io.Writer, rctx *RecordContext) error {
	var sb strings.Builder

	// Groups prefix
	if len(rctx.Groups) > 0 {
		sb.WriteString(strings.Join(rctx.Groups, "."))
		sb.WriteString(" ")
	}

	// Level
	sb.WriteString("[")
	if f.UseColor {
		sb.WriteString(levelColor(rctx.Record.Level))
	}
	sb.WriteString(strings.ToLower(rctx.Record.Level.String()))
	if f.UseColor {
		sb.WriteString(colorReset)
	}
	sb.WriteString("]: ")

	// Write message with interpolated attributes
	f.writeMessage(&sb, rctx)

	sb.WriteString("\n")
	_, err := w.Write([]byte(sb.String()))
	return err
}

// Writes the message, handling crex errors specially.
func (f *PrettyFormatter) writeMessage(sb *strings.Builder, rctx *RecordContext) {
	sb.WriteString(rctx.Record.Message)

	// Single pass through attributes
	rctx.Record.Attrs(func(attr slog.Attr) bool {
		// Resolve LogValuer if present
		resolvedValue := attr.Value.Resolve()

		if errMap, ok := asCrexError(resolvedValue); ok {
			f.writeCrexError(sb, errMap)
			return true
		}
		if f.verbose {
			f.writeInlineAttr(sb, attr)
		}
		return true
	})
}

// Writes an attribute inline in the message.
func (f *PrettyFormatter) writeInlineAttr(sb *strings.Builder, attr slog.Attr) {
	sb.WriteString(" ")
	sb.WriteString(attr.Key)
	sb.WriteString("=")
	sb.WriteString(formatValue(attr.Value))
}

// Formats a value for inline display.
func formatValue(v slog.Value) string {
	switch v.Kind() {
	case slog.KindDuration:
		return v.Duration().String()
	case slog.KindTime:
		return v.Time().Format("15:04:05")
	case slog.KindGroup:
		var parts []string
		for _, attr := range v.Group() {
			parts = append(parts, attr.Key+"="+formatValue(attr.Value))
		}
		return "{" + strings.Join(parts, " ") + "}"
	default:
		return v.String()
	}
}

// Writes a crex error in "message: reason. fallback" format.
func (f *PrettyFormatter) writeCrexError(sb *strings.Builder, errMap map[string]slog.Value) {
	if reason, ok := errMap["reason"]; ok && reason.String() != "" {
		sb.WriteString(": ")
		sb.WriteString(reason.String())
	}

	if fallback, ok := errMap["fallback"]; ok && fallback.String() != "" {
		sb.WriteString(". ")
		sb.WriteString(fallback.String())
	}

	// In verbose mode, include additional error details
	if f.verbose {
		for key, val := range errMap {
			if key != "reason" && key != "fallback" && key != "description" {
				sb.WriteString(" ")
				sb.WriteString(key)
				sb.WriteString("=")
				sb.WriteString(formatValue(val))
			}
		}
	}
}

// Checks whether a value is a crex error group by looking for the sentinel marker.
func asCrexError(val slog.Value) (map[string]slog.Value, bool) {
	if val.Kind() != slog.KindGroup {
		return nil, false
	}

	errMap := make(map[string]slog.Value)
	for _, attr := range val.Group() {
		errMap[attr.Key] = attr.Value
	}

	// Check for crex error sentinel marker
	if marker, ok := errMap[crexErrorMarker]; !ok || !marker.Bool() {
		return nil, false
	}

	return errMap, true
}

// Returns the ANSI color code for a log level.
func levelColor(level slog.Level) string {
	switch level {
	case slog.LevelDebug:
		return "\033[37m" // White
	case slog.LevelInfo:
		return "\033[32m" // Green
	case slog.LevelWarn:
		return "\033[33m" // Yellow
	case slog.LevelError:
		return "\033[31m" // Red
	default:
		return "\033[0m" // Reset
	}
}
